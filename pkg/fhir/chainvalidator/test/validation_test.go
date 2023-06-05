package test

import (
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/fhir/chainvalidator"
	"testing"

	"github.com/tj/assert"
)

func TestValidatorInvoke(t *testing.T) {
	v := chainvalidator.NewValidation(equalNumbersValidator, &equalNumbersValidatorParams{
		a: 1, b: 2, c: 3,
	})
	err := v.Validate(_cb)

	assert.Error(t, err)

	vErr := err.(*cerror.ValidationError)

	assert.Equal(t, "numbers:are not equal", vErr.Err().Error())

	v = chainvalidator.NewValidation(equalNumbersValidator, &equalNumbersValidatorParams{
		a: 1, b: 1, c: 1,
	})

	err = v.Validate(_cb)

	assert.NoError(t, err)
}
