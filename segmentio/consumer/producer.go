package main

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider/kafka"
	"kafka-polygon/pkg/cerror"
	"strconv"

	uuid "github.com/satori/go.uuid"
	goKafka "github.com/segmentio/kafka-go"
)

func writeMessages(ctx context.Context, topic string, countOfMessages int) (int, error) {
	producer := kafka.NewOpenClient(newKafkaProducerConfig())
	defer producer.Stop()

	msgs := []goKafka.Message{}

	for i := 0; i < countOfMessages; i++ {
		messageID := uuid.NewV4()
		key := []byte(fmt.Sprintf("%q", messageID.String()+"-"+strconv.Itoa(i)))

		wd := &event.WorkflowData{
			ID: messageID.String(),
			Header: event.Header{
				RequestID: messageID.String(),
			},
			Workflow: event.Workflow{
				ID:          messageID.String(),
				Schema:      "producer_schema",
				Step:        topic,
				StepPayload: key,
			},
		}

		msgs = append(msgs, goKafka.Message{
			Topic: topic,
			Key:   key,
			Value: wd.ToByte(),
		})
	}

	err := SendRetry(ctx, producer, topic, msgs, true)
	if err != nil {
		return 0, err
	}

	return countOfMessages, nil
}

// SendRetry is made to be able to send batch of messages
func SendRetry(ctx context.Context, c *kafka.OpenClient, topic string, msgs []goKafka.Message, retry bool) error {
	tr := c.GetTransport(context.Background())
	defer tr.CloseIdleConnections()

	wr := c.NewWriter(ctx, tr)
	defer func() {
		errWR := wr.Close()
		if errWR != nil {
			_ = cerror.NewF(ctx,
				cerror.KafkaToKind(errWR),
				"sendMessage topic: %s. %s", topic, errWR.Error()).LogError()
		}
	}()

	err := wr.WriteMessages(context.Background(), msgs...)
	if err != nil {
		if retry && c.IsCheckRetry(err) {
			_ = cerror.NewF(ctx,
				cerror.KafkaToKind(err),
				"sendMessage topic: %s. Trying to retry. %s", topic, err.Error()).
				LogError()

			return c.ProcessRetry(ctx, wr, topic, msgs)
		}

		return cerror.NewF(
			ctx, cerror.KafkaToKind(err), "sendMessage topic: %s. %s", topic, err.Error()).
			LogError()
	}

	return nil
}
