package cerror

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror/trace"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"runtime"
)

type CError struct {
	err     error
	ctx     context.Context
	kind    Kind
	ops     []string
	stack   string
	payload interface{}
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
		_internalMethodsNestingLevel = 2
		_nestingLevelsToLog          = 6
	)

	skipOp := _internalMethodsNestingLevel
	op := []string{}

	for i := 0; i < _nestingLevelsToLog; i++ {
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
		stack: fmt.Sprintf("%+v", trace.WithStack(err)),
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
		WithValue(logger.FieldErrorGroup, e.Kind().Group().String()).
		WithValue(logger.FieldErrorKind, e.Kind().String()).
		WithValue(logger.FieldErrorCode, e.Kind().HTTPCode()).
		WithValue(logger.FieldPayload, e.Payload()).
		WithValue(logger.FieldOperations, e.Ops()))

	log.Log(logger.NewEvent(e.Ctx(), logger.LevelTrace, "").
		WithValue(logger.FieldError, e.Error()).
		WithValue(logger.FieldStackTrace, e.StackTrace()))

	return e
}
