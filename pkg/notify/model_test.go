package notify_test

import (
	"kafka-polygon/pkg/notify"
	"net/http"
	"testing"
	"time"

	"github.com/tj/assert"
)

func TestMessageResponse(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	nMR := notify.MessageResponse{
		MessageID:          "test-message-id",
		ErrorCode:          1,
		Message:            "test-message",
		ResponseStatusCode: http.StatusOK,
		SubmittedAt:        now,
	}

	assert.Equal(t, "test-message-id", nMR.MessageID)
	assert.Equal(t, "test-message", nMR.Message)
	assert.Equal(t, int64(1), nMR.ErrorCode)
	assert.Equal(t, http.StatusOK, nMR.ResponseStatusCode)
	assert.Equal(t, now, nMR.SubmittedAt)
}

func TestMessageStatusRequest(t *testing.T) {
	t.Parallel()

	nSR := notify.MessageStatusRequest{
		ID:    "test-id",
		MsgID: "test-message-id",
	}

	assert.Equal(t, "test-id", nSR.ID)
	assert.Equal(t, "test-message-id", nSR.MsgID)
}

func TestStatusResponse(t *testing.T) {
	t.Parallel()

	nSR := notify.StatusResponse{
		Status:  "test-status",
		Message: "test-message",
	}

	assert.Equal(t, "test-status", nSR.Status)
	assert.Equal(t, "test-message", nSR.Message)
}
