package openapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	middlewareOapi "kafka-polygon/pkg/middleware/fiber/openapi"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var testSchema = `openapi: "3.0.0"
info:
  version: 1.0.0
  title: TestServer
servers:
  - url: http://deepmap.ai
paths:
  /resource:
    get:
      operationId: getResource
      parameters:
        - name: id
          in: query
          schema:
            type: integer
            minimum: 10
            maximum: 100
      responses:
        '200':
          description: success
          content:
            application/json:
              schema:
                properties:
                  name:
                    type: string
                  id:
                    type: integer
    post:
      operationId: createResource
      responses:
        '204':
          description: No content
      requestBody:
        required: true
        content:
          application/json:
            schema:
              properties:
                name:
                  type: string
  /protected_resource:
    get:
      operationId: getProtectedResource
      security:
        - BearerAuth:
            - someScope
      responses:
        '204':
          description: no content
  /protected_resource2:
    get:
      operationId: getProtectedResource2
      security:
        - BearerAuth:
            - otherScope
      responses:
        '204':
          description: no content
  /protected_resource_401:
    get:
      operationId: getProtectedResource401
      security:
        - BearerAuth:
            - unauthorized
      responses:
        '401':
          description: no content
  /resource_uuid:
    post:
      operationId: postResourceUuid
      requestBody:
        required: true
        content:
          'application/json':
            schema:
              $ref: '#/components/schemas/fieldUUID'
      responses:
        '201':
          description: Created
        '500':
          description: General error
  /resource_array:
    post:
      operationId: postResourceArray
      requestBody:
        required: true
        content:
          'application/json':
            schema:
              $ref: '#/components/schemas/ArrayInput'
      responses:
        '200':
          description: Created
        '500':
          description: General error
  /resource_req_params:
    post:
      operationId: postResourceReqParams
      requestBody:
        required: true
        content:
          'application/json':
            schema:
              $ref: '#/components/schemas/ReqBody'
      responses:
        '200':
          description: Created
        '500':
          description: General error
components:
  schemas:
    fieldUUID:
      type: object
      properties:
        uuid:
          $ref: '#/components/schemas/UUID'
      required:
        - uuid
    UUID:
      type: string
      pattern: '^[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}$'
    ArrayInput:
      type: array
      minItems: 1
      items:
        $ref: "#/components/schemas/ArrayItem"
    ArrayItem:
      type: object
      properties:
        id:
          $ref: "#/components/schemas/UUID"
      required:
        - id
    ReqBody:
      type: object
      properties:
        field1:
          $ref: "#/components/schemas/UUID"
        field2:
          type: "string"
      required:
        - field1
        - field2
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
`

type vErrorPayload struct {
	Error vError `json:"error"`
}

type vError struct {
	Errors  *map[string]interface{} `json:"errors,omitempty"`
	Message string                  `json:"message"`
	Type    string                  `json:"type"`
}

func TestOapiRequestValidator(t *testing.T) {
	t.Parallel()

	var (
		rByte []byte
		resp  *http.Response
	)

	swagger, err := openapi3.NewLoader().LoadFromData([]byte(testSchema))
	require.NoError(t, err, "Error initializing swagger")

	app := fiber.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := middlewareOapi.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
				// The fiber context should be propagated into here.
				eCtx := middlewareOapi.GetFiberContext(c)
				assert.NotNil(t, eCtx)
				// As should user data
				assert.EqualValues(t, "hi!", middlewareOapi.GetUserData(c))

				for _, s := range input.Scopes {
					if s == "someScope" {
						return nil
					}
					if s == "unauthorized" {
						return fiber.ErrUnauthorized
					}
				}
				return fiber.ErrForbidden
			},
		},
		UserData: "hi!",
	}

	swagger.Servers = nil
	app.Use(middlewareOapi.OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	app.Get("/resource", func(c *fiber.Ctx) error {
		called = true
		return nil
	})

	t.Run("Let's send a good request, it should pass", func(t *testing.T) {
		resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/resource", nil))
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.True(t, called, "Handler should have been called")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "", resp.Header.Get(fiber.HeaderAllow))
		called = false
	})

	t.Run("Send an out-of-spec parameter", func(t *testing.T) {
		resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/resource?id=500", nil))
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	})

	t.Run("Send a bad parameter type", func(t *testing.T) {
		resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/resource?id=foo", nil))
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	})

	// Add a handler for the POST message
	app.Post("/resource", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	called = false

	//nolint:dupl
	t.Run("Send a good request body", func(t *testing.T) {
		body := struct {
			Name string `json:"name"`
		}{
			Name: "Marcin",
		}
		bodyByte, _ := json.Marshal(body)
		req := httptest.NewRequest(fiber.MethodPost, "/resource", bytes.NewBuffer(bodyByte))
		req.Header.Add("Content-Type", "application/json")
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	})

	t.Run("Send a malformed body", func(t *testing.T) {
		body := struct {
			Name int `json:"name"`
		}{
			Name: 7,
		}
		bodyByte, _ := json.Marshal(body)
		resp, err = app.Test(httptest.NewRequest(fiber.MethodPost, "/resource", bytes.NewBuffer(bodyByte)))
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	})

	app.Get("/protected_resource", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	t.Run("Call a protected function to which we have access", func(t *testing.T) {
		resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/protected_resource", nil))
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	})

	app.Get("/protected_resource2", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	t.Run("Call a protected function to which we dont have access", func(t *testing.T) {
		resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/protected_resource2", nil))
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	})

	app.Get("/protected_resource_401", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	t.Run("Call a protected function without credentials", func(t *testing.T) {
		resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/protected_resource_401", nil))
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	})

	called = false

	// Add a handler for the POST message
	app.Post("/resource_uuid", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	//nolint:dupl
	t.Run("Send a good request body resource_uuid", func(t *testing.T) {
		body := struct {
			UUID string `json:"uuid"`
		}{
			UUID: "b3ec741f-04df-4135-a450-961ef7c569cf",
		}
		bodyByte, _ := json.Marshal(body)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_uuid", bytes.NewBuffer(bodyByte))
		req.Header.Add("Content-Type", "application/json")
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	})

	called = false

	// Add a handler for the POST message
	app.Post("/resource_array", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusOK)
	})

	t.Run("Send a good request body resource_array", func(t *testing.T) {
		body := []struct {
			ID string `json:"id"`
		}{
			{ID: "b3ec741f-04df-4135-a450-961ef7c569cf"},
			{ID: "b3ec741f-04df-4135-a450-961ef7c569cc"},
		}
		bodyByte, _ := json.Marshal(body)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_array", bytes.NewBuffer(bodyByte))
		req.Header.Add("Content-Type", "application/json")
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	})

	t.Run("Validate request body resource_array", func(t *testing.T) {
		body := make([]string, 0)
		bodyByte, _ := json.Marshal(body)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_array", bytes.NewBuffer(bodyByte))
		req.Header.Add("Content-Type", "application/json")
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		rByte, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		respErr := &vErrorPayload{}
		err = json.Unmarshal(rByte, respErr)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.NotNil(t, respErr.Error.Errors)
		mapErrors := *respErr.Error.Errors
		be, ok := mapErrors["body"]
		assert.True(t, ok)
		msg, ok := be.(map[string]interface{})["message"]
		assert.True(t, ok)
		assert.Equal(t, "minimum number of items is 1", fmt.Sprintf("%s", msg))
	})

	t.Run("Not validate one param in request body resource_array", func(t *testing.T) {
		body := []struct {
			ID string `json:"id"`
		}{
			{ID: "b3ec741f-04df-4135-a450-961ef7c569cf"},
			{ID: "b3ec741f-04df-4135-a450-961ef7c569ccb"},
		}
		bodyByte, _ := json.Marshal(body)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_array", bytes.NewBuffer(bodyByte))
		req.Header.Add("Content-Type", "application/json")
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		rByte, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		respErr := &vErrorPayload{}
		err = json.Unmarshal(rByte, respErr)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.NotNil(t, respErr.Error.Errors)
		mapErrors := *respErr.Error.Errors
		be, ok := mapErrors["$[1].id"]
		assert.True(t, ok)
		msg, ok := be.(map[string]interface{})["message"]
		assert.True(t, ok)
		//nolint:lll
		assert.Equal(t,
			"string \"b3ec741f-04df-4135-a450-961ef7c569ccb\" doesn't match the regular expression \"^[0-9a-f]{8}\\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\\b[0-9a-f]{12}$\"",
			fmt.Sprintf("%s", msg))
	})

	app = fiber.New()

	options = middlewareOapi.Options{
		Options: openapi3filter.Options{
			MultiError: true,
		},
	}

	swagger.Servers = nil
	app.Use(middlewareOapi.OapiRequestValidatorWithOptions(swagger, &options))

	called = false

	// Add a handler for the POST message
	app.Post("/resource_req_params", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	t.Run("Send a good request body resource_req_params", func(t *testing.T) {
		body := struct {
			Field1 string `json:"field1"`
			Field2 string `json:"field2"`
		}{
			Field1: "b3ec741f-04df-4135-a450-961ef7c569cf",
			Field2: "b4ec741f-04df-4135-a450-961ef7c569cf",
		}
		bodyByte, _ := json.Marshal(body)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_req_params", bytes.NewBuffer(bodyByte))
		req.Header.Add("Content-Type", "application/json")
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	})

	t.Run("Not validate params in request body resource_req_params", func(t *testing.T) {
		body := struct {
			Field1 string `json:"field1"`
		}{}
		bodyByte, _ := json.Marshal(body)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_req_params", bytes.NewBuffer(bodyByte))
		req.Header.Add("Content-Type", "application/json")
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		rByte, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		respErr := &vErrorPayload{}
		err = json.Unmarshal(rByte, respErr)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.NotNil(t, respErr.Error.Errors)
		mapErrors := *respErr.Error.Errors
		assert.Equal(t, 2, len(mapErrors))
		be, ok := mapErrors["$.field1"]
		assert.True(t, ok)
		msg, ok := be.(map[string]interface{})["message"]
		assert.True(t, ok)
		assert.Equal(t,
			"string \"\" doesn't match the regular expression \"^[0-9a-f]{8}\\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\\b[0-9a-f]{12}$\"",
			fmt.Sprintf("%s", msg))

		be, ok = mapErrors["$.field2"]
		assert.True(t, ok)
		msg, ok = be.(map[string]interface{})["message"]
		assert.True(t, ok)
		assert.Equal(t, "property \"field2\" is missing", fmt.Sprintf("%s", msg))
	})
}
