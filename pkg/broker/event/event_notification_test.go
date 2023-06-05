package event_test

import (
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestNotificationEventType(t *testing.T) {
	t.Parallel()

	e := event.NotificationData{}

	var (
		i interface{} = e
		p interface{} = &e
	)

	_, ok := i.(event.BaseEvent)
	assert.Equal(t, ok, false)

	_, ok = p.(event.BaseEvent)
	assert.Equal(t, ok, true)

	_, ok = p.(event.NotificationEvent)

	assert.Equal(t, ok, true)
	assert.Equal(t, "", e.GetID())
}

func TestNotificationEvent(t *testing.T) {
	t.Parallel()

	e := event.NotificationData{
		ID:         "tests-id",
		WorkflowID: "test-wf-id",
		Channel:    "test-channel",
		Provider:   "test-provider",
		Header:     expHeader,
		Metadata:   expMeta,
		Debug:      true,
	}

	assert.Equal(t, expHeader, e.GetHeader())
	assert.Equal(t, expMeta, e.GetMeta())
	assert.Equal(t, "tests-id", e.GetID())
	assert.Equal(t, true, e.GetDebug())
	assert.NotNil(t, e.ToByte())
	assert.Equal(t, true, len(e.ToByte()) > 0)
	assert.Equal(t, "test-wf-id", e.GetWorkflowID())
	assert.Equal(t, "test-channel", e.GetChannel())
	assert.Equal(t, "test-provider", e.GetProvider())
}

func TestNotificationEventUnmarshal(t *testing.T) {
	t.Parallel()

	expectedData := map[string]interface{}{
		"id":          "tests-id",
		"workflow_id": "test-wf-id",
		"channel":     "test-channel",
		"provider":    "test-provider",
		"header":      expHeader,
	}

	b, err := json.Marshal(expectedData)
	require.NoError(t, err)

	msg := event.Message{
		Value: b,
	}
	e := event.NotificationData{}
	err = e.Unmarshal(msg)
	require.NoError(t, err)

	assert.Equal(t, expectedData["header"], e.GetHeader())
	assert.Equal(t, expectedData["id"], e.GetID())
	assert.Equal(t, expectedData["workflow_id"], e.GetWorkflowID())
	assert.Equal(t, expectedData["channel"], e.GetChannel())
	assert.Equal(t, expectedData["provider"], e.GetProvider())
	assert.Equal(t, false, e.GetDebug())
}

func TestNotificationEventError(t *testing.T) {
	t.Parallel()

	msg := event.Message{
		Value: []byte(`{"num1":6.13,"strs1":{"a","b"}}`),
	}

	e := &event.NotificationData{}
	err := e.Unmarshal(msg)
	require.Error(t, err)
}
