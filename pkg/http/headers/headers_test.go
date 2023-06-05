package headers_test

import (
	"context"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/http/headers"
	"net/http"
	"testing"

	"github.com/tj/assert"
	"github.com/valyala/fasthttp"
)

func TestAddHeadersFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	for _, v := range consts.RequestHeadersToSave() {
		//nolint:staticcheck
		ctx = context.WithValue(ctx, v, v)
	}

	req, _ := http.NewRequest(http.MethodGet, "localhost", http.NoBody)

	headers.AddHeadersFromContext(ctx, req)

	for _, v := range consts.RequestHeadersToSave() {
		assert.Equal(t, v, req.Header.Get(v))
	}

	fastReq := fasthttp.AcquireRequest()

	headers.AddHeadersFromContext(ctx, fastReq)

	for _, v := range consts.RequestHeadersToSave() {
		assert.Equal(t, v, string(fastReq.Header.Peek(v)))
	}
}

func TestAddHeadersFromContextDefaultValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	req, _ := http.NewRequest(http.MethodGet, "test", http.NoBody)

	headers.AddHeadersFromContext(ctx, req)

	for _, v := range consts.RequestHeadersToSave() {
		if v == consts.HeaderXClientID {
			assert.Equal(t, consts.BackendAppName, req.Header.Get(v))
		} else {
			assert.Equal(t, "", req.Header.Get(v))
		}
	}

	fastReq := fasthttp.AcquireRequest()

	headers.AddHeadersFromContext(ctx, fastReq)

	for _, v := range consts.RequestHeadersToSave() {
		if v == consts.HeaderXClientID {
			assert.Equal(t, consts.BackendAppName, string(fastReq.Header.Peek(v)))
		} else {
			assert.Equal(t, "", string(fastReq.Header.Peek(v)))
		}
	}
}

func TestContextHeadersSync(t *testing.T) {
	t.Parallel()

	nCtx := context.Background()
	oldCtx := context.Background()

	for _, v := range consts.RequestHeadersToSave() {
		//nolint:staticcheck
		oldCtx = context.WithValue(oldCtx, v, v)
	}

	nCtx = headers.ContextHeadersSync(oldCtx, nCtx)

	for _, key := range consts.RequestHeadersToSave() {
		val, ok := nCtx.Value(key).(string)
		assert.True(t, ok)
		assert.NotEmpty(t, val)
	}
}
