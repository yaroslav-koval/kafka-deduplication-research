package config

import (
	"context"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"os"
	"strconv"
)

type TestContainerConfig struct {
	Pool              *dockertest.Pool
	IsDefaultPortUses bool
}

func (c *TestContainerConfig) FromEnv() {
	c.IsDefaultPortUses = false
	v, isExts := os.LookupEnv("USE_DEFAULT_PORT_IN_TESTS")

	log.Log(logger.NewEventF(context.Background(),
		logger.LevelDebug,
		"is env USE_DEFAULT_PORT_IN_TESTS exists = %v and value = %v", isExts, v))

	if isExts {
		if bVal, err := strconv.ParseBool(v); err != nil {
			c.IsDefaultPortUses = bVal
		}
	}
}
