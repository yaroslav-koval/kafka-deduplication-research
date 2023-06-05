package cerror

import (
	"context"
	"fmt"
)

type MultiError struct {
	errors []error
}

func NewMultiError(ctx context.Context, errs ...error) *MultiError {
	return &MultiError{
		errors: errs,
	}
}

// WithErrors add error to multi error
func (m *MultiError) WithErrors(err ...error) {
	m.errors = append(m.errors, err...)
}

// Errors returns array of errors.
func (m *MultiError) Errors() []error {
	return m.errors
}

// Error returns text representation of the validation error.
func (m *MultiError) Error() string {
	return "check error fields to understand multierror"
}

// Fields returns validation errors in map[string]string structure
func (m *MultiError) Fields() map[string]interface{} {
	result := make(map[string]interface{})
	otherErrIndex := 1

	for _, e := range m.errors {
		switch err := e.(type) {
		case *ValidationError:
			for i, v := range err.Fields() {
				result[i] = v
			}
		case *MultiError:
			for i, v := range err.Fields() {
				result[i] = v
			}
		case *CError:
			k := fmt.Sprintf("__other_%d", otherErrIndex)
			result[k] = map[string]string{"message": err.Error()}
			otherErrIndex++
		}
	}

	return result
}
