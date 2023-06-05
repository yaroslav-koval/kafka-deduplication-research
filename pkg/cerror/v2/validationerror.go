package cerror

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type ValidationError struct {
	*CError
	errFields map[string]string
}

// NewValidationError creates error based on given map that describes
// fields with errors where keys are field names and values are error descriptions.
func NewValidationError(ctx context.Context, errFields map[string]string) *ValidationError {
	errs := make([]string, 0)
	for k, v := range errFields {
		errs = append(errs, k+":"+v)
	}

	cErr := newWithOp(ctx, KindValidation, fmt.Errorf("%s", strings.Join(errs, ";")))

	return &ValidationError{
		CError:    cErr.WithHTTPCode(http.StatusUnprocessableEntity).WithPayload(errFields),
		errFields: errFields,
	}
}

// Error returns text representation of the validation error.
func (e *ValidationError) Error() string {
	return "some fields are invalid"
}

// Fields returns underlying validation errors in common structure.
func (e *ValidationError) Fields() map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range e.errFields {
		m[k] = map[string]string{"message": v}
	}

	return m
}

// WithPayload adds given payload to validation error.
// This is useful to add some additional data that is related to error.
func (e *ValidationError) WithPayload(payload interface{}) *ValidationError {
	_ = e.CError.WithPayload(payload)
	return e
}

// LogError logs error in common structure with error level.
func (e *ValidationError) LogError() *ValidationError {
	_ = e.CError.LogError()
	return e
}

// LogWarn logs error in common structure with warn level.
func (e *ValidationError) LogWarn() *ValidationError {
	_ = e.CError.LogWarn()
	return e
}

// LogFatal logs error in common structure with fatal level.
// !!! Doesn't cause panic
func (e *ValidationError) LogFatal() *ValidationError {
	_ = e.CError.LogFatal()
	return e
}
