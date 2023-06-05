package cerror

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
)

var ErrNotFound = fmt.Errorf("not found")

// KindError should provide an underlying error's kind.
type KindError interface {
	Kind() Kind
}

// OpError should provide a short description of an underlying error's initiator.
type OpError interface {
	Ops() []string
}

// CtxError should provide an underlying error's context.
type CtxError interface {
	Ctx() context.Context
}

// StackTraceError should provide an underlying error's stack trace.
type StackTraceError interface {
	StackTrace() string
}

// FieldsError should provide an underlying error's error fields.
// It is mostly used for validation error that have error description for each field.
type FieldsError interface {
	Fields() map[string]interface{}
}

// PayloadError should provide an underlying error's payload.
// It is mostly used for errors that have some related data.
type PayloadError interface {
	Payload() interface{}
}

// ResponseErrorWrap is structure that describes error http responses
// from our services. It can be used to unmarshal unsuccess responses' body
type ResponseErrorWrap struct {
	Error ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Message string                 `json:"message,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Group   string                 `json:"group,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
}

// BuildErrorResponse builds HTTP response payload based on given error.
// It checks whether error meets specific interfaces and takes values
// returned from interface's methods make payload with common error fields.
func BuildErrorResponse(err error) *ResponseErrorWrap {
	if err == nil {
		return nil
	}

	errKind := ErrKind(err)
	r := ResponseError{
		Message: err.Error(),
		Type:    errKind.String(),
		Group:   errKind.Group().String(),
	}

	if e, ok := err.(FieldsError); ok {
		r.Errors = e.Fields()
	}

	return &ResponseErrorWrap{Error: r}
}

// LogHTTPHandlerError defines appropriate context and calls LogHTTPHandlerErrorCtx.
// If error meets CtxError interface, it takes context from error.
// Otherwise, background context will be used.
func LogHTTPHandlerError(err error) {
	var ctx context.Context
	if e, ok := err.(CtxError); ok {
		ctx = e.Ctx()
	} else {
		ctx = context.Background()
	}

	LogHTTPHandlerErrorCtx(ctx, err)
}

// LogHTTPHandlerErrorCtx writes result error log after HTTP request processed with error.
// It checks whether error meets specific interfaces and takes values
// returned from interface's methods to write log in common structure.
func LogHTTPHandlerErrorCtx(ctx context.Context, err error) {
	var kind Kind

	kind = KindInternal
	if e, ok := err.(KindError); ok {
		kind = e.Kind()
	}

	e := logger.NewEvent(ctx, logLevelByKind(kind), "http request was processed with error").WithErr(err)

	e.WithValue(logger.FieldErrorKind, kind.String())
	e.WithValue(logger.FieldErrorCode, kind.HTTPCode())

	log.Log(e)
}

func logLevelByKind(kind Kind) logger.Level {
	if kind == KindNotExist {
		return logger.LevelWarn
	}

	return logger.LevelError
}
