package testutil

import (
	"context"
	"errors"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/testutil/config"

	"github.com/go-zookeeper/zk"
	"github.com/ory/dockertest/v3/docker"
)

type DockerZKContainer struct {
	*DockerTestContainer
}

func NewDockerZKContainer(cfg *config.TestContainerConfig) *DockerZKContainer {
	c := &DockerZKContainer{
		DockerTestContainer: &DockerTestContainer{
			Name:                GenerateCorrectName("zookeeper"),
			Repository:          "wurstmeister/zookeeper",
			Tag:                 "3.4.6",
			Port:                "2181",
			TestContainerConfig: cfg,
			Host:                "localhost",
			activeConn:          0,
		},
	}
	c.Env = []string{
		"ALLOW_ANONYMOUS_LOGIN=yes",
	}

	return c
}

func (c *DockerZKContainer) GetDockerRunOptions() *dockertest.RunOptions {
	opts := c.DockerTestContainer.GetDockerRunOptions()
	opts.Hostname = "zookeeper"
	opts.PortBindings = map[docker.Port][]docker.PortBinding{
		docker.Port(fmt.Sprintf("%s/tcp", c.Port)): {{
			HostIP:   "0.0.0.0",
			HostPort: fmt.Sprintf("%s/tcp", c.Port),
		}},
	}
	opts.ExposedPorts = []string{c.Port}
	opts.Privileged = true

	return opts
}

func (c *DockerZKContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerZKContainer {
	c.DockerTestContainer.RunWithOptions(opt)
	return c
}

func (c *DockerZKContainer) Connect() *DockerZKContainer {
	ctx := context.Background()

	conn, _, err := zk.Connect([]string{fmt.Sprintf("127.0.0.1:%s", c.Port)}, defaultTimeout)
	if err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not connect zookeeper: %s", err).
			LogFatal()
		return nil
	}
	defer conn.Close()

	if err := c.Pool.Retry(func() error {
		switch conn.State() {
		case zk.StateHasSession, zk.StateConnected:
			return nil
		default:
			return errors.New("not yet connected")
		}
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not connect to zookeeper in docker with port %v", c.Port).
			LogFatal()
		return nil
	}
	c.activeConn++

	return c
}

func (c *DockerZKContainer) Close() *DockerZKContainer {
	c.closeConnection()

	return c
}
