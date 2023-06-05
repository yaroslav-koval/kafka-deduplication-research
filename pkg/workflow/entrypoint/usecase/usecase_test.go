package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/workflow/entity"
	"kafka-polygon/pkg/workflow/entrypoint/usecase"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/tj/assert"
)

const (
	_reqID = "1"
)

var (
	_bgCtx          = context.Background()
	_bgCtxWithReqID = context.WithValue(_bgCtx, consts.HeaderXRequestID, _reqID) //nolint:staticcheck
)

func TestSearchWorkflows(t *testing.T) {
	searchParams := entity.SearchWorkflowParams{Status: entity.PointerWorkflowStatus("failed")}
	expRes := &entity.SearchWorkflowResult{Workflows: make([]*entity.Workflow, 1)}
	expErr := fmt.Errorf("err")

	uc := usecase.New(nil, &mockStore{
		searchWorkflowsFunc: func(ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error) {
			assert.Equal(t, searchParams, params)

			return expRes, expErr
		},
	})
	actRes, actErr := uc.SearchWorkflows(_bgCtx, searchParams)

	assert.Equal(t, expRes, actRes)
	assert.Equal(t, expErr, actErr)
}

func TestRestartWorkflow(t *testing.T) {
	expGetWorkflowByIDResult := &entity.Workflow{ID: entity.ID("123")}

	store := &mockStore{
		getWorkflowByIDFunc: func(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
			assert.Equal(t, expGetWorkflowByIDResult.ID, id)

			return expGetWorkflowByIDResult, nil
		},
	}

	uc := usecase.New(nil, store)

	// workflow must be in appropriate status
	actErr := uc.RestartWorkflow(_bgCtx, expGetWorkflowByIDResult.ID, nil)
	assert.Equal(t, "workflow is not in FAILED status", actErr.Error())
	// set valid status
	expGetWorkflowByIDResult.Status = entity.WorkflowStatusFailed

	// workflow must have input or at least was step with input data
	actErr = uc.RestartWorkflow(_bgCtx, expGetWorkflowByIDResult.ID, nil)
	assert.Equal(t, "both workflow input and last step data are empty", actErr.Error())
	// set valid input
	expGetWorkflowByIDResult.Input = []byte("input data")

	// workflow is restarted with expected params
	expGetWorkflowByIDResult.SchemaName = entity.WorkflowSchemaName("schema")
	expGetWorkflowByIDResult.Steps = []*entity.WorkflowStep{
		{Name: "step1", Data: []byte("step1 data")},
		{Name: "step2", Data: []byte("step2 data")},
	}
	expwWorkflowSchemaResult, err := entity.NewWorkflowSchema(
		_bgCtx,
		expGetWorkflowByIDResult.SchemaName,
		entity.NewWorkflowSchemaSimpleStep(expGetWorkflowByIDResult.Steps[0].Name, "topic1", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep(expGetWorkflowByIDResult.Steps[1].Name, "topic2", new(stepWorkerTest)))
	assert.NoError(t, err)

	orchestractor := &mockOrchestrator{
		workflowSchemaFunc: func(ctx context.Context, wsn entity.WorkflowSchemaName) (*entity.WorkflowSchema, error) {
			assert.Equal(t, expGetWorkflowByIDResult.SchemaName, wsn)

			return expwWorkflowSchemaResult, nil
		},
		restartFromFunc: func(
			ctx context.Context,
			workflowID entity.ID,
			schemaName entity.WorkflowSchemaName,
			stepName entity.WorkflowSchemaStepName,
			payload json.RawMessage) (entity.ID, error) {
			assert.Equal(t, expGetWorkflowByIDResult.ID, workflowID)
			assert.Equal(t, expGetWorkflowByIDResult.SchemaName, schemaName)
			assert.Equal(t, expGetWorkflowByIDResult.Steps[1].Name, stepName)
			assert.Equal(t, expGetWorkflowByIDResult.Steps[1].Data, payload)

			return expGetWorkflowByIDResult.ID, nil
		},
	}

	store.newIDFunc = func() entity.ID {
		return entity.ID(uuid.NewV4().String())
	}

	store.createWorkflowHistoryFunc = func(ctx context.Context, wh *entity.WorkflowHistory) error {
		assert.NotEmpty(t, wh.ID)
		assert.False(t, wh.CreatedAt.IsZero())

		assert.Equal(t, &entity.WorkflowHistory{
			ID:                wh.ID,
			CreatedAt:         wh.CreatedAt,
			Type:              entity.WorkflowHistoryTypeRestart,
			Input:             expGetWorkflowByIDResult.Steps[1].Data,
			InputPrevious:     expGetWorkflowByIDResult.Steps[1].Data,
			StepName:          expGetWorkflowByIDResult.Steps[1].Name,
			WorkflowID:        expGetWorkflowByIDResult.ID,
			WorkflowStatus:    expGetWorkflowByIDResult.Status,
			WorkflowError:     expGetWorkflowByIDResult.Error,
			WorkflowErrorKind: expGetWorkflowByIDResult.ErrorKind,
			RequestID:         converto.StringPointer(_reqID),
		}, wh)

		return nil
	}

	uc = usecase.New(orchestractor, store)
	_ = uc.RestartWorkflow(_bgCtxWithReqID, expGetWorkflowByIDResult.ID, nil)
}

func TestRestartWorkflowFrom(t *testing.T) {
	expGetWorkflowByIDResult := &entity.Workflow{ID: entity.ID("123")}

	// required arguments
	actErr := usecase.New(nil, nil).RestartWorkflowFrom(_bgCtx, expGetWorkflowByIDResult.ID, "", nil)
	assert.Equal(t, "step name and payload are required", actErr.Error())

	store := &mockStore{
		getWorkflowByIDFunc: func(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
			assert.Equal(t, expGetWorkflowByIDResult.ID, id)

			return expGetWorkflowByIDResult, nil
		},
	}

	uc := usecase.New(nil, store)

	// workflow must be in appropriate status
	expFrom := entity.WorkflowSchemaStepName("step2")
	expPayload := []byte("{}")

	actErr = uc.RestartWorkflowFrom(_bgCtx, expGetWorkflowByIDResult.ID, expFrom, expPayload)
	assert.Equal(t, "workflow is not in SUCCESS status", actErr.Error())
	// set valid status
	expGetWorkflowByIDResult.Status = entity.WorkflowStatusSuccess

	// workflow is restarted with expected params
	expGetWorkflowByIDResult.SchemaName = entity.WorkflowSchemaName("schema")
	expGetWorkflowByIDResult.Steps = []*entity.WorkflowStep{
		{Name: "step1", Data: []byte("step1 data")},
		{Name: "step2", Data: []byte("step2 data")},
	}

	orchestractor := &mockOrchestrator{
		startFromFunc: func(
			ctx context.Context,
			schemaName entity.WorkflowSchemaName,
			stepName entity.WorkflowSchemaStepName,
			parentID *entity.ID,
			payload json.RawMessage) (entity.ID, error) {
			assert.Equal(t, expGetWorkflowByIDResult.SchemaName, schemaName)
			assert.Equal(t, expFrom, stepName)
			assert.Equal(t, &expGetWorkflowByIDResult.ID, parentID)
			assert.Equal(t, expPayload, []byte(payload))

			return entity.ID(expGetWorkflowByIDResult.ID.String() + "1"), nil
		},
	}

	store.newIDFunc = func() entity.ID {
		return entity.ID(uuid.NewV4().String())
	}
	store.createWorkflowHistoryFunc = func(ctx context.Context, wh *entity.WorkflowHistory) error {
		assert.NotEmpty(t, wh.ID)
		assert.False(t, wh.CreatedAt.IsZero())

		assert.Equal(t, &entity.WorkflowHistory{
			ID:                wh.ID,
			CreatedAt:         wh.CreatedAt,
			Type:              entity.WorkflowHistoryTypeRestartFrom,
			Input:             expPayload,
			InputPrevious:     expGetWorkflowByIDResult.Steps[1].Data,
			StepName:          expFrom,
			WorkflowID:        expGetWorkflowByIDResult.ID,
			WorkflowStatus:    expGetWorkflowByIDResult.Status,
			WorkflowError:     expGetWorkflowByIDResult.Error,
			WorkflowErrorKind: expGetWorkflowByIDResult.ErrorKind,
			RequestID:         converto.StringPointer(_reqID),
		}, wh)

		return nil
	}

	uc = usecase.New(orchestractor, store)
	_ = uc.RestartWorkflowFrom(_bgCtxWithReqID, expGetWorkflowByIDResult.ID, expFrom, expPayload)
}

type mockOrchestrator struct {
	workflowSchemaFunc func(ctx context.Context, wsn entity.WorkflowSchemaName) (*entity.WorkflowSchema, error)
	startFromFunc      func(
		ctx context.Context,
		schemaName entity.WorkflowSchemaName,
		stepName entity.WorkflowSchemaStepName,
		parentID *entity.ID,
		payload json.RawMessage) (entity.ID, error)
	restartFromFunc func(
		ctx context.Context,
		workflowID entity.ID,
		schemaName entity.WorkflowSchemaName,
		stepName entity.WorkflowSchemaStepName,
		payload json.RawMessage) (entity.ID, error)
}

func (m *mockOrchestrator) WorkflowSchema(ctx context.Context, wsn entity.WorkflowSchemaName) (*entity.WorkflowSchema, error) {
	return m.workflowSchemaFunc(ctx, wsn)
}

func (m *mockOrchestrator) StartFrom(
	ctx context.Context,
	schemaName entity.WorkflowSchemaName,
	stepName entity.WorkflowSchemaStepName,
	parentID *entity.ID,
	payload json.RawMessage) (entity.ID, error) {
	return m.startFromFunc(ctx, schemaName, stepName, parentID, payload)
}

func (m *mockOrchestrator) RestartFrom(
	ctx context.Context,
	workflowID entity.ID,
	schemaName entity.WorkflowSchemaName,
	stepName entity.WorkflowSchemaStepName,
	payload json.RawMessage) (entity.ID, error) {
	return m.restartFromFunc(ctx, workflowID, schemaName, stepName, payload)
}

type mockStore struct {
	newIDFunc                 func() entity.ID
	getWorkflowByIDFunc       func(ctx context.Context, id entity.ID) (*entity.Workflow, error)
	searchWorkflowsFunc       func(ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error)
	createWorkflowHistoryFunc func(ctx context.Context, wh *entity.WorkflowHistory) error
}

func (m *mockStore) NewID() entity.ID {
	return m.newIDFunc()
}

func (m *mockStore) GetWorkflowByID(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
	return m.getWorkflowByIDFunc(ctx, id)
}

func (m *mockStore) SearchWorkflows(ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error) {
	return m.searchWorkflowsFunc(ctx, params)
}

func (m *mockStore) CreateWorkflowHistory(ctx context.Context, wh *entity.WorkflowHistory) error {
	return m.createWorkflowHistoryFunc(ctx, wh)
}

type stepWorkerTest struct{}

func (s *stepWorkerTest) Run(ctx context.Context, e event.WorkflowEvent) (json.RawMessage, error) {
	return []byte("{}"), nil
}
