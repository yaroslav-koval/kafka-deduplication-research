package tracing

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	serviceName = "trace-demo"
	environment = "production"
)

type JaegerConfig struct {
	UseAgent                     bool
	AgentHost                    string
	AgentPort                    string
	AgentReconnectingIntervalSec int64
	URL                          string
}

type JaegerProvider struct {
	cfg JaegerConfig
	tp  *trace.TracerProvider
}

type JaegerOptions struct {
	Environment string
	ServiceName string
	Attrs       map[string]string
	Expr        *jaeger.Exporter
}

func NewProviderJaeger(cfg JaegerConfig) *JaegerProvider {
	pj := &JaegerProvider{
		cfg: cfg,
	}
	pj.Use(nil)

	return pj
}

func (pj *JaegerProvider) Use(opts *JaegerOptions) {
	var (
		exp *jaeger.Exporter
		err error
	)

	attrs := make([]attribute.KeyValue, 0)
	srvName := serviceName
	envName := environment

	if opts != nil {
		srvName = opts.ServiceName
		envName = opts.Environment

		for key, val := range opts.Attrs {
			attrs = append(attrs, attribute.String(key, val))
		}

		// init custom exporter
		if opts.Expr != nil {
			exp = opts.Expr
		}
	}

	if exp == nil {
		if pj.cfg.UseAgent {
			exp, err = pj.expAgentEndpoint()
		} else {
			exp, err = pj.exprCollectorEndpoint()
		}

		if err != nil {
			_ = cerror.NewF(
				context.Background(),
				cerror.KindInternal,
				err.Error(),
				"providerJaeger::init exporter jaeger").LogError()
		}
	}

	attrs = append(attrs,
		semconv.ServiceNameKey.String(srvName),
		attribute.String("environment", envName))

	pj.tp = trace.NewTracerProvider(
		// Always be sure to batch in production.
		trace.WithBatcher(exp),
		// Record information about this application in a Resource.
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			attrs...,
		)),
	)
}

func (pj *JaegerProvider) HasSdkTracerProvider() bool {
	return pj.tp != nil
}

func (pj *JaegerProvider) GetSdkTracerProvider() *trace.TracerProvider {
	return pj.tp
}

func (pj *JaegerProvider) exprCollectorEndpoint() (*jaeger.Exporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(pj.cfg.URL)))
}

func (pj *JaegerProvider) expAgentEndpoint() (*jaeger.Exporter, error) {
	rInterval := time.Duration(pj.cfg.AgentReconnectingIntervalSec) * time.Second

	return jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost(pj.cfg.AgentHost),
		jaeger.WithAgentPort(pj.cfg.AgentPort),
		jaeger.WithAttemptReconnectingInterval(rInterval)))
}
