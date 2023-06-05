package provider_test

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider"
	"kafka-polygon/pkg/broker/store"
	"testing"

	"github.com/tj/assert"
)

var _bgCtx = context.Background()

type mockedStore struct {
	t      *testing.T
	expErr error
}

func (ms *mockedStore) GetEventInfoByID(_ context.Context, id string) (store.EventProcessData, error) {
	assert.Equal(ms.t, e.ID, id)
	return store.EventProcessData{Status: store.EventStatusNew}, ms.expErr
}

func (ms *mockedStore) PutEventInfo(_ context.Context, id string, _ store.EventProcessData) error {
	assert.Equal(ms.t, e.ID, id)

	return ms.expErr
}

var (
	buff buffWriter
	e    = event.WorkflowData{
		ID: "test-id",
		Workflow: event.Workflow{
			ID:          "test-wf-id",
			Schema:      "test-type",
			Step:        "test-task",
			StepPayload: json.RawMessage("{}"),
		},
		Header: event.Header{
			RequestID: "test-x-request-id",
		},
	}
)

func TestHandlerProcessing(t *testing.T) {
	t.Parallel()

	ms := &mockedStore{t: t}
	fn := provider.HandlerWorkflow(func(ctx context.Context, workflowEvent event.WorkflowEvent, ed store.EventProcessData) error {
		_, err := buff.Write(workflowEvent.ToByte())
		assert.NoError(t, err)
		return nil
	})

	sendMsg := event.Message{
		Value: e.ToByte(),
	}

	hp := provider.NewHandlerProcessing(ms)
	_, err := hp.Run(_bgCtx, fn, sendMsg)
	assert.NoError(t, err)
	assert.Equal(t, string(e.ToByte()), buff.String())
}

func TestHandlerProcessingError(t *testing.T) {
	t.Parallel()

	listErr := []struct {
		storeErr, fnErr error
	}{
		{
			storeErr: errEmpty,
			fnErr:    nil,
		},
		{
			storeErr: nil,
			fnErr:    errEmpty,
		},
	}

	for _, val := range listErr {
		ms := &mockedStore{t: t, expErr: val.storeErr}
		fn := provider.HandlerWorkflow(func(ctx context.Context, workflowEvent event.WorkflowEvent, ed store.EventProcessData) error {
			return val.fnErr
		})

		sendMsg := event.Message{
			Value: e.ToByte(),
		}

		hp := provider.NewHandlerProcessing(ms)
		_, err := hp.Run(_bgCtx, fn, sendMsg)
		assert.Error(t, err)
		assert.Equal(t, errEmpty.Error(), err.Error())
	}
}
