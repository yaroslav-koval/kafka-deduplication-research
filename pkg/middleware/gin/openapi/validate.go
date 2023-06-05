package openapi

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/middleware"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
)

const (
	UserDataKey   = "oapi-codegen/user-data"
	GinContextKey = "oapi-codegen/gin-context"
)

// Create validator middleware from a YAML file path
func OapiValidatorFromYamlFile(path string) (gin.HandlerFunc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", path, err)
	}

	swagger, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s as Swagger YAML: %s",
			path, err)
	}

	return OapiRequestValidator(swagger), nil
}

// This is an gin middleware function which validates incoming HTTP requests
// to make sure that they conform to the given OAPI 3.0 specification. When
// OAPI validation fails on the request, we return an HTTP/400 with error message
func OapiRequestValidator(swagger *openapi3.T) gin.HandlerFunc {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

// ErrorHandler is called when there is an error in validation
type ErrorHandler func(c *gin.Context, statusCode int, err error)

// Options to customize request validation. These are passed through to
// openapi3filter.
type Options struct {
	ErrorHandler ErrorHandler
	Options      openapi3filter.Options
	ParamDecoder openapi3filter.ContentParameterDecoder
	UserData     interface{}
	Skipper      *bool
	SkipperGroup []string
}

// Create a validator from a swagger object, with validation options
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) gin.HandlerFunc {
	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		return func(ctx *gin.Context) {
			cErr := cerror.New(ctx, cerror.KindInternal, err).LogError()
			ctx.AbortWithStatusJSON(cerror.ErrKind(cErr).HTTPCode(), cerror.BuildErrorResponse(cErr))
		}
	}

	skipper := getSkipperFromOptions(options)

	return func(c *gin.Context) {
		if skipper || checkSkipperGroup(c, options.SkipperGroup) {
			c.Next()
			return
		}

		err := ValidateRequestFromContext(c, router, options)
		if err != nil {
			if options != nil && options.ErrorHandler != nil {
				options.ErrorHandler(c, cerror.ErrKind(err).HTTPCode(), err)
				// in case the handler didn't internally call Abort, stop the chain
				c.Abort()
			} else {
				c.AbortWithStatusJSON(cerror.ErrKind(err).HTTPCode(), cerror.BuildErrorResponse(err))
			}
		}

		c.Next()
	}
}

// ValidateRequestFromContext is called from the middleware above and actually does the work
// of validating a request.
func ValidateRequestFromContext(c *gin.Context, router routers.Router, options *Options) error {
	req := c.Request
	route, pathParams, err := router.FindRoute(req)

	// We failed to find a matching route for the request.
	if err != nil {
		switch e := err.(type) {
		case *routers.RouteError:
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return cerror.NewF(
				c,
				cerror.KindBadParams,
				"%s", e.Reason).LogError()
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return cerror.NewF(
				c,
				cerror.KindInternal,
				"error validating route: %s", err.Error()).LogError()
		}
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:     req,
		PathParams:  pathParams,
		Route:       route,
		QueryParams: c.Request.URL.Query(),
	}

	// Pass the gin context into the request validator, so that any callbacks
	// which it invokes make it available.
	//
	//nolint:staticcheck,revive
	requestContext := context.WithValue(context.Background(), GinContextKey, c)

	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData) //nolint:staticcheck,revive
	}

	err = openapi3filter.ValidateRequest(requestContext, validationInput)
	if err != nil {
		switch e := err.(type) {
		case *openapi3filter.RequestError:
			// We've got a bad request
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			return middleware.CValidateRequestError(c, e)
		case openapi3.MultiError:
			return middleware.CValidateMultiError(c, e, nil)
		case *openapi3filter.SecurityRequirementsError:
			return cerror.NewF(c, cerror.KindForbidden, "%s %s", e.Error(), err).LogError()
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return cerror.NewF(
				c,
				cerror.KindInternal,
				"error validating request: %s", err).LogError()
		}
	}

	return nil
}

// Helper function to get the gin context from within requests. It returns
// nil if not found or wrong type.
func GetGinContext(c context.Context) *gin.Context {
	iface := c.Value(GinContextKey)

	if iface == nil {
		return nil
	}

	ginCtx, ok := iface.(*gin.Context)

	if !ok {
		return nil
	}

	return ginCtx
}

func GetUserData(c context.Context) interface{} {
	return c.Value(UserDataKey)
}

func checkSkipperGroup(c *gin.Context, gr []string) bool {
	req := c.Request
	for _, name := range gr {
		if strings.Contains(req.URL.String(), name) {
			return true
		}
	}

	return false
}

// attempt to get the skipper from the options whether it is set or not
func getSkipperFromOptions(options *Options) bool {
	if options == nil {
		return false
	}

	if options.Skipper == nil {
		return false
	}

	return *options.Skipper
}
