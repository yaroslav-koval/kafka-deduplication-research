package chainvalidator

import (
	"context"
)

type Validation[T any] struct {
	validateFn     func(ctx context.Context, params *T) error
	validateParams *T
}

func NewValidation[T any](validateFn func(ctx context.Context, validateParams *T) error, validateParams *T) *Validation[T] {
	return &Validation[T]{
		validateParams: validateParams,
		validateFn:     validateFn,
	}
}

func (v *Validation[T]) Validate(ctx context.Context) error {
	return v.validateFn(ctx, v.validateParams)
}
