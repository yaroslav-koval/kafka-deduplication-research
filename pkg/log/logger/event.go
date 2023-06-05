package logger

import (
	"context"
	"fmt"
)

// LogEvent provides common interface for log event.
// Event serves as a data source for logging.
type LogEvent interface {
	// Level should provide an underlying log event's level.
	Level() Level
	// Ctx should provide an underlying log event's context.
	Ctx() context.Context
	// Msg should provide an underlying log event's main message.
	Msg() string
	// Err should provide an underlying log event's error if it has so.
	Err() error
	// Fields should provide an underlying log event's related data.
	Fields() Fields
}

type Event struct {
	lvl    Level
	msg    string
	err    error
	ctx    context.Context
	fields Fields
}

// NewEvent creates log event with base required fields.
func NewEvent(ctx context.Context, l Level, msg string) *Event {
	return &Event{lvl: l, msg: msg, ctx: ctx}
}

// NewEventF formats message from message string and provided args and call New.
func NewEventF(ctx context.Context, l Level, msg string, args ...interface{}) *Event {
	return NewEvent(ctx, l, fmt.Sprintf(msg, args...))
}

// WithErr supplies event with given error.
func (e *Event) WithErr(err error) *Event {
	e.err = err
	return e
}

// WithValue supplies event with related key and value.
func (e *Event) WithValue(key string, value interface{}) *Event {
	e.fields = e.fields.AddValue(key, value)

	return e
}

// WithPayload supplies event with given payload.
// Payload is saved to log event related values with system predefined key.
func (e *Event) WithPayload(payload interface{}) *Event {
	e.fields = e.fields.AddValue(FieldPayload, payload)

	return e
}

// Level provides log event's level.
func (e *Event) Level() Level {
	return e.lvl
}

// Ctx provides log event's context.
func (e *Event) Ctx() context.Context {
	if e.ctx == nil {
		return context.Background()
	}

	return e.ctx
}

// Msg provides log event's main message.
func (e *Event) Msg() string {
	return e.msg
}

// Err provides log event's error if it has so.
func (e *Event) Err() error {
	return e.err
}

// Fields provides log event's related key/value data.
func (e *Event) Fields() Fields {
	return e.fields
}
