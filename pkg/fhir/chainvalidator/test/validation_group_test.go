package test

import (
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/fhir/chainvalidator"
	"testing"

	"github.com/tj/assert"
)

func TestParallelInvoke(t *testing.T) {
	v1 := chainvalidator.NewValidation(equalNumbersValidator, &equalNumbersValidatorParams{
		a: 1, b: 2, c: 3,
	})

	v2 := chainvalidator.NewValidation(equalStringsValidator, &equalStringsValidatorParams{
		str1: "a", str2: "b",
	})

	v3 := chainvalidator.NewValidation(sumLessThan10Validator, &sumLessThan10ValidatorParams{
		a: 9, b: 1,
	})

	g := chainvalidator.NewGroup(v1, v2, v3)
	g.EnableParallel()

	err := g.Validate(_cb)
	assert.Error(t, err)

	mErr := err.(*cerror.MultiError)
	assert.Equal(t, 3, len(mErr.Errors()))

	g = chainvalidator.NewGroup(v1, v2, v3)
	err = g.Validate(_cb)
	assert.Error(t, err)

	vErr := err.(*cerror.ValidationError)
	assert.Equal(t, 1, len(vErr.Fields()))
}
