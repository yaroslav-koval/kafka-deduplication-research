package main

import (
	"fmt"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	broker = "localhost:9093"
	topic  = "polygon1"
)

func main() {
	log.SetGlobalLogLevel("trace")

	config := kafka.ConfigMap{
		"bootstrap.servers":  broker,
		"enable.idempotence": true,
		"acks":               "all",
		//"retries":                               10, //https://www.linkedin.com/pulse/kafka-idempotent-producer-rob-golder/#:~:text=The%20recommendation%20is%20to%20leave%20the%20retries%20as%20the%20default%20(the%20maximum%20integer%20value)
		"retry.backoff.ms":                      500,
		"delivery.timeout.ms":                   3000,
		"max.in.flight.requests.per.connection": 5,
	}

	producer, err := kafka.NewProducer(&config)
	if err != nil {
		fmt.Println(err)
	}
	defer producer.Close()

	producer2, err := kafka.NewProducer(&config)
	if err != nil {
		fmt.Println(err)
	}
	defer producer2.Close()

	// Produce two messages with the same key
	key := "my-key"
	value1 := "my-value-1"
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: converto.StringPointer(topic), Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value1),
	}
	err = producer.Produce(message, nil)
	if err != nil {
		fmt.Println(err)
	}
	err = producer2.Produce(message, nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Produced messages successfully!")

	//producer.Flush(10 * 1000)
	//producer.Close()
}
