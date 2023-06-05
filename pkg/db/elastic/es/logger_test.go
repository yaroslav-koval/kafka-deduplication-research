package es_test

import (
	"context"
	"kafka-polygon/pkg/db/elastic/es"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestCJSONLogger(t *testing.T) {
	t.Parallel()

	l := es.CJSONLogger{}
	l.WithContext(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/upper?word=abc", nil)
	w := httptest.NewRecorder()
	resp := w.Result()

	defer resp.Body.Close()

	err := l.LogRoundTrip(req, resp, nil, time.Now(), time.Duration(20)*time.Second)
	require.NoError(t, err)
	assert.Equal(t, true, l.RequestBodyEnabled())
	assert.Equal(t, true, l.ResponseBodyEnabled())
}
