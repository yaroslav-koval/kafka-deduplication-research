package kafka_test

import (
	"context"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider/kafka"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"sync"
	"testing"
	"time"

	goKafka "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
)

type kafkaTestSuite struct {
	suite.Suite
	ctx      context.Context
	brokers  []string
	defTopic string
	z        *testutil.DockerZKContainer
	k        *testutil.DockerKafkaContainer
}

func TestKafkaTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(kafkaTestSuite))
}

func (s *kafkaTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.brokers = []string{"localhost:9093"}
	s.defTopic = "demo_topic"

	inst := testutil.NewDockerUtilInstance()

	s.z = inst.
		InitZK().
		Connect()

	kafkaConf := testutil.KafkaParams{
		ZKHost: s.z.Name,
		ZKPort: "2181",
	}

	s.k = inst.
		InitKafka(kafkaConf).
		Connect()
}

func (s *kafkaTestSuite) TearDownSuite() {
	s.k.Close()
	s.z.Close()
}

func (s *kafkaTestSuite) TestKafka() {
	kc := kafka.NewClient(&kafka.Config{
		Brokers: s.brokers,
		Consumer: kafka.Consumer{
			GroupID:       "test_group_1",
			CommitOnError: true,
			MaxAttempts:   3,
		},
		Producer: kafka.Producer{
			MaxRetry:           10,
			MaxAttempts:        1,
			Balancer:           &goKafka.Hash{},
			WriteTimeout:       10 * time.Second,
			WriterBatchTimeout: 10 * time.Second,
			RequiredAcks:       "all",
		},
		RerunDelay:             10 * time.Second,
		AllowAutoTopicCreation: true,
	})

	expEvent := &event.WorkflowData{
		ID: "test-id",
		Workflow: event.Workflow{
			ID:     "test-wf-id",
			Schema: "test-type",
		},
	}

	err := kc.SendMessage(bgCtx, s.defTopic, expEvent)

	s.NoError(err)

	ok, err := kc.GetIsTopicExists(bgCtx, s.defTopic)
	s.NoError(err)
	s.True(ok)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	chMsg := make(chan *event.WorkflowData, 1)

	handler := kafka.MessageHandler(kafka.HandelFn(func(ctx context.Context, m *goKafka.Message) (event.BaseEvent, error) {
		obj := &event.WorkflowData{}
		em := event.Message{
			Key:   converto.BytePointer(m.Key),
			Value: m.Value,
		}
		errE := obj.Unmarshal(em)
		if errE != nil {
			s.T().Logf("skip msg: %s", string(m.Key))
			return obj, nil
		}

		chMsg <- obj
		wg.Done()

		return obj, nil
	}))

	kc.ListenTopic(bgCtx, s.defTopic, handler)

	select {
	case <-wait(wg):
		msg, ok := <-chMsg
		if !ok || msg == nil {
			s.T().Fatal("not found kafka message")
		}

		s.Equal(expEvent.GetID(), msg.GetID())
		s.Equal(expEvent.GetWorkflow().ID, msg.GetWorkflow().ID)
		s.Equal(expEvent.GetWorkflow().Schema, msg.GetWorkflow().Schema)
	case <-s.ctx.Done():
		s.T().Logf("context.Err %s", s.ctx.Err())
		s.T().Error("context was done too quickly immediately")
	}
}

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)

	go func() {
		wg.Wait()
		ch <- true
	}()

	return ch
}

func (s *kafkaTestSuite) TestKafkaNotValidBroker() {
	kc := kafka.NewClient(&kafka.Config{
		Brokers: []string{"not-valid-broker:9093"},
	})

	expEvent := &event.WorkflowData{
		ID: "test-id-err",
		Workflow: event.Workflow{
			ID:     "test-wf-id-err",
			Schema: "test-type-err",
		},
	}

	err := kc.SendMessage(bgCtx, "not-valid-topic", expEvent)
	s.Error(err)

	s.True(s.Contains(err.Error(), "sendMessage topic: not-valid-topic. dial tcp: lookup not-valid-broker"))
	s.True(s.checkErrKind(cerror.ErrKind(err), []cerror.Kind{
		cerror.KindKafkaIO,
		cerror.KindKafkaOther,
	}))
}

func (s *kafkaTestSuite) TestKafkaNotValidTopic() {
	kc := kafka.NewClient(&kafka.Config{
		Brokers: s.brokers,
	})

	expErrMsg := "sendMessage topic: not-valid-topic. " +
		"[3] Unknown Topic Or Partition: the request is for a " +
		"topic or partition that does not exist on this broker"

	expEvent := &event.WorkflowData{
		ID: "test-id-err",
		Workflow: event.Workflow{
			ID:     "test-wf-id-err",
			Schema: "test-type-err",
		},
	}

	err := kc.SendMessage(bgCtx, "not-valid-topic", expEvent)

	s.Error(err)
	s.Equal(expErrMsg, err.Error())
	s.Equal(cerror.KindKafkaUnknown.String(), cerror.ErrKind(err).String())
}

func (s *kafkaTestSuite) checkErrKind(source cerror.Kind, list []cerror.Kind) bool {
	for _, item := range list {
		if item == source {
			return true
		}
	}

	return false
}
