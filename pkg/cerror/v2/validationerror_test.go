package cerror_test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror/v2"
	"testing"

	"github.com/tj/assert"
)

type requestID string

func TestNewValidationError(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestID("X-Request-ID"), requestIDText)
	fields := map[string]string{"name": "required"}
	err := cerror.NewValidationError(ctx, fields)
	assert.Equal(t, ctx.Value("X-Request-ID"), err.Ctx().Value("X-Request-ID"))
	assert.Equal(t, fmt.Errorf("name:required"), err.Err())
	assert.Equal(t, "some fields are invalid", err.Error())
	assert.Equal(t, cerror.KindValidation, err.Kind())
	assert.Contains(t, err.Ops()[0], "TestNewValidationError")
	assert.Equal(t, fields, err.Payload())
}

func TestValidationErrorFields(t *testing.T) {
	fields := map[string]string{"name": "required", "code": "length"}
	expected := make(map[string]interface{})

	for k, v := range fields {
		expected[k] = map[string]string{"message": v}
	}

	assert.Equal(t, expected, cerror.NewValidationError(context.Background(), fields).Fields())
}

func TestValidationErrorWithPayload(t *testing.T) {
	fieldsEmpty := make(map[string]string)
	expPayload := map[string]string{"test-payload-key": "test-payload-value"}

	err := cerror.NewValidationError(context.Background(), fieldsEmpty)
	assert.Equal(t, map[string]string{}, err.Payload())

	err = err.WithPayload(expPayload)
	assert.Equal(t, expPayload, err.Payload())
}
