package testutil

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/testutil/config"

	"github.com/go-redis/redis/v8"
)

type DockerRedisContainer struct {
	*DockerTestContainer
	rcClient *redis.Client
}

func NewDockerRedisContainer(cfg *config.TestContainerConfig) *DockerRedisContainer {
	c := &DockerRedisContainer{
		DockerTestContainer: &DockerTestContainer{
			Name:                GenerateCorrectName("redis"),
			Repository:          "redis",
			Tag:                 "latest",
			Port:                "6379",
			TestContainerConfig: cfg,
			Host:                "localhost",
			activeConn:          0,
		},
	}

	return c
}

func (c *DockerRedisContainer) GetRedisEnvConfig() *env.Redis {
	return &env.Redis{
		Host:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Password: "",
		DBNumber: 0,
	}
}

func (c *DockerRedisContainer) GetDockerRunOptions() *dockertest.RunOptions {
	return c.DockerTestContainer.GetDockerRunOptions()
}

func (c *DockerRedisContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerRedisContainer {
	c.DockerTestContainer.RunWithOptions(opt)
	return c
}

func (c *DockerRedisContainer) ConnectToRedis() *DockerRedisContainer {
	ctx := context.Background()
	rcEnv := c.GetRedisEnvConfig()

	if c.rcClient != nil {
		return c
	}

	var rcClient *redis.Client

	if err := c.Pool.Retry(func() error {
		var err error
		rcClient = redis.NewClient(&redis.Options{
			Addr:     rcEnv.Host,
			Password: rcEnv.Password,
			DB:       rcEnv.DBNumber,
		})

		_, err = rcClient.Ping(ctx).Result()
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not connect to redis in docker with port %v", c.Port).
			LogFatal()
		return nil
	}

	c.rcClient = rcClient
	c.activeConn++

	return c
}

func (c *DockerRedisContainer) GetClient() *redis.Client {
	return c.rcClient
}

func (c *DockerRedisContainer) Close() {
	c.closeConnection()
}
