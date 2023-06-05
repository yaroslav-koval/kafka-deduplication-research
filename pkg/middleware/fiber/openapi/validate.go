package openapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/middleware"
	"net/http"
	"net/url"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gofiber/fiber/v2"
)

type ctxKey string

const (
	FiberContextKey ctxKey = "oapi-codegen/fiber-context"
	UserDataKey     ctxKey = "oapi-codegen/user-data"
)

func OapiValidatorFromYamlFile(path string) (fiber.Handler, error) {
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

// OapiRequestValidator create a validator from a swagger object.
func OapiRequestValidator(swagger *openapi3.T) fiber.Handler {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

// Options to customize request validation. These are passed through to
// openapi3filter.
type Options struct {
	Options      openapi3filter.Options
	ParamDecoder openapi3filter.ContentParameterDecoder
	UserData     interface{}
	Skipper      *bool
}

// OapiRequestValidatorWithOptions a validator from a swagger object, with validation options
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) fiber.Handler {
	router, errRouter := legacyrouter.NewRouter(swagger)

	if errRouter != nil {
		return func(ctx *fiber.Ctx) error {
			cErr := cerror.New(ctx.Context(), cerror.KindInternal, errRouter).LogError()
			return ctx.Status(cerror.ErrKind(cErr).HTTPCode()).
				JSON(cerror.BuildErrorResponse(cErr))
		}
	}

	skipper := getSkipperFromOptions(options)

	return func(ctx *fiber.Ctx) error {
		if skipper {
			return ctx.Next()
		}

		err := ValidateRequestFromContext(ctx, router, options)
		if err != nil {
			ctx.Status(cerror.ErrKind(err).HTTPCode())

			if errBuild := ctx.JSON(cerror.BuildErrorResponse(err)); errBuild != nil {
				_ = ctx.Next()
				return errBuild
			}

			return nil
		}

		return ctx.Next()
	}
}

// ValidateRequestFromContext is called from the middleware above and actually does the work
// of validating a request.
//
//nolint:funlen
func ValidateRequestFromContext(ctx *fiber.Ctx, router routers.Router, options *Options) error {
	rC := &http.Request{}
	rC.MultipartForm, _ = ctx.MultipartForm()
	rC.Host = ctx.Hostname()
	rC.Method = ctx.Method()
	rC.URL, _ = url.Parse(ctx.OriginalURL())
	rC.Body = io.NopCloser(bytes.NewReader(ctx.Body()))
	rC.Header = http.Header{}

	ctx.Context().Request.Header.VisitAll(func(key, value []byte) {
		rC.Header.Add(string(key), string(value))
	})

	route, pathParams, err := router.FindRoute(rC)

	// We failed to find a matching route for the request.
	if err != nil {
		switch e := err.(type) {
		case *routers.RouteError:
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return cerror.NewF(
				ctx.Context(),
				cerror.KindBadParams,
				"%s", e.Reason).LogError()
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return cerror.NewF(
				ctx.Context(),
				cerror.KindInternal,
				"error validating route: %s", err.Error()).LogError()
		}
	}

	queryArgs := url.Values{}

	ctx.Context().QueryArgs().VisitAll(func(key, value []byte) {
		queryArgs.Add(string(key), string(value))
	})

	validationInput := &openapi3filter.RequestValidationInput{
		Request:     rC,
		PathParams:  pathParams,
		Route:       route,
		QueryParams: queryArgs,
	}

	// Pass the Fiber context into the request validator, so that interface{} callbacks
	// which it invokes make it available.
	requestContext := context.WithValue(context.Background(), FiberContextKey, ctx)

	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder

		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData)
	}

	err = openapi3filter.ValidateRequest(requestContext, validationInput)
	if err != nil {
		switch e := err.(type) {
		case *openapi3filter.RequestError:
			// We've got a bad request
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			return middleware.CValidateRequestError(ctx.Context(), e)
		case openapi3.MultiError:
			return middleware.CValidateMultiError(ctx.Context(), e, func(c context.Context, e error) error {
				secErr, ok := e.(*openapi3filter.SecurityRequirementsError)
				if ok {
					for _, se := range secErr.Errors {
						httpErr, sOk := se.(*fiber.Error)
						if sOk {
							return cerror.New(c, cerror.KindFromHTTPCode(httpErr.Code), httpErr).LogError()
						}
					}
				}

				return nil
			})
		case *openapi3filter.SecurityRequirementsError:
			for _, err := range e.Errors {
				httpErr, ok := err.(*fiber.Error)
				if ok {
					return cerror.New(
						ctx.Context(),
						cerror.KindFromHTTPCode(httpErr.Code),
						httpErr).LogError()
				}
			}

			return cerror.NewF(
				ctx.Context(),
				cerror.KindForbidden,
				"%s %s", e.Error(), err).LogError()
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return cerror.NewF(
				ctx.Context(),
				cerror.KindInternal,
				"error validating request: %s", err).LogError()
		}
	}

	return nil
}

// GetFiberContext Helper function to get the fiber context from within requests. It returns
// nil if not found or wrong type.
func GetFiberContext(c context.Context) *fiber.Ctx {
	iface := c.Value(FiberContextKey)
	if iface == nil {
		return nil
	}

	eCtx, ok := iface.(*fiber.Ctx)
	if !ok {
		return nil
	}

	return eCtx
}

func GetUserData(c context.Context) interface{} {
	return c.Value(UserDataKey)
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

func RegisterAdditionalBodyDecoders(d CustomBodyDecoder) {
	for contType, decodFn := range d {
		openapi3filter.RegisterBodyDecoder(contType, decodFn)
	}
}
