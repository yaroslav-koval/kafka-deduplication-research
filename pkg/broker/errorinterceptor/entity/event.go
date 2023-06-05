package entity

import (
	"context"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/cmd/metadata"
)

type ErrorEvent struct {
	OriginalEvent event.BaseEvent
}

func NewErrorEvent(originalEv event.BaseEvent) *ErrorEvent {
	return &ErrorEvent{
		OriginalEvent: originalEv,
	}
}

func (e *ErrorEvent) GetID() string {
	return e.OriginalEvent.GetID()
}

func (e *ErrorEvent) GetDebug() bool {
	return e.OriginalEvent.GetDebug()
}

func (e *ErrorEvent) WithHeader(ctx context.Context) {
	e.OriginalEvent.WithHeader(ctx)
}

func (e *ErrorEvent) GetHeader() event.Header {
	return e.OriginalEvent.GetHeader()
}

func (e *ErrorEvent) ToByte() []byte {
	return e.OriginalEvent.ToByte()
}

func (e *ErrorEvent) Unmarshal(msg event.Message) error {
	return e.OriginalEvent.Unmarshal(msg)
}

func (e *ErrorEvent) GetMeta() metadata.Meta {
	return e.OriginalEvent.GetMeta()
}

func (e *ErrorEvent) WithMeta(ctx metadata.Meta) {
	e.OriginalEvent.WithMeta(ctx)
}
