package testutil

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/testutil/config"
	"strings"
	"time"

	"github.com/ory/dockertest/v3/docker"
	"github.com/segmentio/kafka-go"
)

type DockerKafkaContainer struct {
	*DockerTestContainer
	param KafkaParams
}

func NewDockerKafkaContainer(cfg *config.TestContainerConfig, param KafkaParams) *DockerKafkaContainer {
	c := &DockerKafkaContainer{
		DockerTestContainer: &DockerTestContainer{
			Name:                GenerateCorrectName("kafka"),
			Repository:          "wurstmeister/kafka",
			Tag:                 "2.12-2.3.0",
			Port:                "9093",
			TestContainerConfig: cfg,
			Host:                "localhost",
			activeConn:          0,
		},
		param: param,
	}

	defTopics := []string{"demo_topic:1:1:compact"}
	defTopics = append(defTopics, param.DefaultTopic...)

	c.Env = []string{
		fmt.Sprintf("KAFKA_CREATE_TOPICS=%s", strings.Trim(strings.Join(defTopics, ","), ",")),
		"KAFKA_ADVERTISED_LISTENERS=INSIDE://localhost:9092,OUTSIDE://localhost:9093",
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT",
		"KAFKA_LISTENERS=INSIDE://0.0.0.0:9092,OUTSIDE://0.0.0.0:9093",
		fmt.Sprintf("KAFKA_ZOOKEEPER_CONNECT=%s:%s", param.ZKHost, param.ZKPort),
		"KAFKA_INTER_BROKER_LISTENER_NAME=INSIDE",
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
		"KAFKA_AUTO_CREATE_TOPICS_ENABLE=true",
		"KAFKA_BROKER_ID=1",
	}

	return c
}

func (c *DockerKafkaContainer) GetDockerRunOptions() *dockertest.RunOptions {
	opts := c.DockerTestContainer.GetDockerRunOptions()
	opts.Hostname = "kafka"
	opts.Links = []string{c.param.ZKHost}

	opts.ExposedPorts = []string{c.Port}
	opts.PortBindings = map[docker.Port][]docker.PortBinding{
		docker.Port(c.Port): {{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%s/tcp", c.Port)}},
	}
	opts.Privileged = true

	return opts
}

func (c *DockerKafkaContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerKafkaContainer {
	c.DockerTestContainer.RunWithOptions(opt)
	return c
}

func (c *DockerKafkaContainer) Connect() *DockerKafkaContainer {
	ctx := context.Background()

	if err := c.Pool.Retry(func() error {
		dial := &kafka.Dialer{
			Timeout:   defaultTimeout,
			DualStack: true,
		}

		topic := "demo_topic"
		addr := kafka.TCP([]string{fmt.Sprintf("localhost:%s", c.Port)}...)
		conn, err := dial.DialContext(ctx, addr.Network(), addr.String())
		if err != nil {
			return fmt.Errorf("failed to dial leader: %s", err)
		}

		err = conn.SetRequiredAcks(-1)
		if err != nil {
			return fmt.Errorf("failed to SetRequiredAcks: %s", err)
		}

		err = conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			return fmt.Errorf("failed to SetWriteDeadline: %s", err)
		}

		logf := func(msg string, a ...interface{}) {
			fmt.Printf(msg, a...)
			fmt.Println()
		}

		w := kafka.NewWriter(kafka.WriterConfig{
			Brokers:      []string{fmt.Sprintf("localhost:%s", c.Port)},
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			Dialer:       dial,
			RequiredAcks: -1,
			Logger:       kafka.LoggerFunc(logf),
			ErrorLogger:  kafka.LoggerFunc(logf),
		})
		w.AllowAutoTopicCreation = true

		err = w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte("Key-A"),
				Value: []byte("Hello World!"),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to write messages: %s", err)
		}

		if err := w.Close(); err != nil {
			return fmt.Errorf("failed to close writer: %s", err)
		}

		if err := conn.Close(); err != nil {
			return fmt.Errorf("failed to close writer: %s", err)
		}

		return nil
	}); err != nil {
		_ = cerror.NewF(ctx,
			cerror.KindInternal,
			"could not connect to kafka in docker with port %v => err: %s", c.Port, err.Error()).
			LogFatal()

		return nil
	}
	c.activeConn++

	return c
}

func (c *DockerKafkaContainer) Close() *DockerKafkaContainer {
	c.closeConnection()

	return c
}
