package log_test

import (
	"bytes"
	"encoding/json"
	"kafka-polygon/pkg/http/consts"
	pkglog "kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/middleware"
	"kafka-polygon/pkg/middleware/fiber/log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/tj/assert"
)

type testLogMsg struct {
	UserID    string `json:"userID"`
	ClientID  string `json:"clientID"`
	RequestID string `json:"requestID"`
	Method    string `json:"method"`
	Status    int    `json:"status"`
	Level     string `json:"level"`
	Latency   int    `json:"latency"`
	URL       string `json:"url"`
	LogType   string `json:"logType"`
	Service   string `json:"serviceName"`
	UserAgent string `json:"userAgent"`
	Response  string `json:"response"`
}

func TestAudit(t *testing.T) {
	buf := new(bytes.Buffer)
	pkglog.SetGlobalOutput(buf)

	expected := testLogMsg{
		UserID:    "user-id",
		ClientID:  "client-id",
		RequestID: "req-id",
		Method:    http.MethodGet,
		Status:    http.StatusOK,
		Level:     logger.LevelInfo.String(),
		URL:       "/",
		LogType:   "audit",
		Service:   "srv",
		UserAgent: "test agent",
		Response:  "response",
	}

	app := fiber.New()

	app.Use(log.Audit(middleware.NewAuditOptions().WithServiceName(expected.Service).WithResponse()))

	app.Get(expected.URL, func(c *fiber.Ctx) error {
		c.Context().SetUserValue(consts.HeaderXUserID, expected.UserID)
		c.Context().SetUserValue(consts.HeaderXClientID, expected.ClientID)
		c.Context().SetUserValue(consts.HeaderXRequestID, expected.RequestID)
		if c.Get("isError") != "" {
			return c.SendStatus(500)
		}
		return c.SendString(expected.Response)
	})

	req := httptest.NewRequest(http.MethodGet, expected.URL, nil)

	req.Header.Set(fiber.HeaderUserAgent, expected.UserAgent)

	resp, err := app.Test(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err)

	var actual testLogMsg

	_ = json.Unmarshal(buf.Bytes(), &actual)
	expected.Latency = actual.Latency // exclude from comparing

	assert.Equal(t, expected, actual)
	assert.NotEqual(t, 0, actual.Latency)

	buf.Reset()
	req.Header.Set("isError", "true")

	resp, _ = app.Test(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	_ = json.Unmarshal(buf.Bytes(), &actual)

	assert.Equal(t, "error", actual.Level)
}
