package headers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/middleware/gin/headers"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tj/assert"
)

func TestMain(m *testing.M) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Run the other tests
	os.Exit(m.Run())
}

func TestRequestHeadersToContext(t *testing.T) {
	t.Parallel()

	app := gin.Default()

	app.Use(headers.RequestHeadersToContext(consts.RequestHeadersToSave()))
	app.Use(headers.ResponseHeadersFromContext(consts.RequestHeadersToSave()))

	app.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
		assert.Nil(t, err)

		for _, h := range consts.RequestHeadersToSave() {
			req.Header.Set(h, h)
		}

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusOK, w.Code)
		for _, h := range consts.RequestHeadersToSave() {
			assert.Equal(t, h, w.Header().Get(h))
		}
	})

	t.Run("error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
		assert.Nil(t, err)

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get(consts.HeaderXRequestID))
	})
}

func TestResponseHeadersFromContext(t *testing.T) {
	t.Parallel()

	app := gin.Default()

	app.Use(headers.ResponseHeadersFromContext(consts.ResponseHeadersToSend()))

	app.GET("/", func(c *gin.Context) {
		for _, h := range consts.ResponseHeadersToSend() {
			c.Header(h, h)
		}
		c.String(http.StatusOK, "")
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	assert.Nil(t, err)

	app.ServeHTTP(w, req)
	defer req.Body.Close()

	for _, h := range consts.ResponseHeadersToSend() {
		assert.Equal(t, h, w.Header().Get(h))
	}
}

func TestValidateContentType(t *testing.T) {
	t.Parallel()

	app := gin.Default()

	app.Use(headers.ValidateJSONContentType(http.MethodGet))

	app.POST("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})
	app.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	t.Run("check without content type", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/", http.NoBody)
		assert.Nil(t, err)

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)

		b, err := io.ReadAll(w.Body)

		assert.NoError(t, err)

		m := struct {
			Error struct {
				Message string
				Type    string
			}
		}{}

		_ = json.Unmarshal(b, &m)

		assert.Equal(t, "invalid_content_type", m.Error.Type)
		assert.Equal(
			t,
			fmt.Sprintf("content-type header should be %v. Current value: ", gin.MIMEJSON),
			m.Error.Message)
	})

	t.Run("check with valid content type", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Content-Type", gin.MIMEJSON)
		assert.Nil(t, err)

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("check without content type for method that doesn't require content type", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
		assert.Nil(t, err)

		app.ServeHTTP(w, req)
		defer req.Body.Close()

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
