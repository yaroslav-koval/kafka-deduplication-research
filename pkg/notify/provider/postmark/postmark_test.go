package postmark_test

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/notify"
	"kafka-polygon/pkg/notify/provider/postmark"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	bgCtx        = context.Background()
	expTo        = "to@example.com"
	expFrom      = "from@example.com"
	expBody      = "<h1>Test POSTMARK</h1>"
	expTag       = "test-tag-postmark"
	expMessageID = "test-message-id"
	now          = time.Now().UTC()
)

func TestPostmark(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/email" {
			t.Errorf("Expected to request '/email', got: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)

		mr := notify.MessageResponse{
			SubmittedAt: now,
			MessageID:   expMessageID,
		}

		data, _ := json.Marshal(mr)

		_, _ = w.Write(data)
	}))

	defer server.Close()

	pm := postmark.New(postmark.Options{})
	pm.WithBaseURL(server.URL)

	actualResp, err := pm.UseHTML().
		To(expTo).
		From(expFrom).
		WithBody(expBody).
		WithTag(expTag).Send(bgCtx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, actualResp.ResponseStatusCode)
	assert.Equal(t, expMessageID, actualResp.MessageID)
	assert.Equal(t, now, actualResp.SubmittedAt)
}

func TestPostmarkNoEmptyToError(t *testing.T) {
	t.Parallel()

	pm := postmark.New(postmark.Options{})
	_, err := pm.UseHTML().
		From(expFrom).
		WithBody(expBody).
		WithTag(expTag).Send(bgCtx)
	require.Error(t, err)
	assert.Equal(t, "not empty To", err.Error())
}

func TestPostmarkNoEmptyFromError(t *testing.T) {
	t.Parallel()

	pm := postmark.New(postmark.Options{})
	_, err := pm.UseHTML().
		To(expTo).
		WithBody(expBody).
		WithTag(expTag).Send(bgCtx)
	require.Error(t, err)
	assert.Equal(t, "not empty From", err.Error())
}

func TestPostmarkError(t *testing.T) {
	t.Parallel()

	expErrMessage := "test-err-message"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)

		mr := notify.MessageResponse{
			SubmittedAt: now,
			MessageID:   "",
			ErrorCode:   http.StatusInternalServerError,
			Message:     expErrMessage,
		}

		data, _ := json.Marshal(mr)

		_, _ = w.Write(data)
	}))

	defer server.Close()

	pm := postmark.New(postmark.Options{})
	pm.WithBaseURL(server.URL)

	actualResp, err := pm.UseHTML().
		To(expTo).
		From(expFrom).
		WithBody(expBody).
		WithTag(expTag).Send(bgCtx)
	require.Error(t, err)
	assert.Equal(t, "500 test-err-message", err.Error())

	assert.Equal(t, http.StatusInternalServerError, actualResp.ResponseStatusCode)
	assert.Equal(t, "", actualResp.MessageID)
	assert.Equal(t, int64(500), actualResp.ErrorCode)
	assert.Equal(t, expErrMessage, actualResp.Message)
}
