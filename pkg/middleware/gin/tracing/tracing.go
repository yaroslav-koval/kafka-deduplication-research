package tracing

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/tracing"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	DefaultParentSpanKey = "#defaultTracingParentSpanKey"
	DefaultComponentName = "gin-gonic"
	TraceKey             = "__otel_trace_key"
)

var (
	// DefaultTraceConfig is the default Trace middleware config.
	defaultConfig = Config{
		Modify: func(ctx *gin.Context, attrs map[string]string) map[string]string {
			attrs["http.method"] = ctx.Request.Method
			attrs["http.route"] = ctx.Request.RequestURI
			attrs["http.remote_addr"] = ctx.RemoteIP()
			attrs["http.path"] = ctx.FullPath()
			attrs["http.host"] = ctx.Request.Host

			return attrs
		},
		OperationName: func(ctx *gin.Context) string {
			return "HTTP " + ctx.Request.Method + " URL: " + ctx.FullPath()
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
	// Default: func(ctx *gin.Context) string {
	//	 return "HTTP " + ctx.Method() + " URL: " + ctx.Path()
	// }
	OperationName func(*gin.Context) string

	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*gin.Context) bool

	// Modify
	Modify func(*gin.Context, map[string]string) map[string]string
}

// New returns a Trace middleware.
// Trace middleware traces http requests and reporting errors.
func New(tracer tracing.Tracer) gin.HandlerFunc {
	c := defaultConfig
	c.Tracer = tracer

	return NewWithConfig(c)
}

// NewWithConfig returns a Trace middleware with config.
func NewWithConfig(config ...Config) gin.HandlerFunc {
	var cfg *Config
	if len(config) > 0 {
		cfg = &config[0]
	}

	cfg = mergeDefaultCfg(cfg)

	return func(ctx *gin.Context) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(ctx) {
			ctx.Next()
			return
		}

		operationName := cfg.OperationName(ctx)
		tt := cfg.Tracer.GetTrace()
		nCtx, span := tt.Start(cfg.ctx, operationName)

		attrs := make(map[string]string)
		attrs = cfg.Modify(ctx, attrs)

		var err error
		defer func() {
			status := ctx.Writer.Status()

			if err != nil {
				attrs["error.message"] = err.Error()

				if status == http.StatusOK {
					// this is ugly workaround for cases when httpError.code == 0 or error was not httpError and status
					// in request was 200 (OK). In these cases replace status with something that represents an error
					// it could be that error handlers or middlewares up in chain will output different status code to
					// client. but at least we send something better than 200 to jaeger
					status = http.StatusInternalServerError
				}

				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}

			attrs["http.status_code"] = fmt.Sprintf("%d", uint16(status))

			if reqID := ctx.Value(consts.HeaderXRequestID); reqID != nil {
				attrs["http.requestID"] = fmt.Sprintf("%s", reqID)
			}

			for key, val := range attrs {
				span.SetAttributes(attribute.Key(key).String(val))
			}

			span.End()
		}()

		ctx.Set(TraceKey, nCtx)

		ctx.Next()

		err = ctx.Err()
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
