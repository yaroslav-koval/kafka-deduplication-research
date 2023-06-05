package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	logs "kafka-polygon/pkg/log"
	"kafka-polygon/pkg/middleware"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestAuditOptions(t *testing.T) {
	t.Parallel()

	o := middleware.NewAuditOptions()

	assert.Empty(t, o.ServiceName())
	assert.False(t, o.IsWithResponse())

	o.WithServiceName("srv").WithResponse()

	assert.Equal(t, "srv", o.ServiceName())
	assert.True(t, o.IsWithResponse())
}

type testLogMsg struct {
	Level       string `json:"level"`
	Status      int    `json:"status"`
	Method      string `json:"method"`
	Latency     int    `json:"latency"`
	URL         string `json:"url"`
	LogType     string `json:"logType"`
	ServiceName string `json:"serviceName"`
	UserAgent   string `json:"userAgent"`
	RequestID   string `json:"requestID"`
	ClientID    string `json:"clientID"`
	UserID      string `json:"userID"`
	Response    string `json:"response"`
}

func TestRenderAuditLogger(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	logs.SetGlobalOutput(buf)
	logs.SetGlobalLogLevel("debug")

	tests := []struct {
		code        int
		name, level string
	}{
		{
			name:  "ok",
			level: "info",
			code:  http.StatusOK,
		},
		{
			name:  "not-found",
			level: "warn",
			code:  http.StatusNotFound,
		},
		{
			name:  "error",
			level: "error",
			code:  http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		respBody := fmt.Sprintf("test-response-body-%s", test.name)
		expAuditData := middleware.AuditData{
			ServiceName: fmt.Sprintf("test-%s", test.name),
			StatusCode:  test.code,
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("/test-%s", test.name),
			UserAgent:   fmt.Sprintf("test-user-agent-%s", test.name),
			Headers: &middleware.AuditHeaders{
				XRequestID: fmt.Sprintf("test-x-request-id-%s", test.name),
				XUserID:    fmt.Sprintf("test-x-user-id-%s", test.name),
				XClientID:  fmt.Sprintf("test-x-client-id-%s", test.name),
			},
			ResponseBody: bytes.NewBuffer([]byte(respBody)),
		}

		err := middleware.RenderAuditLogger(ctx, func() (*middleware.AuditData, error) {
			buf.Reset()
			return &expAuditData, nil
		})
		require.NoError(t, err)

		var actualMsg testLogMsg

		_ = json.Unmarshal(buf.Bytes(), &actualMsg)

		assert.Equal(t, test.level, actualMsg.Level)
		assert.Equal(t, expAuditData.StatusCode, actualMsg.Status)
		assert.Equal(t, expAuditData.Method, actualMsg.Method)
		assert.Equal(t, expAuditData.UserAgent, actualMsg.UserAgent)
		assert.Equal(t, expAuditData.ServiceName, actualMsg.ServiceName)
		assert.Equal(t, expAuditData.Headers.XRequestID, actualMsg.RequestID)
		assert.Equal(t, expAuditData.Headers.XClientID, actualMsg.ClientID)
		assert.Equal(t, expAuditData.Headers.XUserID, actualMsg.UserID)
		assert.Equal(t, expAuditData.ResponseBody.String(), actualMsg.Response)
	}
}

func TestRenderAuditLoggerError(t *testing.T) {
	ctx := context.Background()
	expErr := errors.New("renderAuditLogger err")

	err := middleware.RenderAuditLogger(ctx, func() (*middleware.AuditData, error) {
		return &middleware.AuditData{}, expErr
	})
	require.Error(t, err)
	assert.Equal(t, expErr, err)
}
