package main

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	brokerList := "localhost:9093"

	config := kafka.ConfigMap{
		"bootstrap.servers":                     brokerList,
		"enable.idempotence":                    true,
		"transactional.id":                      "my-transactional-id",
		"transaction.timeout.ms":                5000,
		"max.in.flight.requests.per.connection": 1,
	}

	producer, err := kafka.NewProducer(&config)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	err = producer.InitTransactions(nil)
	if err != nil {
		log.Fatal(err)
	}

	err = producer.BeginTransaction()
	if err != nil {
		log.Fatal(err)
	}

	topic := "polygon2"
	value := "Hello, Kafka!"
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(value),
	}
	err = producer.Produce(message, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = producer.CommitTransaction(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Produced transactional message successfully!")
}
