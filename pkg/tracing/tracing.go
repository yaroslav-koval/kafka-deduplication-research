package tracing

import (
	"context"
	"errors"
	"kafka-polygon/pkg/cerror"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracingTimeout     = 5 * time.Second
	errNoEmptyProvider = errors.New("not set provider")
	errNoEmptyTracer   = errors.New("not set trace.Tracer")
)

type Tracer interface {
	GetContext() context.Context
	HasTrace() bool
	GetTrace() trace.Tracer
	Trace(compName string)
	Span(ctx context.Context, spanName string, attrs map[string]string)
	SpanFiber(ctx *fiber.Ctx, spanName string, attrs map[string]string)
	SpanError(ctx context.Context, err error, spanName string, attrs map[string]string)
	SpanFiberError(ctx *fiber.Ctx, err error, spanName string, attrs map[string]string)
	Shutdown() error
}

type Provider interface {
	Use(opts *JaegerOptions)
	HasSdkTracerProvider() bool
	GetSdkTracerProvider() *tracesdk.TracerProvider
}

type Tracing struct {
	ctx  context.Context
	prov Provider
	tr   trace.Tracer
}

func New(prov Provider) *Tracing {
	if prov.HasSdkTracerProvider() {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	return &Tracing{
		prov: prov,
		ctx:  context.Background(),
		tr:   nil,
	}
}

func (t *Tracing) GetContext() context.Context {
	return t.ctx
}

func (t *Tracing) HasTrace() bool {
	return t.tr != nil
}

func (t *Tracing) GetTrace() trace.Tracer {
	return t.tr
}

func (t *Tracing) Trace(compName string) {
	if !t.prov.HasSdkTracerProvider() {
		_ = cerror.New(t.GetContext(), cerror.KindInternal, errNoEmptyProvider).LogError()
	}

	t.tr = t.prov.GetSdkTracerProvider().Tracer(compName)
}

func (t *Tracing) Span(ctx context.Context, spanName string, attrs map[string]string) {
	if t.tr == nil {
		_ = cerror.New(t.GetContext(), cerror.KindInternal, errNoEmptyTracer).LogError()
		return
	}

	nCtx, span := t.tr.Start(ctx, spanName)
	defer span.End()

	for key, val := range attrs {
		span.SetAttributes(attribute.Key(key).String(val))
	}

	t.ctx = nCtx
}

func (t *Tracing) SpanFiber(ctx *fiber.Ctx, spanName string, attrs map[string]string) {
	t.Span(ctx.UserContext(), spanName, attrs)
	ctx.SetUserContext(t.GetContext())
}

func (t *Tracing) SpanError(ctx context.Context, err error, spanName string, attrs map[string]string) {
	if t.tr == nil {
		_ = cerror.New(t.GetContext(), cerror.KindInternal, errNoEmptyTracer).LogError()
		return
	}

	nCtx, span := t.tr.Start(ctx, spanName)
	defer span.End()

	for key, val := range attrs {
		span.SetAttributes(attribute.Key(key).String(val))
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	t.ctx = nCtx
}

func (t *Tracing) SpanFiberError(ctx *fiber.Ctx, err error, spanName string, attrs map[string]string) {
	t.SpanError(ctx.UserContext(), err, spanName, attrs)
	ctx.SetUserContext(t.GetContext())
}

func (t *Tracing) Shutdown() error {
	if !t.prov.HasSdkTracerProvider() {
		return errNoEmptyProvider
	}

	// Do not make the application hang when it is shutdown.
	ctx, cancel := context.WithTimeout(t.ctx, tracingTimeout)
	defer cancel()

	if err := t.prov.GetSdkTracerProvider().Shutdown(ctx); err != nil {
		_ = cerror.New(t.GetContext(), cerror.KindInternal, err).LogError()
		return err
	}

	return nil
}
