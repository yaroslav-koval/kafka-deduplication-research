package test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/broker/errorinterceptor"
	"kafka-polygon/pkg/broker/errorinterceptor/entity"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/store"
	"testing"

	"github.com/tj/assert"
)

func TestHandleOriginalFn(t *testing.T) {
	t.Parallel()

	var calledFns []string

	mof := &mockOriginalFn{
		getEventDataFunc: func(_ context.Context) event.BaseEvent {
			calledFns = append(calledFns, "getEventDataFunc")
			return &entity.ErrorEvent{}
		},
		callFnFunc: func(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
			calledFns = append(calledFns, "callFnFunc")
			return nil
		},
	}

	es := errorinterceptor.NewErrorStore(&mockStoreProvider{
		saveEventFunc: func(ctx context.Context, e *entity.FailedBrokerEvent) (*entity.FailedBrokerEvent, error) {
			calledFns = append(calledFns, "saveEventFunc")
			return &entity.FailedBrokerEvent{}, nil
		},
	})

	h := es.HandleOriginalFn(mof, &Event{}, "topic")

	ed := h.GetEventData(_cb)
	ev, ok := ed.(*entity.ErrorEvent)
	assert.True(t, ok)

	_, ok = ev.OriginalEvent.(*Event)
	assert.True(t, ok)

	err := h.CallFn(_cb, ed, store.EventProcessData{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"callFnFunc"}, calledFns)

	calledFns = []string{}

	mof.callFnFunc = func(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
		calledFns = append(calledFns, "callFnFunc")
		return fmt.Errorf("error")
	}

	err = h.CallFn(_cb, ed, store.EventProcessData{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"callFnFunc", "saveEventFunc"}, calledFns)
}
