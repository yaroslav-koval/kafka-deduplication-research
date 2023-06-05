package main

import (
	storeRedis "kafka-polygon/pkg/broker/store/redis"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
)

func newConfluentConsumerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          "confluent",
		"auto.offset.reset": "earliest",
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
