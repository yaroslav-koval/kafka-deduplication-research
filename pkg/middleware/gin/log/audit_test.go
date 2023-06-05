package log_test

import (
	"bytes"
	"encoding/json"
	"kafka-polygon/pkg/http/consts"
	pkglog "kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/middleware"
	"kafka-polygon/pkg/middleware/gin/log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestMain(m *testing.M) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Run the other tests
	os.Exit(m.Run())
}

func TestAudit(t *testing.T) {
	t.Parallel()

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

	app := gin.Default()

	app.Use(log.Audit(middleware.NewAuditOptions().WithServiceName(expected.Service).WithResponse()))

	app.GET(expected.URL, func(c *gin.Context) {
		c.Set(consts.HeaderXUserID, expected.UserID)
		c.Set(consts.HeaderXClientID, expected.ClientID)
		c.Set(consts.HeaderXRequestID, expected.RequestID)
		if c.GetHeader("isError") != "" {
			c.Status(500)
			return
		}

		c.String(http.StatusOK, expected.Response)
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
		assert.Nil(t, err)

		req.Header.Set("User-Agent", expected.UserAgent)

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusOK, w.Code)

		var actual testLogMsg

		_ = json.Unmarshal(buf.Bytes(), &actual)
		expected.Latency = actual.Latency // exclude from comparing

		assert.Equal(t, expected, actual)
		assert.NotEqual(t, 0, actual.Latency)
	})

	t.Run("error", func(t *testing.T) {
		var actual testLogMsg
		buf.Reset()

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
		assert.Nil(t, err)

		req.Header.Set("isError", "true")

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		_ = json.Unmarshal(buf.Bytes(), &actual)
		assert.Equal(t, "error", actual.Level)
	})
}
