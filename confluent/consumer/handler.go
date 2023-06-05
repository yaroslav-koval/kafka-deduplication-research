package main

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/store"
	storeRedis "kafka-polygon/pkg/broker/store/redis"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type SimpleHandler struct {
	Redis          *storeRedis.Store
	ConsumerNumber int
}

func (s *SimpleHandler) GetEventData(_ context.Context) event.BaseEvent {
	return &event.WorkflowData{}
}

func HandleMessage(ctx context.Context, consumer *kafka.Consumer, h *SimpleHandler, errCh chan<- error) {
	for {
		ev := consumer.Poll(1000)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			fmt.Printf("Received message on topic %s: %s\n", *e.TopicPartition.Topic, string(e.Value))
			messageID := string(e.Key[1 : len(e.Key)-1])
			eventData, sErr := h.Redis.GetEventInfoByID(ctx, messageID)
			if sErr != nil {
				if !cerror.IsNotExist(sErr) {
					errCh <- cerror.NewF(ctx, cerror.KindInternal, "Got existed message with id %s", messageID)
					return
				}

				eventData.Status = store.EventStatusNew

				sErr = h.Redis.PutEventInfo(ctx, messageID, eventData)
				if sErr != nil {
					errCh <- cerror.NewF(ctx, cerror.KindInternal, "Couldn't put message with id %s", messageID)
					return
				}

			} else {
				eventData.Status = store.EventStatusHandled
			}
			h.CallFn(ctx, e.Value, eventData)

			_, err := consumer.CommitMessage(e)
			if err != nil {
				cerror.New(ctx, cerror.KindInternal, err)
			}

			time.Sleep(100 * time.Millisecond)
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "Error: %v\n", e)
			if e.Code() == kafka.ErrAllBrokersDown {
				// If all brokers are down, exit the application
				close(errCh)
			}
		default:
			fmt.Printf("Ignored event: %s\n", ev)
		}
	}
}

func (s *SimpleHandler) CallFn(reqCtx context.Context, e interface{}, eventData store.EventProcessData) error {
	if eventData.Status == "handled" {
		log.DebugF(reqCtx, "Handled message received: %s", time.Now())
	}

	log.DebugF(reqCtx, "Consumer number %d", s.ConsumerNumber)

	DecrementMessagesIndicator()

	go func() {
		// this is made if consumer needs more time to start consuming than we gave it by time.Sleep in for cycle in  main
		AtLeastOneMessageChannel <- nil
	}()

	return nil
}
