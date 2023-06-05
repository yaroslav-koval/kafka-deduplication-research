package kafka

import (
	"context"
	"errors"
	"fmt"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/tracing"
	"time"

	goKafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	BrokerKafkaProvider = "kafka"
	TraceKafkaProducer  = "_kafka_producer"
	TraceKafkaConsumer  = "_kafka_consumer"
)

var (
	errKafkaMessageEmpty = errors.New("kafka message empty")
)

type Provider struct {
	cfgCl   *Config
	enabled bool
	cl      KClient
	store   store.Store
	trace   tracing.Tracer
}

func NewKafkaProvider(cfg *Config) *Provider {
	cl := NewClient(cfg)

	return &Provider{
		cfgCl:   cfg,
		enabled: true,
		cl:      cl,
	}
}

func (p *Provider) SetClient(cl KClient) {
	p.cl = cl
}

func (p *Provider) SetEnabled(enable bool) {
	p.enabled = enable
}

func (p *Provider) SetStore(s store.Store) {
	p.store = s
}

func (p *Provider) SetTracing(t tracing.Tracer) {
	p.trace = t
}

func (p *Provider) GetIsTopicExists(ctx context.Context, topic string) (bool, error) {
	return p.cl.GetIsTopicExists(ctx, topic)
}

func (p *Provider) Publish(ctx context.Context, topic string, e event.BaseEvent) error {
	err := p.cl.SendMessage(ctx, topic, e)

	p.syncTrace(ctx, TraceKafkaProducer, topic, e, err)

	return err
}

func (p *Provider) Sync(ctx context.Context, topic string, fn provider.HandlerFn) {
	if p.enabled {
		mHandler := p.getHandler(ctx, topic, fn)
		errCh := p.cl.ListenTopic(ctx, topic, mHandler)

		go p.processListenerErrors(ctx, topic, mHandler, errCh)
	}
}

func (p *Provider) Stop() {
	if p.enabled {
		p.cl.Stop()

		if p.trace != nil {
			_ = p.trace.Shutdown()
		}
	}
}

func (p *Provider) GetType() string {
	return BrokerKafkaProvider
}

func (p *Provider) getHandler(_ context.Context, topic string, fn provider.HandlerFn) HandelFn {
	return func(ctx context.Context, m *goKafka.Message) (event.BaseEvent, error) {
		if m == nil {
			return nil, cerror.New(ctx, cerror.KindKafkaOther, errKafkaMessageEmpty).LogError()
		}

		ph := provider.NewHandlerProcessing(p.store)

		em := event.Message{
			Key:   converto.BytePointer(m.Key),
			Value: m.Value,
		}

		e, err := ph.Run(ctx, fn, em)

		p.syncTrace(ctx, TraceKafkaConsumer, topic, e, err)

		return e, err
	}
}

func (p *Provider) tryRerunTopicListener(ctx context.Context, topic string, mHandler MessageHandler) {
	time.Sleep(p.cfgCl.RerunDelay)
	log.DebugF(ctx, "try re-run consumer by topic = %v", topic)
	p.cl.ListenTopic(ctx, topic, mHandler)
	log.DebugF(ctx, "task to re-run consumer by topic = %v started", topic)
}

func (p *Provider) processListenerErrors(ctx context.Context, topic string, mHandler MessageHandler, errCh chan error) {
	for {
		select {
		case err := <-errCh:
			_ = cerror.NewF(ctx, cerror.KindKafkaOther, "consumer for topic = %s stopped. %s", topic, err.Error()).LogError()

			p.tryRerunTopicListener(ctx, topic, mHandler)

			continue
		case <-ctx.Done():
			return
		}
	}
}

func (p *Provider) syncTrace(ctx context.Context, compName, operName string, e event.BaseEvent, err error) {
	if p.trace != nil {
		p.trace.Trace(compName)
		tt := p.trace.GetTrace()

		_, span := tt.Start(ctx, operName)
		defer span.End()

		attrs := make(map[string]string)

		if e != nil {
			attrs["broker.event.id"] = e.GetID()
			attrs["broker.event.header"] = fmt.Sprintf("%+v", e.GetHeader())
		}

		if err != nil {
			attrs["error.message"] = err.Error()
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		reqID := ctx.Value(consts.HeaderXRequestID)
		if reqID != nil {
			attrs["http.requestID"] = fmt.Sprintf("%s", reqID)
		}

		for key, val := range attrs {
			span.SetAttributes(attribute.Key(key).String(val))
		}
	}
}
