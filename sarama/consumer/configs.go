package main

import (
	storeRedis "kafka-polygon/pkg/broker/store/redis"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
)

func newSaramaConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	//config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	//config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false

	return config
}

func getRedis() (*storeRedis.Store, *redis.Client) {
	rc := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	d, _ := time.ParseDuration("86400s")

	return storeRedis.NewStore(rc, storeRedis.Settings{
		KeyPrefix: "",
		TTL:       d,
	}), rc
}
