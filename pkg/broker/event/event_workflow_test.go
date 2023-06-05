package event_test

import (
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/cmd/metadata"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	expMeta = metadata.Meta{
		Version: "0.0.1",
	}

	expHeader = event.Header{
		RequestID: "test-x-request-id",
	}
)

func TestWorkflowEventType(t *testing.T) {
	t.Parallel()

	e := event.WorkflowData{}

	var (
		i interface{} = e
		p interface{} = &e
	)

	_, ok := i.(event.BaseEvent)
	assert.Equal(t, ok, false)

	_, ok = p.(event.BaseEvent)
	assert.Equal(t, ok, true)

	_, ok = p.(event.WorkflowEvent)
	assert.Equal(t, ok, true)
}

func TestWorkflowEvent(t *testing.T) {
	t.Parallel()

	expPayload, _ := json.Marshal(map[string]string{
		"test-key": "test-value",
	})

	e := event.WorkflowData{
		ID: "test-id",
		Workflow: event.Workflow{
			ID:          "test-wf-id",
			Schema:      "test-type",
			Step:        "test-task",
			StepPayload: expPayload,
		},
		Header: expHeader,
		Debug:  true,
	}

	assert.Equal(t, "test-id", e.GetID())
	assert.Equal(t, "test-wf-id", e.GetWorkflow().ID)
	assert.Equal(t, "test-type", e.GetWorkflow().Schema)
	assert.Equal(t, "test-task", e.GetWorkflow().Step)
	assert.Equal(t, expHeader, e.GetHeader())
	assert.Equal(t, expPayload, []byte(e.GetWorkflow().StepPayload))
	assert.Equal(t, true, e.GetDebug())
	assert.NotNil(t, e.ToByte())
	assert.Equal(t, true, len(e.ToByte()) > 0)
}

func TestWorkflowEventUnmarshal(t *testing.T) {
	t.Parallel()

	expPayload := map[string]interface{}{"key": "val"}
	expPayloadBytes, _ := json.Marshal(expPayload)

	expectedData := map[string]interface{}{
		"id":     "test-id",
		"header": expHeader,
		"workflow": map[string]interface{}{
			"id":           "test-wf-id",
			"schema":       "test-type",
			"step":         "test-task",
			"step_payload": json.RawMessage(expPayloadBytes),
		},
	}

	b, err := json.Marshal(expectedData)

	require.NoError(t, err)

	msg := event.Message{
		Value: b,
	}
	e := event.WorkflowData{}
	err = e.Unmarshal(msg)
	require.NoError(t, err)

	assert.Equal(t, expectedData["id"], e.GetID())
	assert.Equal(t, expectedData["header"], e.GetHeader())

	assert.Equal(t, expectedData["workflow"].(map[string]interface{})["id"], e.GetWorkflow().ID)
	assert.Equal(t, expectedData["workflow"].(map[string]interface{})["schema"], e.GetWorkflow().Schema)
	assert.Equal(t, expectedData["workflow"].(map[string]interface{})["step"], e.GetWorkflow().Step)
	assert.EqualValues(t, expPayloadBytes, []byte(e.GetWorkflow().StepPayload))
}

func TestWorkflowEventError(t *testing.T) {
	t.Parallel()

	msg := event.Message{
		Value: []byte(`{"num":6.13,"strs":{"a","b"}}`),
	}

	e := &event.WorkflowData{}
	err := e.Unmarshal(msg)
	require.Error(t, err)
}
