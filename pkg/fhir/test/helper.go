package test

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"testing"

	"github.com/tj/assert"
)

var _cb = context.Background()

func assertValidationError(t *testing.T, err error, expKey, expValue string) {
	t.Helper()

	vErr, ok := err.(*cerror.ValidationError)

	assert.True(t, ok)
	assert.Equal(t, map[string]string{expKey: expValue}, vErr.Payload())
}

func assertEmptyError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}

	vErr, ok := err.(*cerror.ValidationError)
	if ok {
		assert.Empty(t, vErr.Payload())
		return
	}

	assert.NoError(t, err)
}

//nolint:gosec
const _accessToken = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhenAiOiJhZG1pbi1jbGkifQ.aymmdhucDxN_XraZV4Sh6V7h8t1fyPIHSiiZU3H_x88`
