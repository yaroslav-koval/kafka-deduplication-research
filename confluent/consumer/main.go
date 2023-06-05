package main

import (
	"context"
	"kafka-polygon/pkg/broker/store/redis"
	"kafka-polygon/pkg/log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	topic  = "polygon2"
	broker = "localhost:9093"
)

type ConsumerData struct {
	consumer *kafka.Consumer
	index    int
}

func main() {
	ctx := context.Background()
	log.SetGlobalLogLevel("trace")

	redisStore := getRedis()

	errCh := make(chan error)

	consumers := []*ConsumerData{}

	i := 0

	for {
		c := runConsumer(ctx, redisStore, i, errCh)

		consumers = append(consumers, &ConsumerData{
			consumer: c,
			index:    i,
		})

		i++

		if i > 3 {
			break
		}

		time.Sleep(2 * time.Second)
	}

	<-AtLeastOneMessageChannel
	closeResources(ctx, consumers)
	return
}

func runConsumer(ctx context.Context, rs *redis.Store, i int, errCh chan error) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(newConfluentConsumerConfig())
	if err != nil {
		log.Debug(ctx, err.Error())
	}

	log.DebugF(ctx, "Consumer started: %d", i)

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Debug(ctx, err.Error())
	}

	go HandleMessage(ctx, consumer, &SimpleHandler{Redis: rs, ConsumerNumber: i}, errCh)

	return consumer
}

func closeResources(ctx context.Context, consumers []*ConsumerData) {
	for _, c := range consumers {
		_ = c.consumer.Close()
		log.DebugF(ctx, "Consumer %d closed", c.index)
	}
}
