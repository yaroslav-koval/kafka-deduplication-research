package tracing

import (
	"context"
	"errors"
	"fmt"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/tracing"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	DefaultParentSpanKey = "#defaultTracingParentSpanKey"
	DefaultComponentName = "fiber/v2"
	TraceKey             = "__otel_trace_key"
)

var (
	// DefaultTraceConfig is the default Trace middleware config.
	defaultConfig = Config{
		Modify: func(ctx *fiber.Ctx, attrs map[string]string) map[string]string {
			attrs["http.method"] = ctx.Method()
			attrs["http.route"] = ctx.OriginalURL()
			attrs["http.remote_addr"] = ctx.IP()
			attrs["http.path"] = ctx.Path()
			attrs["http.host"] = ctx.Hostname()

			return attrs
		},
		OperationName: func(ctx *fiber.Ctx) string {
			return "HTTP " + ctx.Method() + " URL: " + ctx.Path()
		},
		ComponentName: DefaultComponentName,
		ParentSpanKey: DefaultParentSpanKey,
	}
)

type Config struct {
	ctx context.Context
	// Tracer
	// Default: NoopTracer
	Tracer tracing.Tracer

	// ParentSpanKey
	// Default: #defaultTracingParentSpanKey
	ParentSpanKey string

	// ComponentName used for describing the tracing component name
	ComponentName string

	// OperationName
	// Default: func(ctx *fiber.Ctx) string {
	//	 return "HTTP " + ctx.Method() + " URL: " + ctx.Path()
	// }
	OperationName func(*fiber.Ctx) string

	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// Modify
	Modify func(*fiber.Ctx, map[string]string) map[string]string
}

// New returns a Trace middleware.
// Trace middleware traces http requests and reporting errors.
func New(tracer tracing.Tracer) fiber.Handler {
	c := defaultConfig
	c.Tracer = tracer

	return NewWithConfig(c)
}

// NewWithConfig returns a Trace middleware with config.
//
//nolint:nestif
func NewWithConfig(config ...Config) fiber.Handler {
	var cfg *Config
	if len(config) > 0 {
		cfg = &config[0]
	}

	cfg = mergeDefaultCfg(cfg)

	return func(ctx *fiber.Ctx) error {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(ctx) {
			return ctx.Next()
		}

		operationName := cfg.OperationName(ctx)
		tt := cfg.Tracer.GetTrace()
		nCtx, span := tt.Start(cfg.ctx, operationName)

		attrs := make(map[string]string)
		attrs = cfg.Modify(ctx, attrs)

		var err error
		defer func() {
			status := ctx.Response().StatusCode()

			if err != nil {
				var httpError *fiber.Error

				if errors.As(err, &httpError) {
					if httpError.Code != 0 {
						status = httpError.Code
					}

					attrs["error.message"] = httpError.Message
				} else {
					attrs["error.message"] = err.Error()
				}

				if status == fiber.StatusOK {
					// this is ugly workaround for cases when httpError.code == 0 or error was not httpError and status
					// in request was 200 (OK). In these cases replace status with something that represents an error
					// it could be that error handlers or middlewares up in chain will output different status code to
					// client. but at least we send something better than 200 to jaeger
					status = fiber.StatusInternalServerError
				}

				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}

			attrs["http.status_code"] = fmt.Sprintf("%d", uint16(status))

			if reqID := ctx.Context().UserValue(consts.HeaderXRequestID); reqID != nil {
				attrs["http.requestID"] = fmt.Sprintf("%s", reqID)
			}

			for key, val := range attrs {
				span.SetAttributes(attribute.Key(key).String(val))
			}

			span.End()
		}()

		ctx.Locals(TraceKey, nCtx)

		err = ctx.Next()

		return err
	}
}

func mergeDefaultCfg(cfg *Config) *Config {
	cfg.ctx = context.Background()

	if !cfg.Tracer.HasTrace() {
		cfg.Tracer.Trace(DefaultComponentName)
	}

	if cfg.ParentSpanKey == "" {
		cfg.ParentSpanKey = defaultConfig.ParentSpanKey
	}

	if cfg.ComponentName == "" {
		cfg.ComponentName = defaultConfig.ComponentName
	}

	if cfg.Modify == nil {
		cfg.Modify = defaultConfig.Modify
	}

	if cfg.OperationName == nil {
		cfg.OperationName = defaultConfig.OperationName
	}

	return cfg
}
