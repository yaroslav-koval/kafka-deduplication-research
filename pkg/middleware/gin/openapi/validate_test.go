package openapi_test

import (
	"context"
	_ "embed"
	"errors"
	middlewareOpenapi "kafka-polygon/pkg/middleware/gin/openapi"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

//go:embed test_spec.yaml
var testSchema []byte

func doGet(t *testing.T, handler http.Handler, rawURL string) *httptest.ResponseRecorder {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	response := testutil.NewRequest().Get(u.RequestURI()).WithHost(u.Host).WithAcceptJson().GoWithHTTPHandler(t, handler)

	return response.Recorder
}

func doPost(t *testing.T, handler http.Handler, rawURL string, jsonBody interface{}) *httptest.ResponseRecorder {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	response := testutil.NewRequest().Post(u.RequestURI()).WithHost(u.Host).WithJsonBody(jsonBody).GoWithHTTPHandler(t, handler)

	return response.Recorder
}

func TestOapiRequestValidator(t *testing.T) {
	t.Parallel()

	swagger, err := openapi3.NewLoader().LoadFromData(testSchema)
	require.NoError(t, err, "Error initializing swagger")

	// Create a new gin router
	g := gin.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := middlewareOpenapi.Options{
		ErrorHandler: func(c *gin.Context, statusCode int, err error) {
			c.String(statusCode, "test: "+err.Error())
		},
		Options: openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
				// The gin context should be propagated into here.
				gCtx := middlewareOpenapi.GetGinContext(c)
				assert.NotNil(t, gCtx)
				// As should user data
				assert.EqualValues(t, "hi!", middlewareOpenapi.GetUserData(c))

				for _, s := range input.Scopes {
					if s == "someScope" {
						return nil
					}
					if s == "unauthorized" {
						return errors.New("unauthorized")
					}
				}
				return errors.New("forbidden")
			},
		},
		UserData: "hi!",
	}

	// Install our OpenApi based request validator
	g.Use(middlewareOpenapi.OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	g.GET("/resource", func(c *gin.Context) {
		called = true
	})
	// Let's send the request to the wrong server, this should fail validation
	{
		rec := doGet(t, g, "http://not.deepmap.ai/resource")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, called, "Handler should not have been called")
	}

	// Let's send a good request, it should pass
	{
		rec := doGet(t, g, "http://deepmap.ai/resource")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Send an out-of-spec parameter
	{
		rec := doGet(t, g, "http://deepmap.ai/resource?id=500")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Send a bad parameter type
	{
		rec := doGet(t, g, "http://deepmap.ai/resource?id=foo")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Add a handler for the POST message
	g.POST("/resource", func(c *gin.Context) {
		called = true
		c.AbortWithStatus(http.StatusNoContent)
	})

	called = false
	// Send a good request body
	{
		body := struct {
			Name string `json:"name"`
		}{
			Name: "Marcin",
		}
		rec := doPost(t, g, "http://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Send a malformed body
	{
		body := struct {
			Name int `json:"name"`
		}{
			Name: 7,
		}
		rec := doPost(t, g, "http://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	g.GET("/protected_resource", func(c *gin.Context) {
		called = true
		c.AbortWithStatus(http.StatusNoContent)
	})

	// Call a protected function to which we have access
	{
		rec := doGet(t, g, "http://deepmap.ai/protected_resource")
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	g.GET("/protected_resource2", func(c *gin.Context) {
		called = true
		c.AbortWithStatus(http.StatusNoContent)
	})
	// Call a protected function to which we don't have access
	{
		rec := doGet(t, g, "http://deepmap.ai/protected_resource2")
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	g.GET("/protected_resource_401", func(c *gin.Context) {
		called = true
		c.AbortWithStatus(http.StatusNoContent)
	})
	// Call a protected function without credentials
	{
		rec := doGet(t, g, "http://deepmap.ai/protected_resource_401")
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Equal(t, "test: security requirements failed: unauthorized security requirements failed: unauthorized", rec.Body.String())
		assert.False(t, called, "Handler should not have been called")
		called = false
	}
}
