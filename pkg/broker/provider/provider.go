package provider

import (
	"context"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/tracing"

	uuid "github.com/satori/go.uuid"
)

type Provider interface {
	GetType() string
	SetEnabled(enable bool)
	SetStore(s store.Store)
	SetTracing(t tracing.Tracer)
	GetIsTopicExists(ctx context.Context, topic string) (bool, error)
	Publish(ctx context.Context, topic string, e event.BaseEvent) error
	Sync(ctx context.Context, topic string, fn HandlerFn)
	Stop()
}

type HandlerProcessing struct {
	store store.Store
}

func NewHandlerProcessing(s store.Store) *HandlerProcessing {
	return &HandlerProcessing{
		store: s,
	}
}

func (hp *HandlerProcessing) Run(ctx context.Context, fn interface{}, msg event.Message) (event.BaseEvent, error) {
	f, ok := fn.(HandlerFn)
	if !ok {
		return nil, cerror.NewF(ctx, cerror.KindInternal, "not supporting type HandlerFn %T", fn).LogError()
	}

	e := f.GetEventData(ctx)

	err := e.Unmarshal(msg)
	if err != nil {
		return nil, cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	ctx = hp.ctxWithRequestID(ctx, e.GetHeader().RequestID)

	var (
		eventData store.EventProcessData
		sErr      error
	)

	if hp.store != nil {
		eventData, sErr = hp.store.GetEventInfoByID(ctx, e.GetID())
		if sErr != nil {
			if !cerror.IsNotExist(sErr) {
				return nil, sErr
			}

			eventData.Status = store.EventStatusNew

			sErr = hp.store.PutEventInfo(ctx, e.GetID(), eventData)
			if sErr != nil {
				return nil, sErr
			}
		}
	}

	newStatus := store.EventStatusHandled

	err = f.CallFn(ctx, e, eventData)
	if err != nil {
		newStatus = store.EventStatusHandledWithError
	}

	if hp.store != nil && eventData.Status != newStatus {
		eventData.Status = newStatus

		sErr = hp.store.PutEventInfo(ctx, e.GetID(), eventData)
		if sErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"couldn't update event status. event_id=%s. old_status=%s. new_status=%s. error=%s",
				e.GetID(), eventData.Status, newStatus, sErr.Error()).LogError()
		}
	}

	return e, err
}

// ctxWithRequestID creates a new context from a given context
// and adds requestID value to it.
// If requestID is empty it will generate a new one.
func (hp *HandlerProcessing) ctxWithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = uuid.NewV4().String()
	}

	return context.WithValue(ctx, consts.HeaderXRequestID, requestID) //nolint:staticcheck
}
