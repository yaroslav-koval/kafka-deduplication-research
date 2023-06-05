package testutil

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/testutil/config"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type DockerElasticContainer struct {
	*DockerTestContainer
	esClient *elasticsearch.Client
}

func NewDockerElasticContainer(cfg *config.TestContainerConfig) *DockerElasticContainer {
	c := &DockerElasticContainer{
		DockerTestContainer: &DockerTestContainer{
			Name:                GenerateCorrectName("elasticsearch"),
			Repository:          "docker.elastic.co/elasticsearch/elasticsearch",
			Tag:                 "7.7.0",
			Port:                "9200",
			TestContainerConfig: cfg,
			Host:                "localhost",
			activeConn:          0,
		},
	}

	c.Env = []string{
		"http.port=9200",
		"cluster.name=docker-cluster",
		"discovery.type=single-node",
		"bootstrap.memory_lock=false",
		"ES_JAVA_OPTS=-Xms512m -Xmx512m",
	}

	return c
}

func (c *DockerElasticContainer) GetElasticEnvConfig() *env.Elastic {
	return &env.Elastic{
		Hosts: []string{fmt.Sprintf("http://%s:%s", c.Host, c.Port)},
	}
}

func (c *DockerElasticContainer) GetDockerRunOptions() *dockertest.RunOptions {
	return c.DockerTestContainer.GetDockerRunOptions()
}

func (c *DockerElasticContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerElasticContainer {
	c.DockerTestContainer.RunWithOptions(opt)
	return c
}

func (c *DockerElasticContainer) ConnectToElastic() *DockerElasticContainer {
	ctx := context.Background()
	endpoint := fmt.Sprintf("http://%s:%s", c.Host, c.Port)

	if c.esClient != nil {
		return c
	}

	var esClient *elasticsearch.Client

	if err := c.Pool.Retry(func() error {
		var err error
		esClient, err = elasticsearch.NewClient(
			elasticsearch.Config{
				Addresses: []string{endpoint},
			})
		if err != nil {
			return err
		}

		pr := esapi.PingRequest{}
		_, err = pr.Do(ctx, esClient)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not connect to elasticsearch in docker with port %v", c.Port).
			LogFatal()
		return nil
	}

	c.esClient = esClient
	c.activeConn++

	return c
}

func (c *DockerElasticContainer) GetClient() *elasticsearch.Client {
	return c.esClient
}

func (c *DockerElasticContainer) Close() {
	c.closeConnection()
}
