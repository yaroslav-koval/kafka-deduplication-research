package logger_test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log/logger"
	"testing"

	"github.com/tj/assert"
)

func TestNew(t *testing.T) {
	for _, l := range []logger.Level{
		logger.LevelTrace,
		logger.LevelDebug,
		logger.LevelInfo,
		logger.LevelWarn,
		logger.LevelError,
		logger.LevelFatal,
	} {
		msg := l.String()
		e := logger.NewEvent(context.Background(), l, msg)
		assert.Equal(t, l, e.Level())
		assert.Equal(t, msg, e.Msg())
	}
}

func TestEventCtx(t *testing.T) {
	for _, c := range []context.Context{
		context.Background(),
		context.WithValue(context.Background(), consts.HeaderXRequestID, "123"), //nolint:staticcheck
	} {
		e := logger.NewEvent(c, logger.LevelDebug, "")
		assert.Equal(t, c, e.Ctx())
	}
}

func TestEventErr(t *testing.T) {
	e := logger.NewEvent(context.Background(), logger.LevelDebug, "")
	assert.Nil(t, e.Err())

	for _, err := range []error{
		nil,
		fmt.Errorf("err"),
	} {
		assert.Equal(t, err, e.WithErr(err).Err())
	}
}

func TestWithValue(t *testing.T) {
	e := logger.NewEvent(context.Background(), logger.LevelDebug, "")
	assert.Nil(t, e.Fields())

	expected := logger.NewFields()

	for k, v := range map[string]interface{}{
		"name":   "John",
		"count":  2,
		"active": true,
		"amount": 10.5,
	} {
		expected = append(expected, logger.NewField(k, v))

		e.WithValue(k, v)
		assert.Equal(t, expected, e.Fields())
	}

	expectedLen := len(e.Fields())

	e.WithValue("name", "Other John")
	assert.Equal(t, expectedLen, len(e.Fields()))

	for _, v := range e.Fields() {
		if v.Key() == "name" {
			assert.Equal(t, "Other John", v.Value())
		}
	}
}

func TestWithPayload(t *testing.T) {
	e := logger.NewEvent(context.Background(), logger.LevelDebug, "")
	assert.Empty(t, e.Fields())

	expected := struct{ name string }{"John"}

	e.WithPayload(expected)

	var actual interface{}

	for _, v := range e.Fields() {
		if v.Key() == "payload" {
			actual = v.Value()
		}
	}

	assert.Equal(t, expected, actual)
}
