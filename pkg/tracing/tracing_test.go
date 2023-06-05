package tracing_test

import (
	"context"
	"errors"
	"fmt"
	pkgtracing "kafka-polygon/pkg/tracing"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"go.opentelemetry.io/otel"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type MockedProvider struct {
	mock.Mock
	tp *tracesdk.TracerProvider
}

func (m *MockedProvider) Use(opts *pkgtracing.JaegerOptions) {
	m.Called(opts)
}

func (m *MockedProvider) HasSdkTracerProvider() bool {
	return true
}

func (m *MockedProvider) GetSdkTracerProvider() *tracesdk.TracerProvider {
	return m.tp
}

type basicSpanProcesor struct {
	running             bool
	injectShutdownError error
}

func (t *basicSpanProcesor) Shutdown(context.Context) error {
	t.running = false
	return t.injectShutdownError
}

func (t *basicSpanProcesor) OnStart(context.Context, tracesdk.ReadWriteSpan) {}
func (t *basicSpanProcesor) OnEnd(tracesdk.ReadOnlySpan)                     {}
func (t *basicSpanProcesor) ForceFlush(context.Context) error {
	return nil
}

type storingHandler struct {
	errs []error
	m    sync.RWMutex
}

func (s *storingHandler) Handle(err error) {
	s.m.Lock()
	defer s.m.Unlock()

	s.errs = append(s.errs, err)
}

func (s *storingHandler) Reset() {
	s.m.Lock()
	defer s.m.Unlock()
	s.errs = []error{}
}

func (s *storingHandler) GetErrs() []error {
	s.m.Lock()
	defer s.m.Unlock()

	return s.errs
}

var (
	tid trace.TraceID
	sid trace.SpanID
	sc  trace.SpanContext

	handler storingHandler
)

func TestMain(m *testing.M) {
	initHandler()

	// Run the other tests
	os.Exit(m.Run())
}

func TestTracer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stp := tracesdk.NewTracerProvider()
	sp := &basicSpanProcesor{}
	stp.RegisterSpanProcessor(sp)

	sp.running = true

	mockProvider := &MockedProvider{
		tp: stp,
	}

	tt := pkgtracing.New(mockProvider)
	tt.Trace("tracername")

	// Verify that the SchemaURL of the constructed Tracer is correctly populated.
	tracerS := tt.GetTrace()

	for i := 0; i < 3; i++ {
		nameSpan := fmt.Sprintf("test-span-%d", i)
		_, span := tracerS.Start(ctx, nameSpan)
		span.End()

		got := span.IsRecording()
		assert.Equal(t, got, false, nameSpan)
		assert.True(t, span.SpanContext().IsValid())
	}

	require.NoError(t, tt.Shutdown())
}

func TestTracerFailedProcessorShutdown(t *testing.T) {
	t.Parallel()

	defer handler.Reset()

	stp := tracesdk.NewTracerProvider()
	spErr := errors.New("basic span processor shutdown failure")
	sp := &basicSpanProcesor{
		running:             true,
		injectShutdownError: spErr,
	}
	stp.RegisterSpanProcessor(sp)

	mockProvider := &MockedProvider{
		tp: stp,
	}

	tt := pkgtracing.New(mockProvider)

	err := tt.Shutdown()
	assert.Error(t, err)
	assert.Equal(t, err, spErr)
}

func TestTracerFailedProcessorShutdownInUnregister(t *testing.T) {
	t.Parallel()

	defer handler.Reset()

	stp := tracesdk.NewTracerProvider()
	spErr := errors.New("basic span processor shutdown failure")
	sp := &basicSpanProcesor{
		running:             true,
		injectShutdownError: spErr,
	}
	stp.RegisterSpanProcessor(sp)
	stp.UnregisterSpanProcessor(sp)

	assert.Contains(t, handler.GetErrs(), spErr)

	err := stp.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestSpan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stp := tracesdk.NewTracerProvider()

	attr := map[string]string{"attr-key": "attr-val"}

	sp := &basicSpanProcesor{}
	stp.RegisterSpanProcessor(sp)

	sp.running = true

	mockProvider := &MockedProvider{
		tp: stp,
	}

	tt := pkgtracing.New(mockProvider)
	tt.Trace("tracername-with-span")
	tt.Span(ctx, "test-span", attr)

	spanCtx := trace.SpanContextFromContext(tt.GetContext())

	assert.Equal(t, true, spanCtx.HasSpanID())
	assert.Equal(t, true, spanCtx.IsValid())

	require.NoError(t, tt.Shutdown())
}

func TestSpanError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stp := tracesdk.NewTracerProvider()
	attr := map[string]string{"attr-key-err": "attr-val-err"}
	expErr := errors.New("test span error")

	sp := &basicSpanProcesor{}
	stp.RegisterSpanProcessor(sp)

	sp.running = true

	mockProvider := &MockedProvider{
		tp: stp,
	}

	tt := pkgtracing.New(mockProvider)
	tt.Trace("tracername-with-span-err")
	tt.SpanError(ctx, expErr, "test-span-err", attr)

	spanCtx := trace.SpanContextFromContext(tt.GetContext())

	assert.Equal(t, true, spanCtx.HasSpanID())
	assert.Equal(t, true, spanCtx.IsValid())

	require.NoError(t, tt.Shutdown())
}

func initHandler() {
	tid, _ = trace.TraceIDFromHex("01020304050607080102040810203040")
	sid, _ = trace.SpanIDFromHex("0102040810203040")
	sc = trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: 0x1,
	})

	otel.SetErrorHandler(&handler)
}
