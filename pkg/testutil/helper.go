package testutil

import (
	"kafka-polygon/pkg/cerror"
	"testing"

	"github.com/tj/assert"
)

func AssertCError(t *testing.T, expText, expKind string, err error) {
	t.Helper()

	cErr, ok := err.(*cerror.CError)
	assert.True(t, ok)
	assert.Equal(t, expKind, cErr.Kind().String())
	assert.Equal(t, expText, cErr.Error())
}

func AssertCErrorsEqual(t *testing.T, expErr, actErr error) {
	t.Helper()

	expErrC, ok := expErr.(*cerror.CError)
	if !ok {
		t.Fatal("expected error is not *cerror.CError")
	}

	actErrC, ok := actErr.(*cerror.CError)
	if !ok {
		t.Fatal("actual error is not *cerror.CError")
	}

	assert.Equal(t, expErrC.Kind().String(), actErrC.Kind().String())
	assert.Equal(t, expErrC.Error(), actErrC.Error())
}

func AssertValidationError(t *testing.T, expKey, expValue string, err error) {
	t.Helper()

	vErr, ok := err.(*cerror.ValidationError)

	assert.True(t, ok)
	assert.Equal(t, map[string]string{expKey: expValue}, vErr.Payload())
}

func AssertValidationErrorsEqual(t *testing.T, expErr, actErr error) {
	t.Helper()

	expErrV, ok := expErr.(*cerror.ValidationError)
	if !ok {
		t.Fatal("expected error is not *cerror.ValidationError")
	}

	actErrV, ok := actErr.(*cerror.ValidationError)
	if !ok {
		t.Fatal("actual error is not *cerror.ValidationError")
	}

	assert.Equal(t, expErrV.Payload(), actErrV.Payload())
}
