package event_test

import (
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/broker/event"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestConnectorEventType(t *testing.T) {
	t.Parallel()

	e := event.ConnectorData{}

	var (
		i interface{} = e
		p interface{} = &e
	)

	_, ok := i.(event.BaseEvent)
	assert.Equal(t, ok, false)

	_, ok = p.(event.BaseEvent)
	assert.Equal(t, ok, true)

	_, ok = p.(event.ConnectorEvent)
	assert.Equal(t, ok, true)
}

func TestConnectorEvent(t *testing.T) {
	t.Parallel()

	now := time.Now()

	e := event.ConnectorData{
		ID:           "test-id",
		Type:         "test-type",
		Notification: "test-notification",
		Instance:     "test-instance",
		Header:       expHeader,
		Time:         now,
		Debug:        true,
		Metadata:     expMeta,
	}

	assert.Equal(t, "test-id", e.GetID())
	assert.Equal(t, "test-type", e.GetType())
	assert.Equal(t, "test-notification", e.GetNotification())
	assert.Equal(t, "test-instance", e.GetInstance())
	assert.Equal(t, expHeader, e.GetHeader())
	assert.Equal(t, now, e.GetTime())
	assert.Equal(t, true, e.GetDebug())
	assert.NotNil(t, e.ToByte())
	assert.True(t, len(e.ToByte()) > 0)
	assert.Equal(t, expMeta, e.GetMeta())
}

func TestConnectorEventUnmarshal(t *testing.T) {
	t.Parallel()

	expectedData := map[string]interface{}{
		"id":           "test-id",
		"type":         "test-type",
		"notification": "test-notification",
		"c":            "test-instance",
		"header":       expHeader,
		"time":         time.Now(),
		"metadata":     expMeta,
	}

	b, err := json.Marshal(expectedData)
	require.NoError(t, err)

	msg := event.Message{
		Value: b,
	}
	e := event.ConnectorData{}
	err = e.Unmarshal(msg)
	require.NoError(t, err)

	assert.Equal(t, expectedData["id"], e.GetID())
	assert.Equal(t, expectedData["type"], e.GetType())
	assert.Equal(t, expectedData["notification"], e.GetNotification())
	assert.Equal(t, expectedData["instance"], e.GetInstance())
	assert.Equal(t, expectedData["header"], e.GetHeader())
	assert.Equal(t, expectedData["metadata"], e.GetMeta())
	assert.Contains(t, fmt.Sprintf("%s", expectedData["time"]), e.GetTime().String())
}

func TestConnectorEventError(t *testing.T) {
	t.Parallel()

	msg := event.Message{
		Value: []byte(`{"num":6.13,"strs":{"a","b"}}`),
	}

	e := &event.ConnectorData{}
	err := e.Unmarshal(msg)
	require.Error(t, err)
}
