package consts_test

import (
	"kafka-polygon/pkg/http/consts"
	"testing"

	"github.com/tj/assert"
)

func TestConst(t *testing.T) {
	for k, v := range map[string]string{
		consts.HeaderXRequestID:     "X-Request-ID",
		consts.HeaderXClientID:      "X-Client-ID",
		consts.HeaderXUserID:        "X-User-ID",
		consts.BackendAppName:       "wasfaty",
		consts.WorkflowRoutesPrefix: "/workflows",
	} {
		assert.Equal(t, v, k)
	}
}

func TestRequestHeadersToSave(t *testing.T) {
	assert.Equal(t, []string{
		consts.HeaderXRequestID,
		consts.HeaderXClientID,
		consts.HeaderXUserID,
	}, consts.RequestHeadersToSave())
}

func TestResponseHeadersToSend(t *testing.T) {
	assert.Equal(t, []string{
		consts.HeaderXRequestID,
	}, consts.ResponseHeadersToSend())
}
