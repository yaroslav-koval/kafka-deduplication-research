package main

import (
	"context"
	"kafka-polygon/pkg/broker"
	"kafka-polygon/pkg/broker/provider/kafka"
	"kafka-polygon/pkg/log"
	"sync"
	"time"
)

const (
	topic = "polygon1"
)

var (
	consumer broker.QueueBroker
)

type ConsumerData struct {
	consumer broker.QueueBroker
	index    int
}

func main() {
	ctx := context.Background()
	log.SetGlobalLogLevel("trace")

	if err := SetUpLogger(); err != nil {
		return
	}
	defer logger.Close(ctx)

	countOfMessages := 200
	i := 0

	consumersToStop := []*ConsumerData{}

	for {
		time.Sleep(2 * time.Second)
		_, err := writeMessages(ctx, topic, countOfMessages)
		if err != nil {
			log.Info(context.Background(), err.Error())
		}

		SetMessagesIndicator(countOfMessages)

		runConsumer(ctx, i)

		<-ConsCh

		i++

		if i > 4 {
			break
		}

		consumersToStop = append(consumersToStop, &ConsumerData{
			consumer: consumer,
			index:    i,
		})
		//consumer.Stop()
	}

	var wg sync.WaitGroup
	for _, c := range consumersToStop {
		wg.Add(1)
		go func(c *ConsumerData) {
			c.consumer.Stop()
			wg.Done()
		}(c)
	}

	//wg.Wait()
}

func runConsumer(ctx context.Context, i int) {
	logger.LogF("Consumer started: %s", time.Now()).InConsole(ctx)

	kafkaC := kafka.NewKafkaProvider(newKafkaProducerConsumerConfig())
	consumer = broker.New(kafkaC)
	redis := getRedis()

	consumer.SetStore(redis)

	consumer.Watch(ctx, topic, &SimpleHandler{ConsumerNumber: i})
}
