package middleware_test

import (
	"context"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/middleware"
	"net/http"
	"testing"

	"github.com/tj/assert"
)

func TestDefaultHeaderXRequestID(t *testing.T) {
	t.Parallel()

	key := consts.HeaderXRequestID
	expVal := "test-val"

	actualVal := middleware.DefaultHeaderXRequestID(key, expVal)
	assert.Equal(t, expVal, actualVal)

	uuidVal := middleware.DefaultHeaderXRequestID(key, "")
	assert.True(t, uuidVal != "" && uuidVal != expVal)

	notCorrKey := "test-key"
	cuuVal := middleware.DefaultHeaderXRequestID(notCorrKey, expVal)
	assert.Equal(t, expVal, cuuVal)
}

func TestValidateContentTypeJSON(t *testing.T) {
	t.Parallel()

	bgCtx := context.Background()
	expMethod := http.MethodGet
	expCT := "application/json"

	err := middleware.ValidateContentTypeJSON(bgCtx, expMethod, expCT)
	assert.Nil(t, err)

	expCT = "text/plain"
	err = middleware.ValidateContentTypeJSON(bgCtx, expMethod, expCT)
	assert.Error(t, err)
	assert.Equal(t,
		"content-type header should be application/json. Current value: text/plain",
		err.Error())
}
