package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	goKafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type OpenClient struct {
	errCntMu            sync.Mutex
	cfg                 *Config
	failedMessagesCount int
	wg                  sync.WaitGroup
	stop                bool
}

func NewOpenClient(cfg *Config) *OpenClient {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg.defaults()

	return &OpenClient{
		cfg: cfg,
	}
}

func (c *OpenClient) ListenTopic(ctx context.Context, topic string, handler MessageHandler) chan error {
	errCh := make(chan error)
	reader := c.NewReader(ctx, topic)

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		defer func() {
			log.DebugF(ctx, "reader for topic %s stopped and close", reader.Config().Topic)

			if err := reader.Close(); err != nil {
				_ = cerror.NewF(ctx,
					cerror.KafkaToKind(err),
					"[kafka] listenTopicByReaderConfig reader.Close error. %s", err.Error()).
					LogError()
			}
		}()

		log.DebugF(ctx, "start listening kafka topic [%s]", reader.Config().Topic)

		for {
			if c.stop {
				break
			}

			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					log.DebugF(ctx,
						"failed fetch message from kafka for topic: %v. %s",
						reader.Config().Topic, err.Error())

					break
				}

				cErr := cerror.NewF(
					ctx,
					cerror.KafkaToKind(err),
					"failed fetch message from kafka for topic: %v. %s", reader.Config().Topic, err.Error()).
					LogError()

				errCh <- cErr

				break
			}

			log.DebugF(ctx,
				"consume message from kafka topic: %s. partition: %d. offset: %d. key: %s.",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				string(msg.Key))

			e, err := handler.Handle(ctx, &msg)
			if err != nil {
				errCnt := c.incErrCnt()
				if !c.cfg.Consumer.CommitOnError || c.cfg.CommitOnErrorMessagesCount < errCnt {
					cErr := cerror.NewF(
						ctx,
						cerror.KafkaToKind(err),
						"failed process message in handler for topic: %v. key = %s. %s",
						msg.Topic, string(msg.Key), err.Error()).
						LogError()
					errCh <- cErr

					break
				}

				log.DebugF(ctx,
					"message was processed with error and skipped. total count of skipped messages: %v. allowed count: %v. key = %s",
					errCnt,
					c.cfg.CommitOnErrorMessagesCount,
					string(msg.Key))
			}

			ctxWithValues := context.WithValue(ctx, consts.HeaderXRequestID, e.GetHeader().RequestID) //nolint:staticcheck

			err = reader.CommitMessages(context.Background(), msg)
			if err != nil {
				cErr := cerror.NewF(
					ctxWithValues,
					cerror.KafkaToKind(err),
					"failed to commit message. topic: %s. partition %d. offset: %d. key: %s. value: %s. %s",
					msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value), err.Error()).
					LogError()
				errCh <- cErr

				break
			}
		}
	}()

	return errCh
}

func (c *OpenClient) GetIsTopicExists(ctx context.Context, topic string) (bool, error) {
	conn, err := c.GetConnection(ctx)
	if err != nil {
		return false, err
	}

	defer func() {
		errConn := conn.Close()
		if errConn != nil {
			_ = cerror.NewF(ctx,
				cerror.KafkaToKind(err),
				"[kafka] GetIsTopicExists wr.Close error: %s", errConn.Error()).
				LogError()
		}
	}()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false, cerror.NewF(ctx,
			cerror.KafkaToKind(err),
			"[kafka] GetIsTopicExists conn.ReadPartitions error: %s", err.Error()).
			LogError()
	}

	for i := range partitions {
		if partitions[i].Topic == topic {
			return true, nil
		}
	}

	return false, nil
}

func (c *OpenClient) SendMessage(ctx context.Context, topic string, e event.BaseEvent) error {
	key := e.GetID()
	if c.cfg.UseKeyDoubleQuote {
		key = fmt.Sprintf("%q", e.GetID())
	}

	return c.SendRetry(ctx, topic, []goKafka.Message{
		{
			Topic: topic,
			Key:   []byte(key),
			Value: e.ToByte(),
		},
	}, true)
}

func (c *OpenClient) Stop() {
	c.stop = true
	c.wg.Wait()
}

func (c *OpenClient) GetConnection(ctx context.Context) (*goKafka.Conn, error) {
	dialer := c.NewDialer(ctx)
	addr := goKafka.TCP(c.cfg.Brokers...)

	conn, err := dialer.DialContext(ctx, addr.Network(), addr.String())
	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.KafkaToKind(err),
			"[kafka] getConnection dialer.DialContext error: %s", err.Error()).
			LogError()
	}

	defer func() {
		errConn := conn.Close()
		if errConn != nil {
			_ = cerror.NewF(ctx,
				cerror.KafkaToKind(err),
				"[kafka] getConnection conn.Close error: %s", errConn.Error()).
				LogError()
		}
	}()

	controller, err := conn.Controller()
	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.KindKafkaOther,
			"[kafka] getConnection conn.Controller error: %s", err.Error()).
			LogError()
	}

	controllerConn, err := dialer.Dial(addr.Network(), net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.KafkaToKind(err),
			"[kafka] getConnection dialer.Dial error: %s", err.Error()).
			LogError()
	}

	return controllerConn, nil
}

func (c *OpenClient) NewDialer(ctx context.Context) *goKafka.Dialer {
	return &goKafka.Dialer{
		Timeout:       defaultDialerTimeout,
		DualStack:     true,
		TLS:           c.getTLS(ctx),
		SASLMechanism: c.getSASL(ctx),
	}
}

func (c *OpenClient) NewReader(ctx context.Context, topic string) *goKafka.Reader {
	var l goKafka.Logger
	if c.cfg.LoggerEnabled {
		l = &kafkaLogger{ctx: ctx, level: logger.LevelTrace}
	}

	return goKafka.NewReader(goKafka.ReaderConfig{
		Logger:                 l,
		ErrorLogger:            &kafkaLogger{ctx: ctx, level: logger.LevelError},
		Brokers:                c.cfg.Brokers,
		Topic:                  topic,
		GroupID:                c.cfg.Consumer.GroupID,
		MinBytes:               c.cfg.Consumer.MinBytes,
		MaxBytes:               c.cfg.Consumer.MaxBytes,
		MaxWait:                c.cfg.Consumer.MaxWait,
		ReadLagInterval:        c.cfg.Consumer.ReadLagInterval,
		QueueCapacity:          c.cfg.Consumer.QueueCapacity,
		ReadBatchTimeout:       c.cfg.Consumer.ReadBatchTimeout,
		HeartbeatInterval:      c.cfg.Consumer.HeartbeatInterval,
		PartitionWatchInterval: c.cfg.Consumer.PartitionWatchInterval,
		SessionTimeout:         c.cfg.Consumer.SessionTimeout,
		RebalanceTimeout:       c.cfg.Consumer.RebalanceTimeout,
		JoinGroupBackoff:       c.cfg.Consumer.JoinGroupBackoff,
		ReadBackoffMin:         c.cfg.Consumer.ReadBackoffMin,
		ReadBackoffMax:         c.cfg.Consumer.ReadBackoffMax,
		MaxAttempts:            c.cfg.Consumer.MaxAttempts,
		Dialer:                 c.NewDialer(ctx),
		RetentionTime:          oneDayTime,
		IsolationLevel:         goKafka.ReadCommitted,
	})
}

func (c *OpenClient) GetTransport(ctx context.Context) *goKafka.Transport {
	return &goKafka.Transport{
		Dial: (&net.Dialer{
			Timeout:   defaultTransportDialCtxTimeout,
			DualStack: true,
		}).DialContext,
		DialTimeout: defaultTransportDialTimeout,
		IdleTimeout: defaultTransportIdleTimeout,
		MetadataTTL: defaultTransportMetadataTTL,
		TLS:         c.getTLS(ctx),
		SASL:        c.getSASL(ctx),
		Context:     ctx,
	}
}

func (c *OpenClient) NewWriter(ctx context.Context, tr *goKafka.Transport) *goKafka.Writer {
	b := goKafka.Balancer(&goKafka.Murmur2Balancer{})
	if c.cfg.Producer.Balancer != nil {
		b = c.cfg.Producer.Balancer
	}

	var l goKafka.Logger
	if c.cfg.LoggerEnabled {
		l = &kafkaLogger{ctx: ctx, level: logger.LevelDebug}
	}

	return &goKafka.Writer{
		Logger:                 l,
		ErrorLogger:            &kafkaLogger{ctx: ctx, level: logger.LevelError},
		Balancer:               b,
		Addr:                   goKafka.TCP(c.cfg.Brokers...),
		MaxAttempts:            c.cfg.Producer.MaxAttempts,
		BatchSize:              c.cfg.Producer.WriterBatchSize,
		Transport:              tr,
		AllowAutoTopicCreation: c.cfg.AllowAutoTopicCreation,
		Async:                  c.cfg.Producer.Async,
		WriteBackoffMin:        c.cfg.Producer.WriteBackoffMin,
		WriteBackoffMax:        c.cfg.Producer.WriteBackoffMax,
		BatchTimeout:           c.cfg.Producer.BatchTimeout,
		WriteTimeout:           c.cfg.Producer.WriteTimeout,
		ReadTimeout:            c.cfg.Producer.ReadTimeout,
		RequiredAcks:           c.cfg.Producer.RequiredAcks.ToKafkaAcks(),
	}
}

func (c *OpenClient) incErrCnt() int {
	c.errCntMu.Lock()
	defer c.errCntMu.Unlock()
	c.failedMessagesCount++

	return c.failedMessagesCount
}

func (c *OpenClient) SendRetry(ctx context.Context, topic string, msgs []goKafka.Message, retry bool) error {
	tr := c.GetTransport(context.Background())
	defer tr.CloseIdleConnections()

	wr := c.NewWriter(ctx, tr)
	defer func() {
		errWR := wr.Close()
		if errWR != nil {
			_ = cerror.NewF(ctx,
				cerror.KafkaToKind(errWR),
				"sendMessage topic: %s. %s", topic, errWR.Error()).LogError()
		}
	}()

	err := wr.WriteMessages(context.Background(), msgs...)
	if err != nil {
		if retry && c.IsCheckRetry(err) {
			_ = cerror.NewF(ctx,
				cerror.KafkaToKind(err),
				"sendMessage topic: %s. Trying to retry. %s", topic, err.Error()).
				LogError()

			return c.ProcessRetry(ctx, wr, topic, msgs)
		}

		return cerror.NewF(
			ctx, cerror.KafkaToKind(err), "sendMessage topic: %s. %s", topic, err.Error()).
			LogError()
	}

	return nil
}

func (c *OpenClient) ProcessRetry(ctx context.Context,
	wr *goKafka.Writer, topic string, msgs []goKafka.Message) error {
	var errWR error
	for i := 0; i < c.cfg.Producer.MaxRetry; i++ {
		errWR = wr.WriteMessages(ctx, msgs...)
		if errors.Is(errWR, goKafka.LeaderNotAvailable) ||
			errors.Is(errWR, context.DeadlineExceeded) ||
			c.IsCheckRetry(errWR) {
			time.Sleep(c.cfg.Producer.MaxAttemptsDelay)
			continue
		}

		if errWR == nil {
			return nil
		}

		_ = cerror.NewF(
			ctx, cerror.KafkaToKind(errWR), "sendMessage topic: %s. %d attempt failed. %s", topic, i, errWR.Error()).
			LogError()
	}

	if errWR != nil {
		return cerror.New(ctx, cerror.KafkaToKind(errWR), errWR).LogError()
	}

	return nil
}

func (c *OpenClient) IsCheckRetry(errObj interface{}) bool {
	switch errWR := errObj.(type) {
	case goKafka.WriteErrors:
		return errWR.Count() > 0
	case error:
		return strings.Contains(errWR.Error(), "error writing messages to") &&
			strings.Contains(errWR.Error(), "i/o timeout")
	}

	return false
}

//nolint:nestif
func (c *OpenClient) getTLS(ctx context.Context) *tls.Config {
	var cfgTLS *tls.Config

	if c.cfg.TLS.Enabled {
		cfgTLS.MinVersion = tls.VersionTLS12

		if c.cfg.TLS.InsecureSkipVerify {
			cfgTLS.InsecureSkipVerify = true
		}

		if c.cfg.TLS.ClientCertFile != "" && c.cfg.TLS.ClientKeyFile != "" {
			// Load client cert
			cert, err := tls.LoadX509KeyPair(c.cfg.TLS.ClientCertFile, c.cfg.TLS.ClientKeyFile)
			if err != nil {
				log.Log(logger.NewEventF(ctx, logger.LevelError, "[TLS] tls.LoadX509KeyPair clientCA: %s", err))

				return nil
			}

			cfgTLS.Certificates = []tls.Certificate{cert}
		}

		if c.cfg.TLS.RootCACertFile != "" {
			// Load CA cert
			caCert, err := os.ReadFile(c.cfg.TLS.RootCACertFile)
			if err != nil {
				log.Log(logger.NewEventF(ctx, logger.LevelError, "[TLS] tls.RootCACertFile rootCA: %s", err))

				return nil
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			cfgTLS.RootCAs = caCertPool
		}
	}

	return cfgTLS
}

func (c *OpenClient) getSASL(ctx context.Context) sasl.Mechanism {
	var (
		mechanism sasl.Mechanism
		err       error
	)

	switch c.cfg.SASL.SecurityProtocol {
	case securityProtocolPlain:
		mechanism = plain.Mechanism{
			Username: c.cfg.SASL.Username,
			Password: c.cfg.SASL.Password,
		}
	case securityProtocolSCRAM:
		mechanism, err = scram.Mechanism(
			c.getSASLAlgorithm(),
			c.cfg.SASL.Username,
			c.cfg.SASL.Password)

		if err != nil {
			log.Log(logger.NewEventF(ctx, logger.LevelError, "[SASL] scram.Mechanism: %s", err))
		}
	}

	return mechanism
}

func (c *OpenClient) getSASLAlgorithm() scram.Algorithm {
	var alg scram.Algorithm

	switch c.cfg.SASL.Algorithm {
	case saslSCRAMAlgorithmSHA256:
		alg = scram.SHA256
	case saslSCRAMAlgorithmSHA512:
		alg = scram.SHA512
	}

	return alg
}
