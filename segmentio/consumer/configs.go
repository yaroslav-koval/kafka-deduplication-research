package main

import (
	"kafka-polygon/pkg/broker/provider/kafka"
	storeRedis "kafka-polygon/pkg/broker/store/redis"
	"time"

	"github.com/go-redis/redis/v8"
)

func newKafkaProducerConsumerConfig() *kafka.Config {
	kafkaC := newKafkaProducerConfig()

	d, _ := time.ParseDuration("10s")
	d2, _ := time.ParseDuration("-1s")

	kafkaC.Consumer = kafka.Consumer{
		GroupID:         "segmentio",
		MinBytes:        1000,
		MaxBytes:        1000000,
		MaxWait:         d,
		ReadLagInterval: d2,
		CommitOnError:   false,
		MaxAttempts:     3,
	}

	kafkaC.CommitOnErrorMessagesCount = 1

	return kafkaC
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
			WriterBatchSize:    100,
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

func getRedis() *storeRedis.Store {
	rc := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	d, _ := time.ParseDuration("86400s")

	return storeRedis.NewStore(rc, storeRedis.Settings{
		KeyPrefix: "",
		TTL:       d,
	})
}
