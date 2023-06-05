package errorinterceptor

import (
	"context"
	"kafka-polygon/pkg/broker/errorinterceptor/entity"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider"
	"kafka-polygon/pkg/broker/store"
)

type Store interface {
	SaveEvent(ctx context.Context, e *entity.FailedBrokerEvent) (*entity.FailedBrokerEvent, error)
}

type ErrorInterceptor struct {
	store Store
}

type ErrorHandler struct {
	originalEvent       event.BaseEvent
	errorStoreHandlerFn func(context.Context, *entity.ErrorEvent, store.EventProcessData) error
}

func NewErrorStore(st Store) *ErrorInterceptor {
	return &ErrorInterceptor{
		store: st,
	}
}

func (ei *ErrorInterceptor) HandleOriginalFn(originalFn provider.HandlerFn, e event.BaseEvent, topic string) provider.HandlerFn {
	fn := func(ctx context.Context, event *entity.ErrorEvent, eventData store.EventProcessData) error {
		err := originalFn.CallFn(ctx, event.OriginalEvent, eventData)
		if err != nil {
			_, saveErr := ei.store.SaveEvent(ctx, &entity.FailedBrokerEvent{
				ID:     event.OriginalEvent.GetID(),
				Topic:  topic,
				Status: entity.StatusNew,
				Data:   event.OriginalEvent.ToByte(),
				Error:  err.Error(),
			})
			if saveErr != nil {
				return err
			}
		}

		return nil
	}

	return &ErrorHandler{
		originalEvent:       e,
		errorStoreHandlerFn: fn,
	}
}

func (eh ErrorHandler) GetEventData(_ context.Context) event.BaseEvent {
	return entity.NewErrorEvent(eh.originalEvent)
}

func (eh ErrorHandler) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	if ed, ok := e.(*entity.ErrorEvent); ok {
		return eh.errorStoreHandlerFn(reqCtx, ed, eventData)
	}

	return nil
}
