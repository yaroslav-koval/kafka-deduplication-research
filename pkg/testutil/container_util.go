package testutil

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/testutil/config"

	dc "github.com/ory/dockertest/v3/docker"
)

type DockerTestContainer struct {
	*config.TestContainerConfig
	Name       string
	Tag        string
	Repository string
	Port       string
	Test       string
	Env        []string
	Host       string
	resource   *dockertest.Resource
	activeConn int
}

func (c *DockerTestContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerTestContainer {
	ctx := context.Background()
	if err := c.Pool.Retry(func() error {
		resource, ok := c.Pool.ContainerByName(c.Name)
		if !ok {
			var err error
			resource, err = c.Pool.RunWithOptions(opt, func(config *dc.HostConfig) {
				config.AutoRemove = true
				config.RestartPolicy = dc.RestartPolicy{
					Name: "no",
				}
			})
			if err != nil {
				return err
			}
		}
		c.resource = resource
		c.Port = resource.GetPort(c.GetPortBinding())
		log.Log(logger.NewEventF(ctx, logger.LevelDebug, "current port = %v", c.Port))
		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not start %s container with name %v", c.Repository, c.Name).LogFatal()
		return nil
	}

	return c
}

func (c *DockerTestContainer) GetCurrPort(id string) string {
	return c.resource.GetPort(c.GetPortBindingByPort(id))
}

func (c *DockerTestContainer) GetDockerRunOptions() *dockertest.RunOptions {
	opt := &dockertest.RunOptions{
		Name:         c.Name,
		Repository:   c.Repository,
		Tag:          c.Tag,
		ExposedPorts: []string{c.Port},
		Env:          c.Env,
	}

	if c.IsDefaultPortUses {
		opt.PortBindings = map[dc.Port][]dc.PortBinding{
			dc.Port(c.GetPortBinding()): {{HostIP: "", HostPort: c.Port}},
		}
	}

	return opt
}

func (c *DockerTestContainer) GetPortBinding() string {
	return c.GetPortBindingByPort(c.Port)
}

func (c *DockerTestContainer) GetPortBindingByPort(port string) string {
	return fmt.Sprintf("%v/tcp", port)
}

func (c *DockerTestContainer) RunWithPoolRetry(fn func() error) error {
	return c.Pool.Retry(fn)
}

func (c *DockerTestContainer) GetResourcesInfo() *DockerTestContainer {
	return c
}

func (c *DockerTestContainer) closeConnection() {
	c.activeConn--
	if c.activeConn == 0 {
		_ = c.resource.Close()
	}
}

func GenerateCorrectName(name string) string {
	return fmt.Sprintf(defaultTemplateContainerName, name)
}
