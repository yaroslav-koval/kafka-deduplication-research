package fiber_test

import (
	"context"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/http/consts"
	httpFiber "kafka-polygon/pkg/http/fiber"
	"kafka-polygon/pkg/http/mock"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestCustomBodyDecoder(t *testing.T) {
	t.Parallel()

	srvCfg := env.Service{
		Name: "test",
	}
	hsrvCfg := env.HTTPServer{
		Port: "8000",
	}

	server := httpFiber.NewServer(&httpFiber.ServerConfig{
		Service: srvCfg,
		Server:  hsrvCfg,
	}).WithHealthCheckRoute()

	t.Run("Send a good request body health check", func(t *testing.T) {
		resp, err := server.Fiber().Test(httptest.NewRequest(fiber.MethodGet, "/health", nil))
		utils.AssertEqual(t, nil, err)
		defer resp.Body.Close()
		utils.AssertEqual(t, http.StatusOK, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
	})

	server.Fiber().Post("/error_500", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusInternalServerError)
	})
	t.Run("Send a bad request body err", func(t *testing.T) {
		resp, err := server.Fiber().Test(httptest.NewRequest(fiber.MethodPost, "/error_500", nil))
		utils.AssertEqual(t, nil, err)
		defer resp.Body.Close()
		utils.AssertEqual(t, http.StatusInternalServerError, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
	})
}

func TestFiberSwagger(t *testing.T) {
	t.Parallel()

	srvCfg := env.Service{
		Name: "test-swagger",
	}
	hsrvCfg := env.HTTPServer{
		Port: "8001",
	}

	swagger, err := mock.GetSwagger()
	require.NoError(t, err)

	server := httpFiber.NewServer(&httpFiber.ServerConfig{
		Service: srvCfg,
		Server:  hsrvCfg,
		Swagger: swagger,
	})

	server.WithGetSpecRoute()

	t.Run("Send a good request swagger spec", func(t *testing.T) {
		resp, err := server.Fiber().Test(httptest.NewRequest(fiber.MethodGet, "/spec", nil))
		utils.AssertEqual(t, nil, err)
		defer resp.Body.Close()
		utils.AssertEqual(t, http.StatusOK, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
	})
}

func TestFiberWithDefaultKit(t *testing.T) {
	t.Parallel()

	srvCfg := env.Service{
		Name: "test-default-kit",
	}
	hsrvCfg := env.HTTPServer{
		Port: "8002",
	}

	swagger, err := mock.GetSwagger()
	require.NoError(t, err)

	server := httpFiber.NewServer(&httpFiber.ServerConfig{
		Service: srvCfg,
		Server:  hsrvCfg,
		Swagger: swagger,
	})

	server.WithDefaultKit()

	t.Run("Send a good request default kit", func(t *testing.T) {
		resp, err := server.Fiber().Test(httptest.NewRequest(fiber.MethodGet, "/health", nil))
		utils.AssertEqual(t, nil, err)
		defer resp.Body.Close()
		utils.AssertEqual(t, http.StatusOK, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
	})

	t.Run("Send a good request default kit swagger", func(t *testing.T) {
		resp, err := server.Fiber().Test(httptest.NewRequest(fiber.MethodGet, "/spec", nil))
		utils.AssertEqual(t, nil, err)
		defer resp.Body.Close()

		utils.AssertEqual(t, http.StatusOK, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
		utils.AssertEqual(t, "application/json", resp.Header.Get(fiber.HeaderContentType))

		i, err := strconv.ParseInt(resp.Header.Get(fiber.HeaderContentLength), 10, 64)
		require.NoError(t, err)

		assert.True(t, i > 0)
	})

	t.Run("Send a bad_request default kit", func(t *testing.T) {
		resp, err := server.Fiber().Test(httptest.NewRequest(fiber.MethodGet, "/404", nil))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp.Body)
		utils.AssertEqual(t, http.StatusBadRequest, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
		utils.AssertEqual(t, "application/json", resp.Header.Get(fiber.HeaderContentType))

		i, err := strconv.ParseInt(resp.Header.Get(fiber.HeaderContentLength), 10, 64)
		require.NoError(t, err)

		assert.True(t, i > 0)
		assert.NotEqual(t, "", resp.Header.Get(consts.HeaderXRequestID))
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
		server := httpFiber.NewServer(&httpFiber.ServerConfig{
			Service: srvCfg,
			Server:  hsrvCfg,
		})
		utils.AssertEqual(t, true, server.Shutdown(context.Background()) == nil)
	})
}
