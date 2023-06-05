package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9093",
		"group.id":          "my-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	topic := "polygon2"
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatal(err)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go handler(consumer)

	select {
	case sig := <-sigchan:
		fmt.Printf("Caught signal %v: terminating\n", sig)
		return
	}
}

func handler(consumer *kafka.Consumer) {
	for {
		ev := consumer.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			fmt.Printf("Received message on topic %s: %s\n", *e.TopicPartition.Topic, string(e.Value))
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "Error: %v\n", e)
			if e.Code() == kafka.ErrAllBrokersDown {
				// If all brokers are down, exit the application
				panic(e)
			}
		default:
			fmt.Printf("Ignored event: %s\n", ev)
		}

		time.Sleep(1 * time.Second)
	}
}
