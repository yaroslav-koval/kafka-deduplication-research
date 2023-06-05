package event_test

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	payload = event.DebeziumPayload{
		Source: event.DebeziumSource{
			Table: "test-table",
		},
		BeforeValues: map[string]interface{}{"key": "val"},
		AfterValues:  map[string]interface{}{"key": "val"},
		Op:           "test-op",
	}
	key = event.KeyData{
		Payload: event.KeyPayload{
			ID: "test-key-id",
		},
	}
)

func TestDebeziumDataType(t *testing.T) {
	t.Parallel()

	e := event.DebeziumData{}

	var (
		i interface{} = e
		p interface{} = &e
	)

	_, ok := i.(event.BaseEvent)
	assert.Equal(t, ok, false)

	_, ok = p.(event.BaseEvent)
	assert.Equal(t, ok, true)

	_, ok = p.(event.DebeziumEvent)

	assert.Equal(t, ok, true)
	assert.Equal(t, "", e.GetID())
}

func TestDebeziumData(t *testing.T) {
	t.Parallel()

	e := event.DebeziumData{
		Key:      &key,
		Payload:  &payload,
		Header:   expHeader,
		Metadata: expMeta,
		Debug:    true,
	}
	e.WithContext(context.Background())

	assert.Equal(t, key.Payload.ID, e.GetID())
	assert.Equal(t, true, e.GetDebug())
	assert.Equal(t, expHeader, e.GetHeader())
	assert.Equal(t, expMeta, e.GetMeta())
	assert.NotNil(t, e.ToByte())
	assert.Equal(t, true, len(e.ToByte()) > 0)
	assert.Equal(t, &payload, e.GetPayload())
}

func TestDebeziumDataUnmarshal(t *testing.T) {
	t.Parallel()

	expectedData := map[string]interface{}{
		"payload":  payload,
		"header":   expHeader,
		"metadata": expMeta,
	}

	b, err := json.Marshal(expectedData)
	require.NoError(t, err)

	msg := event.Message{
		Value: b,
	}
	e := event.DebeziumData{}
	err = e.Unmarshal(msg)
	require.NoError(t, err)

	assert.Equal(t, "", e.GetID())
	assert.Equal(t, false, e.GetDebug())
	assert.Equal(t, &payload, e.GetPayload())
	assert.Equal(t, expectedData["header"], e.GetHeader())
	assert.Equal(t, expectedData["metadata"], e.GetMeta())
}

func TestDebeziumDataError(t *testing.T) {
	t.Parallel()

	msg := event.Message{
		Value: []byte(`{"num2":6.13,"strs2":{"a","b"}}`),
	}

	e := &event.DebeziumData{}
	err := e.Unmarshal(msg)
	require.Error(t, err)
}
