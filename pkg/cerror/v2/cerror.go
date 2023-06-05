package cerror

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"runtime"

	"github.com/pkg/errors"
)

type CError struct {
	err      error
	ctx      context.Context
	kind     Kind
	httpCode int
	ops      []string
	stack    string
	payload  interface{}
	attrs    *ErrAttributes
}

// New wraps given error in common error structure.
func New(ctx context.Context, kind Kind, err error) *CError {
	return newWithOp(ctx, kind, err)
}

// NewF formats error based on given message and arguments and calls New.
func NewF(ctx context.Context, kind Kind, msg string, args ...interface{}) *CError {
	return newWithOp(ctx, kind, fmt.Errorf(msg, args...))
}

func newWithOp(ctx context.Context, kind Kind, err error) *CError {
	if ctx == nil {
		ctx = context.Background()
	}

	const (
		// how many functions calls are performed inside the current pkg
		// after calling New or NewF. is useful to exclude them from the stack traces
		internalMethodsNestingLevel = 2
		// how many operations to include in the shortened stack trace(ops field)
		nestingLevelsToLog = 6
	)

	skipOp := internalMethodsNestingLevel
	op := []string{}

	for i := 0; i < nestingLevelsToLog; i++ {
		pc, _, line, ok := runtime.Caller(skipOp)

		if ok {
			op = append(op, fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), line))
		}

		skipOp++
	}

	return &CError{
		err:   err,
		ctx:   ctx,
		kind:  kind,
		ops:   op,
		stack: fmt.Sprintf("%+v", errors.WithStack(err)),
	}
}

// Error returns text representation of the error.
func (e *CError) Error() string {
	return e.err.Error()
}

// Err returns underlying error.
func (e *CError) Err() error {
	return e.err
}

// Ops returns an array of short descriptions of place in code that initiated error.
func (e *CError) Ops() []string {
	return e.ops
}

// Kind returns underlying error kind.
func (e *CError) Kind() Kind {
	return e.kind
}

// Ctx returns underlying error context.
func (e *CError) Ctx() context.Context {
	return e.ctx
}

// StackTrace returns underlying error stack trace.
func (e *CError) StackTrace() string {
	return e.stack
}

// Payload returns underlying error payload.
func (e *CError) Payload() interface{} {
	return e.payload
}

// WithPayload adds given payload to error.
// This is useful to add some additional data that is related to error.
func (e *CError) WithPayload(payload interface{}) *CError {
	e.payload = payload
	return e
}

// HTTPCode returns a http code for underlying error.
// The second parameter indicates whether the error has a valid related http code.
func (e *CError) HTTPCode() (int, bool) {
	return e.httpCode, e.httpCode >= 400 && e.httpCode < 600
}

// WithHTTPCode sets http code that should be used for such error in http responses.
func (e *CError) WithHTTPCode(code int) *CError {
	e.httpCode = code
	return e
}

// Attributes returns error related attributes
func (e *CError) Attributes() *ErrAttributes {
	return e.attrs
}

// WithAttributes adds additional information about the error.
func (e *CError) WithAttributes(ea *ErrAttributes) *CError {
	e.attrs = ea

	return e
}

// WithAttributesHTTP adds the http specific error attributes.
// Use to provide additional information about an error that
// occured during http communication.
func (e *CError) WithAttributesHTTP(eav ErrAttributesValuesHTTP) *CError {
	e.attrs = &ErrAttributes{
		Type:    ErrAttributeTypeAPI,
		Subtype: ErrAttributeSubtypeHTTP,
		Code:    eav.Code,
	}

	return e
}

// WithAttributesRedis adds the redis specific error attributes.
// Use to provide additional information about an error that
// occured during redis interaction.
func (e *CError) WithAttributesRedis(eav ErrAttributesValuesRedis) *CError {
	e.attrs = &ErrAttributes{
		Type:    ErrAttributeTypeDB,
		Subtype: ErrAttributeSubtypeRedis,
		CodeStr: KindFromRedis(eav.Err).String(),
	}

	return e
}

// WithAttributesRedis adds the minio specific error attributes.
// Use to provide additional information about an error that
// occured during minio interaction.
func (e *CError) WithAttributesMinio(eav ErrAttributesValuesMinio) *CError {
	e.attrs = &ErrAttributes{
		Type:    ErrAttributeTypeMediaStorage,
		Subtype: ErrAttributeSubtypeMinio,
		CodeStr: KindFromMinio(eav.Err).String(),
	}

	return e
}

// WithAttributesRedis adds the kafka specific error attributes.
// Use to provide additional information about an error that
// occured during kafka interaction.
func (e *CError) WithAttributesKafka(eav ErrAttributesValuesKafka) *CError {
	e.attrs = &ErrAttributes{
		Type:    ErrAttributeTypeMsgBroker,
		Subtype: ErrAttributeSubtypeKafka,
		CodeStr: KindFromKafka(eav.Err).String(),
	}

	return e
}

// WithAttributesRedis adds the elastic specific error attributes.
// Use to provide additional information about an error that
// occured during elastic interaction.
func (e *CError) WithAttributesElastic(eav ErrAttributesValuesElastic) *CError {
	e.attrs = &ErrAttributes{
		Type:    ErrAttributeTypeDB,
		Subtype: ErrAttributeSubtypeElastic,
		CodeStr: KindFromElastic(eav.Err).String(),
	}

	return e
}

// WithAttributesRedis adds the postgres specific error attributes.
// Use to provide additional information about an error that
// occured during postgres interaction.
func (e *CError) WithAttributesPostgres(eav ErrAttributesValuesPostgres) *CError {
	e.attrs = &ErrAttributes{
		Type:    ErrAttributeTypeDB,
		Subtype: ErrAttributeSubtypePostgres,
		CodeStr: KindFromPostgres(eav.Err).String(),
	}

	return e
}

// LogError logs error in common structure with error level.
func (e *CError) LogError() *CError {
	return e.log(logger.LevelError)
}

// LogWarn logs error in common structure with warn level.
func (e *CError) LogWarn() *CError {
	return e.log(logger.LevelWarn)
}

// LogFatal logs error in common structure with fatal level.
// !!! Doesn't cause panic
func (e *CError) LogFatal() *CError {
	return e.log(logger.LevelFatal)
}

func (e *CError) log(lvl logger.Level) *CError {
	log.Log(logger.NewEvent(e.Ctx(), lvl, "").
		WithErr(e.Err()).
		WithValue(logger.FieldErrorKind, e.Kind().String()).
		WithValue(logger.FieldErrorCode, HTTPCode(e)).
		WithValue(logger.FieldErrorAttributes, e.attrs).
		WithValue(logger.FieldPayload, e.Payload()).
		WithValue(logger.FieldOperations, e.Ops()))

	log.Log(logger.NewEvent(e.Ctx(), logger.LevelTrace, "").
		WithValue(logger.FieldError, e.Error()).
		WithValue(logger.FieldStackTrace, e.StackTrace()))

	return e
}
