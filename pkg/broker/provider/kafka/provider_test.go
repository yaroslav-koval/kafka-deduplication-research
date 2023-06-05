package kafka_test

import (
	"context"
	"kafka-polygon/pkg/broker/event"
	pKafka "kafka-polygon/pkg/broker/provider/kafka"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	bgCtx = context.Background()
	e     = event.WorkflowData{
		ID: "test-id",
		Workflow: event.Workflow{
			ID:     "test-wf-id",
			Schema: "test-type",
			Step:   "test-task",
		},
		Header: event.Header{
			RequestID: "test-x-request-id",
		},
	}
)

type MockedKafka struct {
	mock.Mock
}

func (mk *MockedKafka) ListenTopic(_ context.Context, topic string, handler pKafka.MessageHandler) chan error {
	_ = mk.Called(topic, handler)
	return nil
}

func (mk *MockedKafka) GetIsTopicExists(_ context.Context, topic string) (bool, error) {
	_ = mk.Called(topic)
	return true, nil
}

func (mk *MockedKafka) SendMessage(ctx context.Context, topic string, e event.BaseEvent) error {
	args := mk.Called(ctx, topic, e)
	return args.Error(0)
}

func (mk *MockedKafka) Stop() {
	_ = mk.Called()
}

func TestKafkaProvider(t *testing.T) {
	t.Parallel()

	mk := &MockedKafka{}
	expected := []interface{}{
		bgCtx,
		"test-topic",
		&e,
	}

	mk.On("SendMessage", expected...).Return(nil)

	kp := pKafka.NewKafkaProvider(nil)
	kp.SetClient(mk)

	err := kp.Publish(bgCtx, "test-topic", &e)
	require.NoError(t, err)
}

func TestKafkaProviderError(t *testing.T) {
	t.Parallel()

	mk := &MockedKafka{}
	expected := []interface{}{
		bgCtx,
		"test-topic-err",
		&e,
	}

	mk.On("SendMessage", expected...).Return(kafka.BrokerNotAvailable)

	kp := pKafka.NewKafkaProvider(nil)
	kp.SetClient(mk)

	err := kp.Publish(bgCtx, "test-topic-err", &e)
	require.Error(t, err)
	assert.Equal(t, kafka.BrokerNotAvailable, err)
}

func TestKafkaGetIsTopicExists(t *testing.T) {
	t.Parallel()

	mk := &MockedKafka{}
	expectedTopic := "test-topic"

	mk.On("GetIsTopicExists", expectedTopic).Return(true, nil)

	kp := pKafka.NewKafkaProvider(nil)
	kp.SetClient(mk)

	check, err := kp.GetIsTopicExists(bgCtx, expectedTopic)
	require.NoError(t, err)
	assert.Equal(t, true, check)
}
