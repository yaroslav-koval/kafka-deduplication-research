package event_test

import (
	"context"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/http/consts"
	"testing"

	"github.com/tj/assert"
)

func TestHeader(t *testing.T) {
	t.Parallel()

	expReqID := "test-x-request-id"

	e := event.Header{}

	assert.Equal(t, "", e.RequestID)

	//nolint:staticcheck
	ctx := context.WithValue(context.Background(), consts.HeaderXRequestID, expReqID)

	e.XRequestIDFromContext(ctx)

	assert.Equal(t, expReqID, e.RequestID)
}

func TestHeaderEmpty(t *testing.T) {
	t.Parallel()

	e := event.Header{}

	//nolint:staticcheck
	ctx := context.WithValue(context.Background(), consts.HeaderXRequestID, "")

	e.XRequestIDFromContext(ctx)

	assert.Equal(t, "", e.RequestID)
}
