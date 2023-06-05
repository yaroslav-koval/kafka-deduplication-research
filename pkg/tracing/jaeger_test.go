package tracing_test

import (
	"context"
	"fmt"
	pkgtracing "kafka-polygon/pkg/tracing"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func newUDPListener() (net.PacketConn, error) {
	return net.ListenPacket("udp", "127.0.0.1:0")
}

func TestJaegerProvider(t *testing.T) {
	t.Parallel()

	mockServer, err := newUDPListener()
	require.NoError(t, err)

	defer mockServer.Close()

	assert.True(t, len(handler.GetErrs()) == 0)

	host, port, err := net.SplitHostPort(mockServer.LocalAddr().String())
	require.NoError(t, err)

	// 1500 spans, size 79559, does not fit within one UDP packet with the default size of 65000.
	n := 1500
	s := make(tracetest.SpanStubs, n).Snapshots()

	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(jaeger.WithAgentHost(host), jaeger.WithAgentPort(port)),
	)
	require.NoError(t, err)

	pj := pkgtracing.NewProviderJaeger(pkgtracing.JaegerConfig{UseAgent: false, URL: ""})
	pj.Use(&pkgtracing.JaegerOptions{
		Environment: "test",
		ServiceName: "test-provider",
		Attrs: map[string]string{
			"test-attr-name": "test-attr-value",
		},
		Expr: exp,
	})

	ctx := context.Background()
	assert.NoError(t, exp.ExportSpans(ctx, s))
	assert.Equal(t, true, pj.HasSdkTracerProvider())

	tracer := pj.GetSdkTracerProvider().Tracer("test-tracer")

	for i := 0; i < 3; i++ {
		_, span := tracer.Start(ctx, fmt.Sprintf("test-span-%d", i))
		span.End()
		assert.True(t, span.SpanContext().IsValid())
	}

	require.NoError(t, pj.GetSdkTracerProvider().Shutdown(ctx))
}

func TestJaegerProviderError(t *testing.T) {
	pj := pkgtracing.JaegerProvider{}
	assert.Equal(t, false, pj.HasSdkTracerProvider())
	assert.Nil(t, pj.GetSdkTracerProvider())
}

func generateALargeSpan() tracetest.SpanStub {
	return tracetest.SpanStub{
		Name: "a-longer-name-that-makes-it-exceeds-limit",
	}
}

func TestJaegerProviderSpanExceedsMaxPacketLimit(t *testing.T) {
	t.Parallel()

	mockServer, err := newUDPListener()
	require.NoError(t, err)

	defer mockServer.Close()

	host, port, err := net.SplitHostPort(mockServer.LocalAddr().String())
	assert.NoError(t, err)

	// 106 is the serialized size of a span with default values.
	maxSize := 106

	largeSpans := tracetest.SpanStubs{generateALargeSpan(), {}}.Snapshots()
	normalSpans := tracetest.SpanStubs{{}, {}}.Snapshots()

	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(jaeger.WithAgentHost(host), jaeger.WithAgentPort(port), jaeger.WithMaxPacketSize(maxSize+1)),
	)
	require.NoError(t, err)

	ctx := context.Background()
	assert.Error(t, exp.ExportSpans(ctx, largeSpans))
	assert.NoError(t, exp.ExportSpans(ctx, normalSpans))
	assert.NoError(t, exp.Shutdown(ctx))
}

func TestJaegerProviderEmitBatchWithMultipleErrors(t *testing.T) {
	t.Parallel()

	mockServer, err := newUDPListener()
	require.NoError(t, err)

	defer mockServer.Close()

	host, port, err := net.SplitHostPort(mockServer.LocalAddr().String())
	assert.NoError(t, err)

	span := generateALargeSpan()
	largeSpans := tracetest.SpanStubs{span, span}.Snapshots()
	// make max packet size smaller than span
	maxSize := len(span.Name)
	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(jaeger.WithAgentHost(host), jaeger.WithAgentPort(port), jaeger.WithMaxPacketSize(maxSize)),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = exp.ExportSpans(ctx, largeSpans)
	assert.Error(t, err)
	require.Contains(t, err.Error(), "multiple errors")
}
