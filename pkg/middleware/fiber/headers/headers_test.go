package headers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/middleware/fiber/headers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/tj/assert"
)

func TestRequestHeadersToContext(t *testing.T) {
	app := fiber.New()

	app.Use(headers.RequestHeadersToContext(consts.RequestHeadersToSave()))
	app.Use(headers.ResponseHeadersFromContext(consts.RequestHeadersToSave()))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for _, h := range consts.RequestHeadersToSave() {
		req.Header.Set(h, h)
	}

	resp, err := app.Test(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err)

	for _, h := range consts.RequestHeadersToSave() {
		assert.Equal(t, h, resp.Header.Get(h))
	}

	// check auto set of request id if request header was not set
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if resp != nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err)

	assert.NotEmpty(t, resp.Header.Get(consts.HeaderXRequestID))
}

func TestResponseHeadersFromContext(t *testing.T) {
	app := fiber.New()

	app.Use(headers.ResponseHeadersFromContext(consts.ResponseHeadersToSend()))

	app.Get("/", func(c *fiber.Ctx) error {
		for _, h := range consts.ResponseHeadersToSend() {
			c.Context().SetUserValue(h, h)
		}
		return c.SendString("")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	resp, err := app.Test(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err)

	for _, h := range consts.ResponseHeadersToSend() {
		assert.Equal(t, h, resp.Header.Get(h))
	}
}

func TestValidateContentType(t *testing.T) {
	app := fiber.New()

	app.Use(headers.ValidateJSONContentType(http.MethodGet))

	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendString("")
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("")
	})

	// check without content type
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	resp, err := app.Test(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)

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
		fmt.Sprintf("content-type header should be %v. Current value: ", fiber.MIMEApplicationJSON),
		m.Error.Message)

	// check with valid content type
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err = app.Test(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// check without content type for method that doesn't require content type
	req = httptest.NewRequest(http.MethodGet, "/", nil)

	resp, err = app.Test(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
