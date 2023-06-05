package notify_test

import (
	"kafka-polygon/pkg/notify"
	"testing"

	"github.com/tj/assert"
)

func TestConstStatus(t *testing.T) {
	t.Parallel()

	sConst := map[string]string{
		notify.StatusNew:        "NEW",
		notify.StatusProcessing: "PROCESSING",
		notify.StatusSent:       "SENT",
		notify.StatusFailed:     "FAILED",
		notify.StatusDelivered:  "DELIVERED",
		notify.StatusSeen:       "SEEN",
		notify.StatusReplied:    "REPLIED",
		notify.StatusQueue:      "QUEUED",
	}

	for actual, expected := range sConst {
		assert.Equal(t, actual, expected)
	}
}

func TestConstChannel(t *testing.T) {
	t.Parallel()

	sConst := map[notify.TypeNotify]string{
		notify.TypeEmail: "email",
		notify.TypeSMS:   "sms",
	}

	for actual, expected := range sConst {
		assert.Equal(t, notify.TypeNotify(expected), actual)
	}
}

func TestConstProvider(t *testing.T) {
	t.Parallel()

	sConst := map[string]string{
		notify.ProviderPostmark: "postmark",
		notify.ProviderTwilio:   "twilio",
	}

	for actual, expected := range sConst {
		assert.Equal(t, expected, actual)
	}
}
