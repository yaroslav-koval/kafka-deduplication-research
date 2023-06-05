package tracing_test

import (
	pkgtracing "kafka-polygon/pkg/middleware/fiber/tracing"
	"kafka-polygon/pkg/tracing"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

var testEmptyHandler = func(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

func TestHandlerTrace(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// tracing
	tt := tracing.New(tracing.NewProviderJaeger(tracing.JaegerConfig{UseAgent: false, URL: ""}))

	app.Use(pkgtracing.New(tt))
	app.Post("/ok", testEmptyHandler)
	app.Options("/ok", testEmptyHandler)

	t.Run("Send a good request body tracer", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/ok", nil))
		utils.AssertEqual(t, nil, err)
		defer resp.Body.Close()
		utils.AssertEqual(t, http.StatusNoContent, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
	})

	app.Post("/err", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusInternalServerError)
	})
	t.Run("Send a bad request body tracer", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/err", nil))
		utils.AssertEqual(t, nil, err)
		defer resp.Body.Close()
		utils.AssertEqual(t, http.StatusInternalServerError, resp.StatusCode)
		utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderAllow))
	})
}
