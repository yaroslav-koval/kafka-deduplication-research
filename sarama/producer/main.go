package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
)

var (
	topic = "polygon2"
)

func main() {
	// Kafka broker addresses
	brokerList := []string{"localhost:9093"}

	// Configure the Sarama producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Net.MaxOpenRequests = 1

	// Create the Kafka producer
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Produce a message
	value := "Hello, Kafka!"
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(value),
	}

	// Send the message
	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Produced message to topic %s, partition %d, offset %d\n", topic, partition, offset)

	// Graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan

	fmt.Println("Producer closed successfully!")
}
