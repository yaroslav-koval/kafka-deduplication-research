package main

import (
	"context"
	"kafka-polygon/pkg/broker/provider/kafka"
	"kafka-polygon/pkg/cerror"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
	goKafka "github.com/segmentio/kafka-go"
)

func writeMessages(ctx context.Context, topicName string, countOfMessages int) (int, error) {
	producer := kafka.NewOpenClient(newKafkaProducerConfig())
	defer producer.Stop()

	msgs := []goKafka.Message{}

	for i := 0; i < countOfMessages; i++ {
		messageID := uuid.NewV4()
		//key := []byte(fmt.Sprintf("%q", messageID.String()+"-"+strconv.Itoa(i)))
		key := []byte(messageID.String() + "-" + strconv.Itoa(i))

		msgs = append(msgs, goKafka.Message{
			Topic: topicName,
			Key:   key,
			Value: key,
		})
	}

	err := SendRetry(ctx, producer, topicName, msgs, true)
	if err != nil {
		return 0, err
	}

	return countOfMessages, nil
}

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

func newKafkaProducerConfig() *kafka.Config {
	d, _ := time.ParseDuration("5s")
	d2, _ := time.ParseDuration("50ms")
	d3, _ := time.ParseDuration("3s")

	return &kafka.Config{
		Brokers:                []string{"localhost:9093"},
		LoggerEnabled:          true,
		UseKeyDoubleQuote:      true,
		AllowAutoTopicCreation: false,
		Producer: kafka.Producer{
			MaxAttempts:        3,
			MaxRetry:           10,
			MaxAttemptsDelay:   d,
			WriteTimeout:       d2,
			WriterBatchSize:    1000,
			WriterBatchTimeout: d3,
		},
		TLS: kafka.TLS{
			Enabled:            false,
			InsecureSkipVerify: false,
			ClientCertFile:     "",
			ClientKeyFile:      "",
			RootCACertFile:     "",
		},
		SASL: kafka.SASL{
			SecurityProtocol: "",
			Algorithm:        "",
			Username:         "",
			Password:         "",
		},
	}
}
