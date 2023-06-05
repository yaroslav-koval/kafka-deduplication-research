package test

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"testing"

	"github.com/tj/assert"
)

var _cb = context.Background()

//nolint:cerrl
var errNoSessionKeyInCtx = cerror.NewF(_cb, cerror.KindInternal, "no session key in context")

func ctxWithReqID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, consts.HeaderXRequestID, id) //nolint:staticcheck
}

func assertErrorEqual(t *testing.T, expErr, actErr error) {
	t.Helper()

	assert.Equal(t, cerror.ErrKind(expErr).String(), cerror.ErrKind(actErr).String())
	assert.Equal(t, expErr.Error(), actErr.Error())
}
