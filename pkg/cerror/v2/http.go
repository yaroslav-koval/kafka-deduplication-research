package cerror

import (
	"context"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
)

// KindError should provide an underlying error's kind.
type KindError interface {
	Kind() Kind
}

// CtxError should provide an underlying error's context.
type CtxError interface {
	Ctx() context.Context
}

// FieldsError should provide an underlying error's error fields.
// It is mostly used for validation error that have error description for each field.
type FieldsError interface {
	Fields() map[string]interface{}
}

// HTTPCodeError represents an error that has a related http code.
type HTTPCodeError interface {
	HTTPCode() (int, bool)
}

// AttributesError represents an error that has additional attributes.
type AttributesError interface {
	Attributes() *ErrAttributes
}

// HTTPResponseErrorWrap is structure that describes error http responses
// from our services. It can be used to unmarshal unsuccess responses' body.
type HTTPResponseErrorWrap struct {
	Error HTTPResponseError `json:"error,omitempty"`
}

type HTTPResponseError struct {
	Message    string                 `json:"message,omitempty"`
	Kind       string                 `json:"kind,omitempty"`
	Attributes *ErrAttributes         `json:"attributes,omitempty"`
	Errors     map[string]interface{} `json:"errors,omitempty"`
}

// BuildErrorHTTPResponse builds HTTP response payload based on given error.
// It checks whether error meets specific interfaces and takes values
// returned from interface's methods make payload with common error fields.
func BuildErrorHTTPResponse(err error) *HTTPResponseErrorWrap {
	if err == nil {
		return nil
	}

	errKind := ErrKind(err)
	r := HTTPResponseError{
		Message: err.Error(),
		Kind:    errKind.String(),
	}

	if e, ok := err.(AttributesError); ok {
		r.Attributes = e.Attributes()
	}

	if e, ok := err.(FieldsError); ok {
		r.Errors = e.Fields()
	}

	return &HTTPResponseErrorWrap{Error: r}
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

	kind = KindOther
	if e, ok := err.(KindError); ok {
		kind = e.Kind()
	}

	e := logger.NewEvent(ctx, logLevelByKind(kind), "http request was processed with error").WithErr(err)

	e.WithValue(logger.FieldErrorKind, kind.String())
	e.WithValue(logger.FieldErrorCode, HTTPCode(err))

	log.Log(e)
}
