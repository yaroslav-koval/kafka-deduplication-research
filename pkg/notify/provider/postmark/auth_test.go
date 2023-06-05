package postmark_test

import (
	"context"
	"encoding/json"
	"io"
	"kafka-polygon/pkg/notify/provider/postmark"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestPostmarkAuthTransport(t *testing.T) {
	expServerToken := "test-server-token"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}

		if r.Header.Get("X-Postmark-Server-Token") != expServerToken {
			t.Errorf("expected X-Postmark-Server-Token header, got: %s", expServerToken)
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{"value": "fine"}`))
	}))

	defer server.Close()

	client := http.Client{
		Transport: &postmark.AuthTransport{
			ServerToken: expServerToken,
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	actualData := &struct {
		Value string `json:"value"`
	}{}

	err = json.Unmarshal(body, actualData)
	require.NoError(t, err)
	assert.Equal(t, "fine", actualData.Value)
}
