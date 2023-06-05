package main

import (
	"context"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/broker/store/redis"
	"kafka-polygon/pkg/cerror"
	"time"

	"github.com/Shopify/sarama"
)

type SimpleHandler struct {
	ConsumerNumber int
}

type Consumer struct {
	Ready   chan bool
	CTX     context.Context
	Redis   *redis.Store
	Handler *SimpleHandler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.Ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message := <-claim.Messages():
			messageID := string(message.Key)
			logger.LogF("Message claimed: key = %s, value = %s, topic = %s, partition = %d, offset = %d",
				messageID, string(message.Value), message.Topic, message.Partition, message.Offset)

			eventData, sErr := consumer.Redis.GetEventInfoByID(consumer.CTX, messageID)
			if sErr != nil {
				if !cerror.IsNotExist(sErr) {
					logRedisError(consumer.CTX, messageID, consumer.Handler.ConsumerNumber, sErr)
					return sErr
				}

				eventData.Status = store.EventStatusNew

				sErr = consumer.Redis.PutEventInfo(consumer.CTX, messageID, eventData)
				if sErr != nil {
					logRedisError(consumer.CTX, messageID, consumer.Handler.ConsumerNumber, sErr)

					if sErr.Error() == "context canceled" {
						err := consumer.Redis.DeleteEventInfoByID(context.Background(), messageID)
						if err != nil {
							logRedisError(consumer.CTX, messageID, consumer.Handler.ConsumerNumber, err)
							return sErr
						}
					}
					return sErr
				}

			} else {
				eventData.Status = store.EventStatusHandled
				err := consumer.Redis.PutEventInfo(consumer.CTX, messageID, eventData)
				logRedisError(consumer.CTX, messageID, consumer.Handler.ConsumerNumber, err)
			}

			err := consumer.Handler.CallFn(consumer.CTX, messageID, eventData)
			if err != nil {
				return err
			}

			session.MarkMessage(message, "")
			session.Commit()

			//time.Sleep(100 * time.Millisecond)

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}

func (s *SimpleHandler) CallFn(reqCtx context.Context, messageID string, eventData store.EventProcessData) error {
	if eventData.Status == "handled" {
		logger.LogF("Duplication. Handled message received: %s. Key %s", time.Now(), messageID).InConsole(reqCtx)
	}

	logger.LogF("Consumer number %d. MessageID %s", s.ConsumerNumber, messageID) //.InConsole(reqCtx)

	DecrementMessagesIndicator()

	//if messagesLeft == 0 {
	//	ConsCh <- nil
	//}

	return nil
}

func logRedisError(ctx context.Context, messageID string, consumerNumber int, err error) {
	cErr, _ := err.(*cerror.CError)
	logger.LogF("Error from redis. MessageID %s. Consumer %d. %v. Ops: %v", messageID, consumerNumber, err, cErr.Ops())
}
