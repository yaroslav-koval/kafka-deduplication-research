package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/fhir/client"
	"kafka-polygon/pkg/log"
	"strings"
	"testing"
	"time"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type clientTestSuite struct {
	suite.Suite
	fhirClient    *client.Client
	fhirClientCfg *env.FHIR
	httpClient    *mockClient
}

type mockClient struct {
	DoFunc func(req *fasthttp.Request, resp *fasthttp.Response) error
}

func (m *mockClient) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return m.DoFunc(req, resp)
}

func (m *mockClient) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return m.DoFunc(req, resp)
}

func TestClientTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(clientTestSuite))
}

func (s *clientTestSuite) SetupSuite() {
	s.httpClient = new(mockClient)
	s.fhirClientCfg = &env.FHIR{
		HostAPI:    "http://hostapi",
		HostSearch: "http://hostsearch",
		User:       "Test user",
	}
	s.fhirClient = client.New(s.fhirClientCfg).WithHTTPClient(s.httpClient)
}

func (s *clientTestSuite) TearDownTest() {
}

func (s *clientTestSuite) TearDownSuite() {
}

func (s *clientTestSuite) TestGetResourceByID() {
	ctx := context.Background()
	resName := "Task"
	resID := fhir.ID("123")
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/%s/%s", s.fhirClientCfg.HostSearch, resName, resID), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		return nil
	}

	err := s.fhirClient.GetResourceByID(ctx, resName, resID, nil)
	s.NoError(err)
}

func (s *clientTestSuite) TestSearchResourceByParams() {
	ctx := context.Background()
	resName := "Task"
	qParams := []*client.QParam{
		{Key: "status", Value: "ok"},
		{Key: "id", Value: "1"},
	}

	var qParamsStr string

	for _, v := range qParams {
		qParamsStr += v.Key + "=" + v.Value + "&"
	}

	qParamsStr = strings.Trim(qParamsStr, "&")

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/%s?%s", s.fhirClientCfg.HostSearch, resName, qParamsStr), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		s.Equal("strict", string(req.Header.Peek("prefer_handling")))
		resp.SetStatusCode(fasthttp.StatusOK)

		_, _ = resp.BodyWriter().Write([]byte("{}"))

		return nil
	}

	_, err := s.fhirClient.SearchResourceByParams(ctx, resName, qParams)
	s.NoError(err)
}

func (s *clientTestSuite) TestCreateBundle() {
	ctx := context.Background()

	b := &fhir.Bundle{ID: fhir.ID("123").String()}

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(s.fhirClientCfg.HostAPI, strings.Trim(req.URI().String(), `/`))
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))
		s.Equal("respond-sync", string(req.Header.Peek("prefer")))

		actualB := new(fhir.Bundle)
		err := json.Unmarshal(req.Body(), actualB)

		s.NoError(err)
		s.Equal(b, actualB)
		resp.SetStatusCode(fasthttp.StatusOK)

		_, _ = resp.BodyWriter().Write([]byte("{}"))

		return nil
	}

	_, err := s.fhirClient.CreateBundle(ctx, b)
	s.NoError(err)
}

func (s *clientTestSuite) TestCreateResource() {
	ctx := context.Background()
	prac := new(fhir.Practitioner)

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/%s", s.fhirClientCfg.HostAPI, prac.ResourceName()), strings.Trim(req.URI().String(), `/`))
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))

		actualReq := map[string]interface{}{}

		err := json.Unmarshal(req.Body(), &actualReq)
		s.NoError(err)

		resp.SetStatusCode(fasthttp.StatusOK)
		_, _ = resp.BodyWriter().Write([]byte("{}"))

		return nil
	}

	err := s.fhirClient.CreateResource(ctx, prac)
	s.NoError(err)
}

func (s *clientTestSuite) TestPatchResource() {
	ctx := context.Background()
	resourceName := "Resource"
	body := []*client.PatchBody{}
	id := uuid.NewV4().String()

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/%s/%s", s.fhirClientCfg.HostAPI, resourceName, id), strings.Trim(req.URI().String(), `/`))
		s.Equal(fasthttp.MethodPatch, string(req.Header.Method()))

		resp.SetStatusCode(fasthttp.StatusOK)
		_, _ = resp.BodyWriter().Write([]byte("{}"))

		return nil
	}

	err := s.fhirClient.PatchResource(ctx, resourceName, fhir.ID(id), body)
	s.NoError(err)
}

func (s *clientTestSuite) TestPutResource() {
	ctx := context.Background()
	resourceName := "Resource"
	expectedDst := &fhir.Parameters{ID: "12345"}
	id := uuid.NewV4().String()
	body := &fhir.Resource{ID: id}
	dst := new(fhir.Parameters)
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/%s/%s", s.fhirClientCfg.HostAPI, resourceName, id), strings.Trim(req.URI().String(), `/`))
		s.Equal(fasthttp.MethodPut, string(req.Header.Method()))

		resp.SetStatusCode(fasthttp.StatusOK)

		b, _ := json.Marshal(expectedDst)
		_, _ = resp.BodyWriter().Write(b)

		return nil
	}

	err := s.fhirClient.PutResource(ctx, body, dst)
	s.NoError(err)
	s.Equal(expectedDst, dst)
}

func (s *clientTestSuite) TestValidateResource() {
	ctx := context.Background()
	resName := "Bundle"
	b := &fhir.Bundle{ID: fhir.ID("123").String()}

	severity := fhir.IssueSeverity("warn")
	code := fhir.IssueType("123")
	oo := fhir.OperationOutcome{Issue: []*fhir.OperationOutcomeIssue{{Severity: &severity, Code: &code}}}
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/%s/$validate", s.fhirClientCfg.HostAPI, resName), req.URI().String())
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))

		actualB := new(fhir.Bundle)
		err := json.Unmarshal(req.Body(), actualB)

		s.NoError(err)
		s.Equal(b, actualB)
		resp.SetStatusCode(fasthttp.StatusOK)

		b, _ := json.Marshal(oo)
		_, _ = resp.BodyWriter().Write(b)

		return nil
	}

	actualOO, err := s.fhirClient.ValidateResource(ctx, b)
	s.NoError(err)
	s.Equal(&oo, actualOO)

	severity = "error"
	oo = fhir.OperationOutcome{Issue: []*fhir.OperationOutcomeIssue{{Severity: &severity, Code: &code}}}
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		resp.SetStatusCode(fasthttp.StatusOK)

		b, _ := json.Marshal(oo)
		_, _ = resp.BodyWriter().Write(b)

		return nil
	}

	_, err = s.fhirClient.ValidateResource(ctx, b)
	s.Error(err)

	cErr, ok := err.(*cerror.CError)

	s.True(ok)
	s.Equal(cerror.KindBadValidation.String(), cErr.Kind().String())
	s.Equal(&oo, cErr.Payload())
}

func (s *clientTestSuite) TestSendRequest() {
	ctx := context.Background()

	qParams := []*client.QParam{
		{Key: "status", Value: "ok"},
		{Key: "id", Value: "1"},
	}

	var qParamsStr string

	for _, v := range qParams {
		qParamsStr += v.Key + "=" + v.Value + "&"
	}

	qParamsStr = strings.Trim(qParamsStr, "&")

	headers := map[string]string{"k1": "v1", "k2": "v2"}
	b := &fhir.Bundle{ID: fhir.ID("123").String()}
	actualB := new(fhir.Bundle)
	r := &client.Request{
		URL:         "http://host:3000/api",
		Method:      fasthttp.MethodPut,
		QueryParams: qParams,
		Headers:     headers,
		Body:        b,
		Destination: actualB,
	}

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s?%s", r.URL, qParamsStr), req.URI().String())
		s.Equal(r.Method, string(req.Header.Method()))
		s.Equal("application/json", string(req.Header.Peek(fasthttp.HeaderContentType)))
		s.Equal(s.fhirClientCfg.User, string(req.Header.Peek("x-consumer-id")))

		for k, v := range r.Headers {
			s.Equal(v, string(req.Header.Peek(k)))
		}

		reqBody := new(fhir.Bundle)
		err := json.Unmarshal(req.Body(), reqBody)

		s.NoError(err)
		s.Equal(b, reqBody)

		resp.SetStatusCode(fasthttp.StatusOK)

		b, _ := json.Marshal(reqBody)
		_, _ = resp.BodyWriter().Write(b)

		return nil
	}

	err := s.fhirClient.SendRequest(ctx, r)
	s.NoError(err)
	s.Equal(b, actualB)

	severity := fhir.IssueSeverity("error")
	code := fhir.IssueType("123")
	oo := fhir.OperationOutcome{Issue: []*fhir.OperationOutcomeIssue{{Severity: &severity, Code: &code}}}
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		resp.SetStatusCode(fasthttp.StatusBadRequest)

		b, _ := json.Marshal(oo)
		_, _ = resp.BodyWriter().Write(b)

		return nil
	}

	err = s.fhirClient.SendRequest(ctx, r)
	s.Error(err)

	cErr, ok := err.(*cerror.CError)

	s.True(ok)
	s.Equal(cerror.KindBadParams.String(), cErr.Kind().String())
	s.Equal(&oo, cErr.Payload())
}
