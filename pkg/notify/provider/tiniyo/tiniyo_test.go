package tiniyo_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/notify/provider/tiniyo"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type clientTestSuite struct {
	suite.Suite
	tiniyoClient  *tiniyo.Client
	tiniyoOptions *tiniyo.Options
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
	s.tiniyoOptions = &tiniyo.Options{
		HostAPI:   "https://api.tiniyo.com/v1/Accounts/%s/Messages",
		AuthID:    "authID",
		AuthToken: "authToken",

		Body: "Test message",
	}
	s.tiniyoClient = tiniyo.New(s.tiniyoOptions).WithHTTPClient(s.httpClient)
}

func (s *clientTestSuite) TearDownTest() {
}

func (s *clientTestSuite) TearDownSuite() {
}

func (s *clientTestSuite) TestSend() {
	ctx := context.Background()

	var err error

	_, err = s.tiniyoClient.Send(ctx)
	s.Error(err)
	s.Equal("not empty To", err.Error())

	s.tiniyoClient.To("380672200333")

	_, err = s.tiniyoClient.Send(ctx)
	s.Error(err)
	s.Equal("not empty From", err.Error())

	s.tiniyoClient.From("TINIYO")

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		var payload []byte

		s.Equal(fmt.Sprintf(s.tiniyoOptions.HostAPI, s.tiniyoOptions.AuthID), req.URI().String())
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Basic ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		payload, err = base64.StdEncoding.DecodeString(string(auth[len(basicAuthPrefix):]))
		s.NoError(err)
		s.Equal("authID:authToken", string(payload))

		dst := new(tiniyo.MessageRequestBody)
		err = json.Unmarshal(req.Body(), dst)
		s.NoError(err)
		s.Equal("TINIYO", dst.From)
		s.Equal("380672200333", dst.To)
		s.Equal("Test message", dst.Body)

		resp.SetStatusCode(fasthttp.StatusOK)

		_, _ = resp.BodyWriter().Write([]byte("{}"))

		return nil
	}

	_, err = s.tiniyoClient.Send(ctx)
	s.NoError(err)

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		resp.SetStatusCode(fasthttp.StatusInternalServerError)

		errResp := tiniyo.MessageErrorResponseBody{
			Message: "error message",
			Code:    500,
			Status:  5001,
		}
		b, _ := json.Marshal(errResp)
		_, _ = resp.BodyWriter().Write(b)

		return nil
	}

	resp, err := s.tiniyoClient.Send(ctx)
	s.NoError(err)

	s.Equal("error message", resp.Message)
	s.EqualValues(500, resp.ErrorCode)
	s.Equal(5001, resp.ResponseStatusCode)
}
