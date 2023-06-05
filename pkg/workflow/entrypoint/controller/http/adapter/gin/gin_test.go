package gin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kafka-polygon/pkg/env"
	pkgGin "kafka-polygon/pkg/http/gin"
	"kafka-polygon/pkg/workflow/entity"
	controllerHTTP "kafka-polygon/pkg/workflow/entrypoint/controller/http"
	adapterGin "kafka-polygon/pkg/workflow/entrypoint/controller/http/adapter/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tj/assert"
)

var (
	_bgCtx = context.Background()
)

type mockUseCase struct {
	searchWorkflowsFunc func(
		ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error)
	restartWorkflowFunc func(
		ctx context.Context, workflowID entity.ID, payload json.RawMessage) error
	restartWorkflowFromFunc func(
		ctx context.Context, workflowID entity.ID, from entity.WorkflowSchemaStepName, payload json.RawMessage) error
}

func (m *mockUseCase) SearchWorkflows(
	ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error) {
	return m.searchWorkflowsFunc(ctx, params)
}

func (m *mockUseCase) RestartWorkflow(ctx context.Context, workflowID entity.ID, payload json.RawMessage) error {
	return m.restartWorkflowFunc(ctx, workflowID, payload)
}

func (m *mockUseCase) RestartWorkflowFrom(
	ctx context.Context, workflowID entity.ID, from entity.WorkflowSchemaStepName, payload json.RawMessage) error {
	return m.restartWorkflowFromFunc(ctx, workflowID, from, payload)
}

func TestSearchWorkflows(t *testing.T) {
	expSearchParams := entity.SearchWorkflowParams{
		ID:     entity.PointerID("123"),
		Status: entity.PointerWorkflowStatus("ok"),
		Paging: &entity.Paging{
			Limit:  1,
			Offset: 1,
		},
	}

	expSearchWorkflowsFuncResult := &entity.SearchWorkflowResult{
		Workflows: []*entity.Workflow{{
			ID:    *expSearchParams.ID,
			Input: []byte("{}"),
		}},
		Paging: entity.Paging{
			Limit:  expSearchParams.Limit,
			Offset: expSearchParams.Offset,
		},
	}

	uc := &mockUseCase{
		searchWorkflowsFunc: func(
			ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error) {
			assert.Equal(t, expSearchParams, params)

			return expSearchWorkflowsFuncResult, nil
		},
	}

	tm := &testModel{
		method: http.MethodGet,
		route: fmt.Sprintf("/workflows/?id=%s&status=%s&limit=%d&offset=%d",
			expSearchParams.ID, expSearchParams.Status, expSearchParams.Limit, expSearchParams.Offset),
		req:          nil,
		dst:          new(controllerHTTP.SearchResponse),
		expectedCode: http.StatusOK,
		assertFn: func(code int, resp interface{}) {
			body, ok := resp.(*controllerHTTP.SearchResponse)
			assert.True(t, ok)
			assert.Equal(t, expSearchWorkflowsFuncResult.Workflows, body.Data)
			assert.Equal(t, expSearchWorkflowsFuncResult.Paging, body.Paging)
		}}

	testByModel(t, newServer(uc), tm)
}

func TestRestartWorkflow(t *testing.T) {
	expWorkflowID := entity.ID("123")
	req := controllerHTTP.RestartWorkflowRequest{
		Payload: []byte("{}"),
	}
	isCalled := false

	uc := &mockUseCase{
		restartWorkflowFunc: func(ctx context.Context, workflowID entity.ID, payload json.RawMessage) error {
			isCalled = true
			assert.Equal(t, expWorkflowID, workflowID)
			assert.Equal(t, req.Payload, payload)

			return nil
		},
	}

	tm := &testModel{
		method:       http.MethodPost,
		route:        fmt.Sprintf("/workflows/restart/%s", expWorkflowID),
		req:          req,
		dst:          nil,
		expectedCode: http.StatusOK,
	}

	testByModel(t, newServer(uc), tm)

	assert.True(t, isCalled)
}

func TestRestartWorkflowFrom(t *testing.T) {
	expWorkflowID := entity.ID("123")
	req := controllerHTTP.RestartWorkflowFromRequest{
		StepName: "step1",
		Payload:  []byte("{}"),
	}
	isCalled := false

	uc := &mockUseCase{
		restartWorkflowFromFunc: func(
			ctx context.Context, workflowID entity.ID, from entity.WorkflowSchemaStepName, payload json.RawMessage) error {
			isCalled = true

			assert.Equal(t, expWorkflowID, workflowID)
			assert.Equal(t, req.StepName, from)
			assert.Equal(t, req.Payload, payload)

			return nil
		},
	}

	tm := &testModel{
		method:       http.MethodPost,
		route:        fmt.Sprintf("/workflows/restart/from/%s", expWorkflowID),
		req:          req,
		dst:          nil,
		expectedCode: http.StatusOK,
	}

	testByModel(t, newServer(uc), tm)

	assert.True(t, isCalled)
}

type testModel struct {
	method       string
	route        string
	req          interface{}
	dst          interface{}
	expectedCode int
	assertFn     func(code int, respBody interface{})
}

func testByModel(t *testing.T, s *pkgGin.Server, m *testModel) {
	t.Helper()

	code, err := makeReq(s, m.method, m.route, m.req, m.dst)
	assert.NoError(t, err)
	assert.Equal(t, m.expectedCode, code)

	if m.assertFn != nil {
		m.assertFn(code, m.dst)
	}
}

func makeReq(s *pkgGin.Server, method, route string, body, dst interface{}) (status int, err error) {
	w := httptest.NewRecorder()
	buf := new(bytes.Buffer)

	if body != nil {
		b, _ := json.Marshal(body)
		_, _ = buf.Write(b)
	}

	req := httptest.NewRequest(method, route, buf)
	req.Header.Set("Content-Type", "application/json")

	s.Gin().ServeHTTP(w, req)

	if w != nil {
		defer w.Flush()
	}

	if dst != nil {
		b, _ := io.ReadAll(w.Body)
		_ = json.Unmarshal(b, dst)
	}

	return w.Code, nil
}

func newServer(uc *mockUseCase) *pkgGin.Server {
	s := pkgGin.NewServer(&pkgGin.ServerConfig{Server: env.HTTPServer{Port: "3000"}}).WithDefaultKit()
	s.Gin().UnescapePathValues = true
	_ = controllerHTTP.RegisterWorkflowRoutes(_bgCtx, adapterGin.NewAdapter(s, uc))

	return s
}
