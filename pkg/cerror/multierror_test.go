package cerror_test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"testing"

	"github.com/tj/assert"
)

func TestNewMultiError(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestID("X-Request-ID"), requestIDText)
	fields := map[string]string{"name": "required"}
	vErr := cerror.NewValidationError(ctx, fields)

	kind := cerror.KindConflict
	originalErr := fmt.Errorf(errText)
	cErr := cerror.New(ctx, kind, originalErr)

	mErr := cerror.NewMultiError(ctx, vErr)
	mErr.WithErrors(cErr)

	assert.Equal(t, "check error fields to understand multierror", mErr.Error())
	errors := mErr.Errors()

	assert.Equal(t, 2, len(errors))

	vExpErr, ok := errors[0].(*cerror.ValidationError)
	assert.True(t, ok)
	assert.Equal(t, ctx.Value(consts.HeaderXRequestID), vExpErr.Ctx().Value(consts.HeaderXRequestID))
	assert.Equal(t, fmt.Errorf("name:required"), vExpErr.Err())
	assert.Equal(t, "read REST Manifest to understand this error", vExpErr.Error())
	assert.Equal(t, cerror.KindBadValidation, vExpErr.Kind())
	assert.Contains(t, vExpErr.Ops()[0], "TestNewMultiError")
	assert.Equal(t, fields, vExpErr.Payload())

	cExpErr, ok := errors[1].(*cerror.CError)
	assert.True(t, ok)
	assert.Equal(t, ctx.Value(consts.HeaderXRequestID), cExpErr.Ctx().Value(consts.HeaderXRequestID))
	assert.Equal(t, originalErr, cExpErr.Err())
	assert.Equal(t, originalErr.Error(), cErr.Error())
	assert.Equal(t, kind, cExpErr.Kind())
	assert.Contains(t, cExpErr.Ops()[0], "TestNewMultiError")
	assert.Nil(t, cExpErr.Payload())
}

func TestMultiErrorFields(t *testing.T) {
	fields := map[string]string{
		"name":      "required",
		"__other_1": "error"}
	expected := make(map[string]interface{})

	for k, v := range fields {
		expected[k] = map[string]string{"message": v}
	}

	ctx := context.WithValue(context.Background(), requestID("X-Request-ID"), requestIDText)
	f := map[string]string{"name": "required"}
	vErr := cerror.NewValidationError(ctx, f)

	kind := cerror.KindConflict
	originalErr := fmt.Errorf(errText)
	cErr := cerror.New(ctx, kind, originalErr)

	mErr := cerror.NewMultiError(ctx, vErr, cErr)

	assert.Equal(t, expected, mErr.Fields())
}
