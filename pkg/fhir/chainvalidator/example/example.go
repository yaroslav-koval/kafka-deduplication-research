//nolint:gomnd
package main

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/fhir/chainvalidator"
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

// go run pkg/fhir/chainvalidator/example/example.go
func main() {
	runParallel()
	runParallelWithSequential()
}

func runParallelWithSequential() {
	validation1 := chainvalidator.NewValidation(equalNumbersValidator, &equalNumbersValidatorParams{
		a: 1, b: 1, c: 1,
	})

	validation2 := chainvalidator.NewValidation(equalStringsValidator, &equalStringsValidatorParams{
		str1: "a", str2: "a",
	})

	validation3 := chainvalidator.NewValidation(sumLessThan10Validator, &sumLessThan10ValidatorParams{
		a: 1, b: 1,
	})

	secondaryGroup := chainvalidator.NewGroup(validation1, validation2)
	secondaryGroup.EnableParallel()

	group := chainvalidator.NewGroup(secondaryGroup, validation3)

	_ = group.Validate(_cb)
}

func runParallel() {
	validation1 := chainvalidator.NewValidation(equalNumbersValidator, &equalNumbersValidatorParams{
		a: 1, b: 1, c: 1,
	})

	validation2 := chainvalidator.NewValidation(equalStringsValidator, &equalStringsValidatorParams{
		str1: "a", str2: "a",
	})

	validation3 := chainvalidator.NewValidation(sumLessThan10Validator, &sumLessThan10ValidatorParams{
		a: 1, b: 1,
	})

	group := chainvalidator.NewGroup(validation1, validation2, validation3)

	group.EnableParallel() // disable to see sequential validations

	_ = group.Validate(_cb)
}

func equalNumbersValidator(ctx context.Context, params *equalNumbersValidatorParams) error {
	if params.a == params.b && params.b == params.c {
		fmt.Printf("validation 1 valid\n")
		return nil
	}

	return cerror.New(ctx, cerror.KindInternal, fmt.Errorf("numbers are not equal")).LogError()
}

func equalStringsValidator(ctx context.Context, params *equalStringsValidatorParams) error {
	if params.str1 == params.str2 {
		fmt.Printf("validation 2 valid\n")
		return nil
	}

	return cerror.New(ctx, cerror.KindInternal, fmt.Errorf("strings are not equal")).LogError()
}

func sumLessThan10Validator(ctx context.Context, params *sumLessThan10ValidatorParams) error {
	if params.a+params.b < 10 {
		fmt.Printf("validation 3 valid\n")
		return nil
	}

	return cerror.New(ctx, cerror.KindInternal, fmt.Errorf("sum of numbers more than 10")).LogError()
}
