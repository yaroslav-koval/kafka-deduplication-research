package provider

import (
	"context"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/store"
)

type HandlerFn interface {
	GetEventData(_ context.Context) event.BaseEvent
	CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error
}

// HandlerWorkflow type of WorkflowEvent
type HandlerWorkflow func(context.Context, event.WorkflowEvent, store.EventProcessData) error

func (hw HandlerWorkflow) GetEventData(_ context.Context) event.BaseEvent {
	return &event.WorkflowData{}
}

func (hw HandlerWorkflow) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	if ed, ok := e.(event.WorkflowEvent); ok {
		return hw(reqCtx, ed, eventData)
	}

	return nil
}

// HandlerDebezium type of DebeziumEvent
type HandlerDebezium func(context.Context, event.DebeziumEvent, store.EventProcessData) error

func (hd HandlerDebezium) GetEventData(ctx context.Context) event.BaseEvent {
	e := event.NewDebeziumData()
	e.WithContext(ctx)

	return e
}

func (hd HandlerDebezium) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	if ed, ok := e.(event.DebeziumEvent); ok {
		return hd(reqCtx, ed, eventData)
	}

	return nil
}

// HandlerMinio type of MinioEvent
type HandlerMinio func(context.Context, event.MinioEvent, store.EventProcessData) error

func (hm HandlerMinio) GetEventData(_ context.Context) event.BaseEvent {
	return &event.MinioData{}
}

func (hm HandlerMinio) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	if ed, ok := e.(event.MinioEvent); ok {
		return hm(reqCtx, ed, eventData)
	}

	return nil
}

// HandlerNotification type of NotificationEvent
type HandlerNotification func(context.Context, event.NotificationEvent, store.EventProcessData) error

func (hn HandlerNotification) GetEventData(_ context.Context) event.BaseEvent {
	return &event.NotificationData{}
}

func (hn HandlerNotification) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	if ed, ok := e.(event.NotificationEvent); ok {
		return hn(reqCtx, ed, eventData)
	}

	return nil
}
