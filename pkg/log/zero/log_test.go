package zero_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/log/zero"
	"testing"

	"github.com/tj/assert"
)

type testEvent struct {
	lvl    logger.Level
	ctx    context.Context
	msg    string
	err    error
	fields logger.Fields
}

func (e *testEvent) Level() logger.Level {
	return e.lvl
}

func (e *testEvent) Ctx() context.Context {
	return e.ctx
}

func (e *testEvent) Msg() string {
	return e.msg
}

func (e *testEvent) Err() error {
	return e.err
}

func (e *testEvent) Fields() logger.Fields {
	return e.fields
}

type testLogMsg struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Error     string `json:"error"`
	RequestID string `json:"requestID"`
}

func TestLoggerSetGlobalOutput(t *testing.T) {
	l := zero.New()
	buf := new(bytes.Buffer)

	l.Log(&testEvent{lvl: logger.LevelInfo})
	assert.Empty(t, buf.Bytes())

	l.SetGlobalOutput(buf)
	l.Log(&testEvent{lvl: logger.LevelInfo})
	assert.NotEmpty(t, buf.Bytes())
}

func TestLoggerLog(t *testing.T) {
	l := zero.New()
	buf := new(bytes.Buffer)
	l.SetGlobalOutput(buf)

	e := &testEvent{
		lvl:    logger.LevelInfo,
		ctx:    context.WithValue(context.Background(), consts.HeaderXRequestID, "123"), //nolint:staticcheck
		msg:    "message",
		err:    fmt.Errorf("err"),
		fields: logger.NewFields(logger.NewField("name", "John")),
	}
	l.Log(e)

	actualMsg := struct {
		testLogMsg
		Name string
	}{}
	_ = json.Unmarshal(buf.Bytes(), &actualMsg)

	assert.Equal(t, testLogMsg{
		Level:     "info",
		Message:   e.Msg(),
		Error:     e.Err().Error(),
		RequestID: fmt.Sprintf("%s", e.ctx.Value(consts.HeaderXRequestID)),
	}, actualMsg.testLogMsg)
	assert.Equal(t, "John", actualMsg.Name)
}

func TestLoggerSetGlobalLogLevel(t *testing.T) {
	l := zero.New()
	buf := new(bytes.Buffer)
	l.SetGlobalOutput(buf)

	for _, v := range []struct {
		globalLvl logger.Level
		emptyLvl  logger.Level
	}{
		{logger.LevelDebug, logger.LevelTrace},
		{logger.LevelInfo, logger.LevelDebug},
		{logger.LevelWarn, logger.LevelInfo},
		{logger.LevelError, logger.LevelWarn},
		{logger.LevelFatal, logger.LevelError},
	} {
		l.SetGlobalLogLevel(v.globalLvl.String())
		buf.Reset()
		l.Log(&testEvent{lvl: v.emptyLvl})
		assert.Empty(t, buf.Bytes())
		buf.Reset()
		l.Log(&testEvent{lvl: v.globalLvl})
		assert.NotEmpty(t, buf.Bytes())
	}

	l.SetGlobalLogLevel(logger.LevelTrace.String())
	buf.Reset()
	l.Log(&testEvent{lvl: logger.LevelTrace})
	assert.NotEmpty(t, buf.Bytes())

	buf.Reset()
	l.SetGlobalLogLevel("invalid level")
	assert.NotEmpty(t, buf.Bytes()) // parse error should appear in log
	buf.Reset()
	l.Log(&testEvent{lvl: logger.LevelDebug})
	assert.Empty(t, buf.Bytes()) // default level=info
	buf.Reset()
	l.Log(&testEvent{lvl: logger.LevelInfo})
	assert.NotEmpty(t, buf.Bytes())
}
