package broker_test

import (
	"context"
	"errors"
	"kafka-polygon/pkg/broker"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider"
	"kafka-polygon/pkg/broker/provider/kafka"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cmd/metadata"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/tracing"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	bgCtx           = context.WithValue(context.Background(), consts.HeaderXRequestID, "test-x-request-id") //nolint:staticcheck
	errProviderSend = errors.New("error send message")
)

type MockedProvider struct {
	mock.Mock
}

func (mp *MockedProvider) SetEnabled(enable bool) {
	_ = mp.Called(enable)
}

func (mp *MockedProvider) SetStore(s store.Store) {
	_ = mp.Called(s)
}

func (mp *MockedProvider) SetTracing(t tracing.Tracer) {
	_ = mp.Called(t)
}

func (mp *MockedProvider) AddHandler(_ context.Context, topic string, fn provider.HandlerFn) {
	_ = mp.Called(topic, fn)
}

func (mp *MockedProvider) GetIsTopicExists(_ context.Context, topic string) (bool, error) {
	args := mp.Called(topic)
	return args.Bool(0), args.Error(1)
}

func (mp *MockedProvider) Publish(ctx context.Context, topic string, e event.BaseEvent) error {
	args := mp.Called(ctx, topic, e)
	return args.Error(0)
}

func (mp *MockedProvider) Sync(_ context.Context, topic string, handler provider.HandlerFn) {
	_ = mp.Called(topic, handler)
}

func (mp *MockedProvider) GetType() string {
	args := mp.Called()
	return args.String(0)
}

func (mp *MockedProvider) Stop() {
	mp.Called()
}

func TestBrokerKafkaSend(t *testing.T) {
	t.Parallel()

	expMetadata := metadata.Meta{
		Version: "0.0.1",
	}

	e := &event.WorkflowData{
		ID: "test-id-1",
		Workflow: event.Workflow{
			ID:     "test-wf-id-1",
			Schema: "test-type-1",
			Step:   "test-task-1",
		},
	}

	expected := []interface{}{
		bgCtx,
		"test-topic",
		e,
	}

	kafkaProvider := new(MockedProvider)
	kafkaProvider.On("GetType").Return(kafka.BrokerKafkaProvider)
	kafkaProvider.On("Publish", expected...).Return(nil)

	bq := broker.New(kafkaProvider)
	bq.SetMeta(expMetadata)
	err := bq.Send(bgCtx, "test-topic", e)
	require.NoError(t, err)
	assert.Equal(t, kafka.BrokerKafkaProvider, bq.Name())
	assert.Equal(t, bgCtx.Value(consts.HeaderXRequestID), e.Header.RequestID)
	assert.Equal(t, expMetadata, e.GetMeta())
}

func TestBrokerKafkaSendError(t *testing.T) {
	t.Parallel()

	e := &event.WorkflowData{
		ID: "test-id-2",
		Workflow: event.Workflow{
			ID:     "test-wf-id-2",
			Schema: "test-type-2",
			Step:   "test-task-2",
		},
	}

	expected := []interface{}{
		bgCtx,
		"test-topic",
		e,
	}

	kafkaProvider := new(MockedProvider)
	kafkaProvider.On("GetType").Return(kafka.BrokerKafkaProvider)
	kafkaProvider.On("Publish", expected...).Return(errProviderSend)

	bq := broker.New(kafkaProvider)
	err := bq.Send(bgCtx, "test-topic", e)
	require.Error(t, err)
	assert.Equal(t, errProviderSend, err)
	assert.Equal(t, bgCtx.Value(consts.HeaderXRequestID), e.Header.RequestID)
}

func TestBrokerKafkaIsTopicExists(t *testing.T) {
	t.Parallel()

	kafkaProvider := new(MockedProvider)
	kafkaProvider.On("GetType").Return(kafka.BrokerKafkaProvider)
	kafkaProvider.On("GetIsTopicExists", "test-topic").Return(true, nil)

	bq := broker.New(kafkaProvider)
	exists, err := bq.GetIsTopicExists(bgCtx, "test-topic")
	require.NoError(t, err)
	assert.Equal(t, true, exists)
	assert.Equal(t, kafka.BrokerKafkaProvider, bq.Name())
}

func TestBrokerName(t *testing.T) {
	t.Parallel()

	kafkaProvider := new(MockedProvider)
	kafkaProvider.On("GetType").Return(kafka.BrokerKafkaProvider)

	bq := broker.New(kafkaProvider)
	assert.Equal(t, kafka.BrokerKafkaProvider, bq.Name())
}
