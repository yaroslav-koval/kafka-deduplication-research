package main

import (
	"context"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/store"
	"time"
)

type SimpleHandler struct {
	ConsumerNumber int
}

func (s *SimpleHandler) GetEventData(_ context.Context) event.BaseEvent {
	return &event.WorkflowData{}
}

func (s *SimpleHandler) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	if eventData.Status == "handled" {
		logger.LogF("Handled message received: %s", time.Now()).InConsole(reqCtx)
	}

	logger.LogF("Consumer number %d. Message value: %s", s.ConsumerNumber, e)

	DecrementMessagesIndicator()

	if messagesLeft == 0 {
		ConsCh <- nil
	}

	return nil
}
