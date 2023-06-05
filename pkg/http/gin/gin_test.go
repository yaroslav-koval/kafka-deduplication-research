package gin_test

import (
	"context"
	"errors"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/http/consts"
	httpGin "kafka-polygon/pkg/http/gin"
	"kafka-polygon/pkg/http/mock"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestMain(m *testing.M) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Run the other tests
	os.Exit(m.Run())
}

func TestCustomBodyDecoder(t *testing.T) {
	t.Parallel()

	srvCfg := env.Service{
		Name: "test",
	}
	hsrvCfg := env.HTTPServer{
		Port: "8001",
	}

	server := httpGin.NewServer(&httpGin.ServerConfig{
		Service: srvCfg,
		Server:  hsrvCfg,
	}).WithHealthCheckRoute()

	t.Run("Send a good request body health check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/health", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
	})

	server.Gin().POST("/error_500", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	t.Run("Send a bad request body err", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/error_500", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
	})
}

func TestGinSwagger(t *testing.T) {
	t.Parallel()

	srvCfg := env.Service{
		Name: "test-swagger",
	}
	hsrvCfg := env.HTTPServer{
		Port: "8001",
	}

	swagger, err := mock.GetSwagger()
	require.NoError(t, err)

	server := httpGin.NewServer(&httpGin.ServerConfig{
		Service: srvCfg,
		Server:  hsrvCfg,
		Swagger: swagger,
	})

	server.WithGetSpecRoute()

	t.Run("Send a good request swagger spec", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/spec", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
	})

	server.Gin().GET("/def_spec", httpGin.DefaultGetSpecHandler(swagger))
	t.Run("Send a good request default spec", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/def_spec", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, gin.MIMEJSON, w.Header().Get("Content-Type"))
		assert.Equal(t, "", w.Header().Get("Allow"))
	})
}

func TestGinWithDefaultKit(t *testing.T) {
	t.Parallel()

	srvCfg := env.Service{
		Name: "test-default-kit",
	}
	hsrvCfg := env.HTTPServer{
		Port: "8002",
	}

	swagger, err := mock.GetSwagger()
	require.NoError(t, err)

	server := httpGin.NewServer(&httpGin.ServerConfig{
		Service: srvCfg,
		Server:  hsrvCfg,
		Swagger: swagger,
	})

	server.WithDefaultKit()

	t.Run("Send a good request default kit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/health", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
	})

	t.Run("Send a good request default kit swagger", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/spec", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("Send a bad_request default kit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/404", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.NotNil(t, w.Body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "", w.Header().Get("Allow"))
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
		assert.NotEqual(t, "", w.Header().Get(consts.HeaderXRequestID))
	})
}

func TestCloseOnShutdown(t *testing.T) {
	t.Parallel()

	srvCfg := env.Service{
		Name: "test",
	}
	hsrvCfg := env.HTTPServer{
		Port:            "8000",
		CloseOnShutdown: true,
		ReadTimeoutSec:  60,
		WriteTimeoutSec: 30,
	}

	t.Run("success", func(t *testing.T) {
		server := httpGin.NewServer(&httpGin.ServerConfig{
			Service: srvCfg,
			Server:  hsrvCfg,
		})
		assert.Equal(t, true, server.Shutdown(context.Background()) == nil)
	})
}

func TestDefaultHealthCheckHandler(t *testing.T) {
	t.Parallel()

	gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	httpGin.DefaultHealthCheckHandler(gCtx)

	assert.Equal(t, http.StatusOK, gCtx.Writer.Status())
}

func TestHandlerFunc(t *testing.T) {
	t.Parallel()

	expErr := errors.New("handler error")

	srvCfg := env.Service{
		Name: "test-handler-func",
	}
	hsrvCfg := env.HTTPServer{
		Port: "8004",
	}

	server := httpGin.NewServer(&httpGin.ServerConfig{
		Service: srvCfg,
		Server:  hsrvCfg,
	}).WithDefaultKit()

	server.Gin().GET("/", httpGin.HandlerFunc(func(c *gin.Context) error {
		c.Status(http.StatusOK)
		return nil
	}))

	server.Gin().GET("/handler_error", httpGin.HandlerFunc(func(c *gin.Context) error {
		return expErr
	}))

	t.Run("Send a check ok", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Send a check system error", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodGet, "/handler_error", http.NoBody)
		require.NoError(t, err)
		defer w.Body.Reset()

		server.Gin().ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, "", w.Header().Get("Allow"))
		assert.Equal(t, "{\"error\":{\"message\":\"handler error\",\"type\":\"other_error\",\"group\":\"http\"}}", w.Body.String())
	})
}
