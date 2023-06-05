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

const (
	securityProtocolSCRAM          = "SCRAM"
	securityProtocolPlain          = "PLAIN"
	saslSCRAMAlgorithmSHA256       = "SCRAM-SHA-256"
	saslSCRAMAlgorithmSHA512       = "SCRAM-SHA-512"
	defaultRerunDelay              = 30 * time.Second
	defaultDialerTimeout           = 10 * time.Second
	defaultMinBackoff              = 100 * time.Millisecond
	defaultMaxReadBackoff          = 1 * time.Second
	oneDayTime                     = 24 * time.Hour
	defaultTransportDialCtxTimeout = 3 * time.Second
	defaultTransportMetadataTTL    = 6 * time.Second
	defaultTransportDialTimeout    = 5 * time.Second
	defaultTransportIdleTimeout    = 30 * time.Second

	consumerMinBytesDefValue       = 10e3 // 10KB
	consumerMaxBytesDefValue       = 10e6 // 10MB
	consumerQueueCapacity          = 100
	consumerMaxWaitDefValue        = 10 * time.Second
	consumerReadBatchTimeout       = 10 * time.Second
	consumerHeartbeatInterval      = 3 * time.Second
	consumerPartitionWatchInterval = 5 * time.Second
	consumerSessionTimeout         = 30 * time.Second
	consumerRebalanceTimeout       = 30 * time.Second
	consumerJoinGroupBackoff       = 5 * time.Second

	producerBatchTimeout = 1 * time.Second
	producerWriteTimeout = 10 * time.Second
	producerReadTimeout  = 10 * time.Second
)

type MessageHandler interface {
	Handle(ctx context.Context, m *goKafka.Message) (event.BaseEvent, error)
}

type HandelFn func(ctx context.Context, m *goKafka.Message) (event.BaseEvent, error)

func (hm HandelFn) Handle(ctx context.Context, m *goKafka.Message) (event.BaseEvent, error) {
	return hm(ctx, m)
}

type kafkaLogger struct {
	ctx   context.Context
	level logger.Level
}

func (l *kafkaLogger) Printf(msg string, args ...interface{}) {
	log.Log(logger.NewEventF(l.ctx, l.level, msg, args...))
}

type KClient interface {
	ListenTopic(ctx context.Context, topic string, handler MessageHandler) chan error
	GetIsTopicExists(ctx context.Context, topic string) (bool, error)
	SendMessage(ctx context.Context, topic string, e event.BaseEvent) error
	Stop()
}

type TLS struct {
	Enabled            bool
	InsecureSkipVerify bool
	ClientCertFile     string
	ClientKeyFile      string
	RootCACertFile     string
}

type SASL struct {
	SecurityProtocol string
	Algorithm        string
	Username         string
	Password         string
}

type Consumer struct {
	GroupID                string
	MinBytes               int
	MaxBytes               int
	MaxWait                time.Duration
	ReadLagInterval        time.Duration
	CommitOnError          bool
	ListenTopics           []string
	QueueCapacity          int
	ReadBatchTimeout       time.Duration
	HeartbeatInterval      time.Duration
	PartitionWatchInterval time.Duration
	SessionTimeout         time.Duration
	RebalanceTimeout       time.Duration
	JoinGroupBackoff       time.Duration
	ReadBackoffMin         time.Duration
	ReadBackoffMax         time.Duration
	MaxAttempts            int
}

func (c *Consumer) initDefault() {
	if c.MinBytes == 0 {
		c.MinBytes = consumerMinBytesDefValue
	}

	if c.MaxBytes == 0 {
		c.MaxBytes = consumerMaxBytesDefValue
	}

	if c.MaxWait.Seconds() == 0 {
		c.MaxWait = consumerMaxWaitDefValue
	}

	if c.QueueCapacity == 0 {
		c.QueueCapacity = consumerQueueCapacity
	}

	if c.ReadBatchTimeout.Seconds() == 0 {
		c.ReadBatchTimeout = consumerReadBatchTimeout
	}

	if c.HeartbeatInterval.Seconds() == 0 {
		c.HeartbeatInterval = consumerHeartbeatInterval
	}

	if c.PartitionWatchInterval.Seconds() == 0 {
		c.PartitionWatchInterval = consumerPartitionWatchInterval
	}

	if c.SessionTimeout.Seconds() == 0 {
		c.SessionTimeout = consumerSessionTimeout
	}

	if c.RebalanceTimeout.Seconds() == 0 {
		c.RebalanceTimeout = consumerRebalanceTimeout
	}

	if c.JoinGroupBackoff.Seconds() == 0 {
		c.JoinGroupBackoff = consumerJoinGroupBackoff
	}

	if c.ReadBackoffMin.Milliseconds() == 0 {
		c.ReadBackoffMin = defaultMinBackoff
	}

	if c.ReadBackoffMax.Seconds() == 0 {
		c.ReadBackoffMax = defaultMaxReadBackoff
	}
}

type RequiredAcks string

func (ra RequiredAcks) ToKafkaAcks() goKafka.RequiredAcks {
	switch ra {
	case "none":
		return goKafka.RequireNone
	case "one":
		return goKafka.RequireOne
	case "all":
		return goKafka.RequireAll
	default:
		return goKafka.RequireNone
	}
}

type Producer struct {
	Balancer           goKafka.Balancer
	MaxAttempts        int
	MaxRetry           int
	MaxAttemptsDelay   time.Duration
	WriteTimeout       time.Duration
	WriterBatchSize    int
	WriterBatchTimeout time.Duration
	WriteBackoffMin    time.Duration
	WriteBackoffMax    time.Duration
	BatchTimeout       time.Duration
	ReadTimeout        time.Duration
	RequiredAcks       RequiredAcks
	Async              bool
}

func (p *Producer) initDefault() {
	if p.BatchTimeout.Seconds() == 0 {
		p.BatchTimeout = producerBatchTimeout
	}

	if p.WriteTimeout.Seconds() == 0 {
		p.WriteTimeout = producerWriteTimeout
	}

	if p.ReadTimeout.Seconds() == 0 {
		p.ReadTimeout = producerReadTimeout
	}

	if p.WriteBackoffMin.Milliseconds() == 0 {
		p.WriteBackoffMin = defaultMinBackoff
	}

	if p.WriteBackoffMax.Seconds() == 0 {
		p.WriteBackoffMax = defaultMaxReadBackoff
	}
}

type Config struct {
	Brokers                    []string
	LoggerEnabled              bool
	CommitOnErrorMessagesCount int
	RerunDelay                 time.Duration
	TLS                        TLS
	SASL                       SASL
	UseKeyDoubleQuote          bool
	AllowAutoTopicCreation     bool
	Consumer                   Consumer
	Producer                   Producer
}

func (c *Config) defaults() {
	// init consumer config default
	c.Consumer.initDefault()

	// init producer config default
	c.Producer.initDefault()

	if c.RerunDelay.Seconds() == 0 {
		c.RerunDelay = defaultRerunDelay
	}
}

type Client struct {
	errCntMu            sync.Mutex
	cfg                 *Config
	failedMessagesCount int
	wg                  sync.WaitGroup
	stop                bool
}

func NewClient(cfg *Config) KClient {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg.defaults()

	return &Client{
		cfg: cfg,
	}
}

func (c *Client) ListenTopic(ctx context.Context, topic string, handler MessageHandler) chan error {
	errCh := make(chan error)
	reader := c.newReader(ctx, topic)

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

func (c *Client) GetIsTopicExists(ctx context.Context, topic string) (bool, error) {
	conn, err := c.getConnection(ctx)
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

func (c *Client) SendMessage(ctx context.Context, topic string, e event.BaseEvent) error {
	key := e.GetID()
	if c.cfg.UseKeyDoubleQuote {
		key = fmt.Sprintf("%q", e.GetID())
	}

	return c.sendRetry(ctx, topic, []goKafka.Message{
		{
			Topic: topic,
			Key:   []byte(key),
			Value: e.ToByte(),
		},
	}, true)
}

func (c *Client) Stop() {
	c.stop = true
	c.wg.Wait()
}

func (c *Client) getConnection(ctx context.Context) (*goKafka.Conn, error) {
	dialer := c.newDialer(ctx)
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

func (c *Client) newDialer(ctx context.Context) *goKafka.Dialer {
	return &goKafka.Dialer{
		Timeout:       defaultDialerTimeout,
		DualStack:     true,
		TLS:           c.getTLS(ctx),
		SASLMechanism: c.getSASL(ctx),
	}
}

func (c *Client) newReader(ctx context.Context, topic string) *goKafka.Reader {
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
		Dialer:                 c.newDialer(ctx),
		RetentionTime:          oneDayTime,
		IsolationLevel:         goKafka.ReadCommitted,
	})
}

func (c *Client) getTransport(ctx context.Context) *goKafka.Transport {
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

func (c *Client) newWriter(ctx context.Context, tr *goKafka.Transport) *goKafka.Writer {
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

func (c *Client) incErrCnt() int {
	c.errCntMu.Lock()
	defer c.errCntMu.Unlock()
	c.failedMessagesCount++

	return c.failedMessagesCount
}

func (c *Client) sendRetry(ctx context.Context, topic string, msgs []goKafka.Message, retry bool) error {
	tr := c.getTransport(context.Background())
	defer tr.CloseIdleConnections()

	wr := c.newWriter(ctx, tr)
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
		if retry && c.isCheckRetry(err) {
			_ = cerror.NewF(ctx,
				cerror.KafkaToKind(err),
				"sendMessage topic: %s. Trying to retry. %s", topic, err.Error()).
				LogError()

			return c.processRetry(ctx, wr, topic, msgs)
		}

		return cerror.NewF(
			ctx, cerror.KafkaToKind(err), "sendMessage topic: %s. %s", topic, err.Error()).
			LogError()
	}

	return nil
}

func (c *Client) processRetry(ctx context.Context,
	wr *goKafka.Writer, topic string, msgs []goKafka.Message) error {
	var errWR error
	for i := 0; i < c.cfg.Producer.MaxRetry; i++ {
		errWR = wr.WriteMessages(ctx, msgs...)
		if errors.Is(errWR, goKafka.LeaderNotAvailable) ||
			errors.Is(errWR, context.DeadlineExceeded) ||
			c.isCheckRetry(errWR) {
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

func (c *Client) isCheckRetry(errObj interface{}) bool {
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
func (c *Client) getTLS(ctx context.Context) *tls.Config {
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

func (c *Client) getSASL(ctx context.Context) sasl.Mechanism {
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

func (c *Client) getSASLAlgorithm() scram.Algorithm {
	var alg scram.Algorithm

	switch c.cfg.SASL.Algorithm {
	case saslSCRAMAlgorithmSHA256:
		alg = scram.SHA256
	case saslSCRAMAlgorithmSHA512:
		alg = scram.SHA512
	}

	return alg
}
