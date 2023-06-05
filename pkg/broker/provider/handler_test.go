package provider_test

import (
	"bytes"
	"context"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	bgCtx     = context.WithValue(context.Background(), consts.HeaderXRequestID, "test-x-request-id") //nolint:staticcheck
	expHeader = event.Header{RequestID: "test-x-request-id"}
	errEmpty  = cerror.NewF(bgCtx, cerror.KindNotExist, "empty error")
)

type buffWriter struct {
	wr bytes.Buffer
	m  sync.Mutex
}

func (sw *buffWriter) Write(data []byte) (n int, err error) {
	sw.m.Lock()
	n, err = sw.wr.Write(data)
	sw.m.Unlock()

	return
}

func (sw *buffWriter) String() string {
	sw.m.Lock()
	defer sw.m.Unlock()

	return sw.wr.String()
}

func TestHandlerWorkflow(t *testing.T) {
	t.Parallel()

	var buff buffWriter

	fn := func(ctx context.Context, workflowEvent event.WorkflowEvent, ed store.EventProcessData) error {
		_, err := buff.Write(workflowEvent.ToByte())
		require.NoError(t, err)

		return nil
	}

	var (
		i   interface{} = fn
		p   interface{} = provider.HandlerWorkflow(fn)
		pFn             = provider.HandlerWorkflow(fn)
	)

	_, ok := i.(provider.HandlerFn)
	assert.Equal(t, ok, false)

	_, ok = p.(provider.HandlerFn)
	assert.Equal(t, ok, true)
	assert.Equal(t, &event.WorkflowData{}, pFn.GetEventData(bgCtx))

	eW := &event.WorkflowData{
		ID: "tests-id",
		Workflow: event.Workflow{
			ID:     "test-wf-id",
			Schema: "test-type",
			Step:   "test-task",
		},
		Header: expHeader,
	}

	err := pFn.CallFn(bgCtx, eW, store.EventProcessData{})
	require.NoError(t, err)
	assert.Equal(t, string(eW.ToByte()), buff.String())
}

func TestHandlerWorkflowError(t *testing.T) {
	t.Parallel()

	fn := provider.HandlerWorkflow(func(ctx context.Context, workflowEvent event.WorkflowEvent, ed store.EventProcessData) error {
		return errEmpty
	})

	ed := &event.WorkflowData{}

	err := fn.CallFn(bgCtx, ed, store.EventProcessData{})
	require.Error(t, err)
	assert.Equal(t, errEmpty, err)
}

func TestHandlerDebezium(t *testing.T) {
	t.Parallel()

	var buff buffWriter

	fn := func(ctx context.Context, debeziumEvent event.DebeziumEvent, ed store.EventProcessData) error {
		_, err := buff.Write(debeziumEvent.ToByte())
		require.NoError(t, err)

		return nil
	}

	var (
		i   interface{} = fn
		p   interface{} = provider.HandlerDebezium(fn)
		pFn             = provider.HandlerDebezium(fn)
	)

	_, ok := i.(provider.HandlerFn)
	assert.Equal(t, ok, false)

	_, ok = p.(provider.HandlerFn)
	assert.Equal(t, ok, true)

	edd := &event.DebeziumData{
		Key: &event.KeyData{},
	}
	edd.WithContext(bgCtx)

	assert.Equal(t, edd, pFn.GetEventData(bgCtx))

	eD := &event.DebeziumData{
		Key: &event.KeyData{
			Payload: event.KeyPayload{ID: "test-payload"},
		},
		Payload: &event.DebeziumPayload{
			Op: "test-op",
		},
		Header: expHeader,
	}

	err := pFn.CallFn(bgCtx, eD, store.EventProcessData{})
	require.NoError(t, err)
	assert.Equal(t, string(eD.ToByte()), buff.String())
}

func TestHandlerDebeziumError(t *testing.T) {
	t.Parallel()

	fn := provider.HandlerDebezium(func(ctx context.Context, debeziumEvent event.DebeziumEvent, ed store.EventProcessData) error {
		return errEmpty
	})

	edd := &event.DebeziumData{}

	err := fn.CallFn(bgCtx, edd, store.EventProcessData{})
	require.Error(t, err)
	assert.Equal(t, errEmpty, err)
}

func TestHandlerMinio(t *testing.T) {
	t.Parallel()

	var buff buffWriter

	fn := func(ctx context.Context, minioEvent event.MinioEvent, ed store.EventProcessData) error {
		_, err := buff.Write(minioEvent.ToByte())
		require.NoError(t, err)

		return nil
	}

	var (
		i   interface{} = fn
		p   interface{} = provider.HandlerMinio(fn)
		pFn             = provider.HandlerMinio(fn)
	)

	_, ok := i.(provider.HandlerFn)
	assert.Equal(t, ok, false)

	_, ok = p.(provider.HandlerFn)
	assert.Equal(t, ok, true)
	assert.Equal(t, &event.MinioData{}, pFn.GetEventData(bgCtx))

	eM := &event.MinioData{
		EventName: "tests-event-name",
		Key:       "test-key",
		Header:    expHeader,
	}

	err := pFn.CallFn(bgCtx, eM, store.EventProcessData{})
	require.NoError(t, err)
	assert.Equal(t, string(eM.ToByte()), buff.String())
}

func TestHandlerMinioError(t *testing.T) {
	t.Parallel()

	fn := provider.HandlerMinio(func(ctx context.Context, minioEvent event.MinioEvent, ed store.EventProcessData) error {
		return errEmpty
	})

	eMD := &event.MinioData{}

	err := fn.CallFn(bgCtx, eMD, store.EventProcessData{})
	require.Error(t, err)
	assert.Equal(t, errEmpty, err)
}

func TestHandlerNotification(t *testing.T) {
	t.Parallel()

	var buff buffWriter

	fn := func(ctx context.Context, notifyEvent event.NotificationEvent, ed store.EventProcessData) error {
		_, err := buff.Write(notifyEvent.ToByte())
		require.NoError(t, err)

		return nil
	}

	var (
		i   interface{} = fn
		p   interface{} = provider.HandlerNotification(fn)
		pFn             = provider.HandlerNotification(fn)
	)

	_, ok := i.(provider.HandlerFn)
	assert.Equal(t, ok, false)

	_, ok = p.(provider.HandlerFn)
	assert.Equal(t, ok, true)
	assert.Equal(t, &event.NotificationData{}, pFn.GetEventData(bgCtx))

	eN := &event.NotificationData{
		ID:         "tests-id",
		WorkflowID: "test-wf-id",
		Channel:    "test-channel",
		Provider:   "test-provider",
		Header:     expHeader,
		Debug:      true,
	}

	err := pFn.CallFn(bgCtx, eN, store.EventProcessData{})
	require.NoError(t, err)
	assert.Equal(t, string(eN.ToByte()), buff.String())
}

func TestHandlerNotificationError(t *testing.T) {
	t.Parallel()

	fn := provider.HandlerNotification(func(ctx context.Context, notifyEvent event.NotificationEvent, ed store.EventProcessData) error {
		return errEmpty
	})

	eND := &event.NotificationData{}

	err := fn.CallFn(bgCtx, eND, store.EventProcessData{})
	require.Error(t, err)
	assert.Equal(t, errEmpty, err)
}
