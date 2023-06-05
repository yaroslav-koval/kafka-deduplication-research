package broker

import (
	"context"
	"errors"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/cmd/metadata"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/tracing"
)

var (
	errNotEmptyTopicName = errors.New("not empty topic name")
)

type QueueBroker interface {
	SetMeta(meta metadata.Meta)
	SetStore(s store.Store)
	SetTracing(t tracing.Tracer)
	GetIsTopicExists(ctx context.Context, topic string) (bool, error)
	Send(ctx context.Context, topic string, e event.BaseEvent) error
	Watch(ctx context.Context, topic string, fn provider.HandlerFn)
	Stop()
	Name() string
}

type Broker struct {
	provider provider.Provider
	metadata metadata.Meta
}

func New(p provider.Provider) QueueBroker {
	return &Broker{
		provider: p,
		metadata: metadata.New(),
	}
}

func (b *Broker) SetMeta(meta metadata.Meta) {
	b.metadata = meta
}

func (b *Broker) SetStore(s store.Store) {
	b.provider.SetStore(s)
}

func (b *Broker) SetTracing(t tracing.Tracer) {
	b.provider.SetTracing(t)
}

func (b *Broker) GetIsTopicExists(ctx context.Context, topic string) (bool, error) {
	if topic == "" {
		return false, cerror.New(ctx, cerror.KindInternal, errNotEmptyTopicName).LogError()
	}

	return b.provider.GetIsTopicExists(ctx, topic)
}

func (b *Broker) Send(ctx context.Context, topic string, e event.BaseEvent) error {
	log.DebugF(ctx, "[queueBroker] send: %s %s %s", b.provider.GetType(), topic, e.GetID())
	e.WithHeader(ctx)
	e.WithMeta(b.metadata)

	return b.provider.Publish(ctx, topic, e)
}

func (b *Broker) Watch(ctx context.Context, topic string, fn provider.HandlerFn) {
	log.DebugF(ctx, "[queueBroker] sync messages %s", b.provider.GetType())
	b.provider.Sync(ctx, topic, fn)
}

func (b *Broker) Stop() {
	b.provider.Stop()
}

func (b *Broker) Name() string {
	return b.provider.GetType()
}
