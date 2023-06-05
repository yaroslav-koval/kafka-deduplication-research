package tracing_test

import (
	pkgtracing "kafka-polygon/pkg/middleware/gin/tracing"
	"kafka-polygon/pkg/tracing"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tj/assert"
)

var testEmptyHandler = func(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func TestMain(m *testing.M) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Run the other tests
	os.Exit(m.Run())
}

func TestHandlerTrace(t *testing.T) {
	t.Parallel()

	app := gin.Default()

	// tracing
	tt := tracing.New(tracing.NewProviderJaeger(tracing.JaegerConfig{UseAgent: false, URL: ""}))

	app.Use(pkgtracing.New(tt))
	app.POST("/ok", testEmptyHandler)
	app.OPTIONS("/ok", testEmptyHandler)

	t.Run("Send a good request body tracer", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/ok", http.NoBody)
		assert.Nil(t, err)

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
	})

	app.POST("/err", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	t.Run("Send a bad request body tracer", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/err", http.NoBody)
		assert.Nil(t, err)

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
	})
}
