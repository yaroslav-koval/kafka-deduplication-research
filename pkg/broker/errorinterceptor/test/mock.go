package test

import (
	"context"
	"kafka-polygon/pkg/broker/errorinterceptor/entity"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cmd/metadata"
)

type mockStoreProvider struct {
	saveEventFunc func(ctx context.Context, e *entity.FailedBrokerEvent) (*entity.FailedBrokerEvent, error)
}

func (m *mockStoreProvider) SaveEvent(ctx context.Context, e *entity.FailedBrokerEvent) (
	*entity.FailedBrokerEvent, error) {
	return m.saveEventFunc(ctx, e)
}

type mockOriginalFn struct {
	getEventDataFunc func(_ context.Context) event.BaseEvent
	callFnFunc       func(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error
}

func (m *mockOriginalFn) GetEventData(ctx context.Context) event.BaseEvent {
	return m.getEventDataFunc(ctx)
}

func (m *mockOriginalFn) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	return m.callFnFunc(reqCtx, e, eventData)
}

type Event struct {
	id string
}

func (e *Event) GetID() string {
	return e.id
}

func (e *Event) GetDebug() bool {
	return false
}

func (e *Event) WithHeader(ctx context.Context) {}

func (e *Event) GetHeader() event.Header {
	return event.Header{}
}

func (e *Event) ToByte() []byte {
	return nil
}

func (e *Event) Unmarshal(msg event.Message) error {
	return nil
}

func (e *Event) GetMeta() metadata.Meta {
	return metadata.Meta{}
}

func (e *Event) WithMeta(ctx metadata.Meta) {}
