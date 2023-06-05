package main

import (
	"context"
	"fmt"
	storeRedis "kafka-polygon/pkg/broker/store/redis"
	"kafka-polygon/pkg/log"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

const (
	topic   = "polygon2"
	groupID = "sarama"
	brokers = "localhost:9093"
)

type ConsumerData struct {
	consumer sarama.ConsumerGroup
	index    int
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log.SetGlobalLogLevel("trace")

	redisStore, redisClient := getRedis()
	defer func() {
		err := redisClient.Close()
		if err != nil {
			logger.LogF("Error on closing redis: %v", err).InConsole(ctx)
		}
	}()

	if err := SetUpLogger(); err != nil {
		logger.LogF(err.Error()).InConsole(ctx)
		return
	}
	defer logger.Close(ctx)

	consumers := []*ConsumerData{}
	i := 0

	var wg sync.WaitGroup

	for {
		consumer := runConsumer(ctx, i, wg, redisStore)

		consumers = append(consumers, &ConsumerData{
			consumer: consumer,
			index:    i,
		})

		i++

		if i > 10 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	time.Sleep(2 * time.Second)
	logger.LogF("Consumption cancelling...").InConsole(ctx)

	cancel()

	logger.LogF("Consumption cancelled").InConsole(ctx)
	wg.Wait()

	logger.LogF("Consumption stopped").InConsole(ctx)

	wg.Add(len(consumers))

	for _, c := range consumers {
		go func(c *ConsumerData) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					msg := fmt.Sprintf("Recovered in f %s", r)
					logger.LogF(msg).InConsole(ctx)
				}
			}()

			if err := c.consumer.Close(); err != nil {
				logger.LogF("Error closing client: %v", err).InConsole(ctx)
			}
		}(c)
	}

	wg.Wait()
}

func runConsumer(ctx context.Context, i int, wg sync.WaitGroup, redisStore *storeRedis.Store) sarama.ConsumerGroup {
	wg.Add(1)

	cfg := newSaramaConsumerConfig()
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), groupID, cfg)
	if err != nil {
		logger.LogF("Error creating consumer group client: %v", err).InConsole(ctx)
	}

	consumer := Consumer{
		Ready:   make(chan bool),
		CTX:     ctx,
		Redis:   redisStore,
		Handler: &SimpleHandler{ConsumerNumber: i},
	}

	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(topic, ","), &consumer); err != nil {
				logger.LogF("Error from consumer: %v", err).InConsole(ctx)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.Ready = make(chan bool)
		}
	}()

	<-consumer.Ready
	logger.LogF("Consumer started: %d", i).InConsole(ctx)

	return client
}
