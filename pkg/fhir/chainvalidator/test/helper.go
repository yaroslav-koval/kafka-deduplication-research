//nolint:gomnd
package test

import (
	"context"
	"kafka-polygon/pkg/cerror"
)

var _cb = context.Background()

type equalNumbersValidatorParams struct {
	a, b, c int
}

type equalStringsValidatorParams struct {
	str1, str2 string
}

type sumLessThan10ValidatorParams struct {
	a, b int
}

func equalNumbersValidator(ctx context.Context, params *equalNumbersValidatorParams) error {
	if params.a == params.b && params.b == params.c {
		return nil
	}

	return cerror.NewValidationError(ctx, map[string]string{"numbers": "are not equal"}).LogError()
}

func equalStringsValidator(ctx context.Context, params *equalStringsValidatorParams) error {
	if params.str1 == params.str2 {
		return nil
	}

	return cerror.NewValidationError(ctx, map[string]string{"strings": "are not equal"}).LogError()
}

func sumLessThan10Validator(ctx context.Context, params *sumLessThan10ValidatorParams) error {
	if params.a+params.b < 10 {
		return nil
	}

	return cerror.NewF(_cb, cerror.KindInternal, "sum of numbers more than 10").LogError()
}
