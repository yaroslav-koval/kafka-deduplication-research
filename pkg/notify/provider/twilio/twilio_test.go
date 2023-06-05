package twilio_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/notify/provider/twilio"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"github.com/twilio/twilio-go/client"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

var (
	bgCtx                = context.Background()
	expTo                = "+380123456789"
	expFrom              = "+480123456789"
	expBody              = "Test Twilio"
	expMessageID         = "test-message-id"
	expMessageServiceSID = "test-message-service-sid"
	expAccountSID        = "AC222222222222222222222222222222"
)

type MockBaseClient struct {
	mock.Mock
	Tw      *client.Client
	BaseURL string
}

func (mbc *MockBaseClient) AccountSid() string {
	args := mbc.Called()
	return args.String(0)
}

func (mbc *MockBaseClient) SetTimeout(timeout time.Duration) {
	_ = mbc.Called(timeout)
}

func (mbc *MockBaseClient) SendRequest(method string, rawURL string, data url.Values,
	headers map[string]interface{}) (*http.Response, error) {
	if mbc.Tw == nil {
		return nil, errors.New("not set tw client")
	}

	return mbc.Tw.SendRequest(method, mbc.BaseURL, data, headers)
}

func TestTwilio(t *testing.T) {
	t.Parallel()

	mockServer := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			d := map[string]interface{}{
				"response": "ok",
			}
			encoder := json.NewEncoder(writer)
			_ = encoder.Encode(&d)
		}))

	defer mockServer.Close()

	tw := twilio.New(&twilio.Options{
		BaseClient: getTestClient(t),
	})
	actualResp, err := tw.From(expFrom).
		To(expTo).
		WithBody(expBody).Send(bgCtx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, actualResp.ResponseStatusCode)
	assert.Equal(t, expMessageID, actualResp.MessageID)
}

func TestTwilioNoEmptyToError(t *testing.T) {
	t.Parallel()

	tw := twilio.New(&twilio.Options{})
	_, err := tw.From(expFrom).
		WithBody(expBody).Send(bgCtx)
	require.Error(t, err)
	assert.Equal(t, "not empty To", err.Error())
}

func TestTwilioNoEmptyFromError(t *testing.T) {
	t.Parallel()

	tw := twilio.New(&twilio.Options{})
	_, err := tw.To(expTo).
		WithBody(expBody).Send(bgCtx)
	require.Error(t, err)
	assert.Equal(t, "not empty From", err.Error())
}

func TestTwilioError(t *testing.T) {
	t.Parallel()

	errorResponse := `{
	"status": 400,
	"code":20001,
	"message":"Bad request",
	"more_info":"https://www.twilio.com/docs/errors/20001"
}`
	errorServer := httptest.NewServer(http.HandlerFunc(
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(400)
			_, _ = resp.Write([]byte(errorResponse))
		}))

	defer errorServer.Close()

	mockBaseClient := &MockBaseClient{
		BaseURL: errorServer.URL,
		Tw: &client.Client{
			Credentials: client.NewCredentials(expAccountSID, "test-auth-token"),
			HTTPClient:  http.DefaultClient,
		},
	}
	mockBaseClient.On("AccountSid").Return(expAccountSID)

	tw := twilio.New(&twilio.Options{
		BaseClient: mockBaseClient,
	})
	actualResp, err := tw.From(expFrom).
		To(expTo).
		WithBody(expBody).Send(bgCtx)
	require.Error(t, err)
	assert.Equal(t, "Status: 400 - ApiError 20001: Bad request (null) More info: https://www.twilio.com/docs/errors/20001", err.Error())
	assert.Equal(t, http.StatusBadRequest, actualResp.ResponseStatusCode)
	assert.Equal(t, int64(20001), actualResp.ErrorCode)
}

func getTestClient(t *testing.T) *client.MockBaseClient {
	now := time.Now().UTC()
	mockCtrl := gomock.NewController(t)
	testClient := client.NewMockBaseClient(mockCtrl)
	testClient.EXPECT().AccountSid().DoAndReturn(func() string {
		return expAccountSID
	}).AnyTimes()

	testClient.EXPECT().SendRequest(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).
		DoAndReturn(func(method string, rawURL string, data url.Values,
			headers map[string]interface{}) (*http.Response, error) {
			response := openapi.ApiV2010Message{
				AccountSid:          converto.StringPointer(expAccountSID),
				ApiVersion:          converto.StringPointer("0.0.1"),
				Body:                converto.StringPointer("test-body"),
				DateCreated:         converto.StringPointer(now.String()),
				DateSent:            converto.StringPointer(now.String()),
				DateUpdated:         converto.StringPointer(now.String()),
				Direction:           converto.StringPointer("test-direction"),
				ErrorCode:           converto.IntPointer(2012),
				ErrorMessage:        converto.StringPointer("test-error-message"),
				From:                converto.StringPointer(expFrom),
				To:                  converto.StringPointer(expTo),
				MessagingServiceSid: converto.StringPointer(expMessageServiceSID),
				Sid:                 converto.StringPointer(expMessageID),
			}

			resp, _ := json.Marshal(response)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(resp)),
			}, nil
		},
		)

	return testClient
}
