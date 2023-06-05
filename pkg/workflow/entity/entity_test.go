package entity_test

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

var (
	_bgCtx = context.Background()
)

type stepWorkerTest struct{}

func (s *stepWorkerTest) Run(ctx context.Context, e event.WorkflowEvent) (json.RawMessage, error) {
	return []byte("{}"), nil
}

func assertMultiValidationError(t *testing.T, expErrMap map[string]string, actErr error) {
	t.Helper()

	vErr, ok := actErr.(*cerror.ValidationError)

	assert.True(t, ok)
	assert.Equal(t, expErrMap, vErr.Payload())
}

func TestIDString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.ID(s).String())
}

func TestPointerID(t *testing.T) {
	s := entity.ID("hello")
	assert.Equal(t, &s, entity.PointerID(s.String()))
}
