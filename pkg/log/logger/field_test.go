package logger_test

import (
	"kafka-polygon/pkg/log/logger"
	"testing"

	"github.com/tj/assert"
)

func TestFieldConst(t *testing.T) {
	for k, v := range map[string]string{
		logger.FieldMessage:         "message",
		logger.FieldRequestID:       "requestID",
		logger.FieldPayload:         "payload",
		logger.FieldStackTrace:      "stackTrace",
		logger.FieldError:           "error",
		logger.FieldErrorKind:       "errorKind",
		logger.FieldErrorCode:       "errorCode",
		logger.FieldErrorAttributes: "errorAttributes",
		logger.FieldStatus:          "status",
		logger.FieldMethod:          "method",
		logger.FieldLatency:         "latency",
		logger.FieldURL:             "url",
		logger.FieldLogType:         "logType",
		logger.FieldServiceName:     "serviceName",
		logger.FieldClientIP:        "clientIP",
		logger.FieldClientID:        "clientID",
		logger.FieldUserID:          "userID",
		logger.FieldUserAgent:       "userAgent",
		logger.FieldResponse:        "response",
	} {
		assert.Equal(t, k, v)
	}
}

func TestField(t *testing.T) {
	for k, v := range map[string]interface{}{
		"a": "string",
		"b": 1,
		"c": 2.5,
		"d": true,
		"e": nil,
	} {
		f := logger.NewField(k, v)
		assert.Equal(t, k, f.Key())
		assert.Equal(t, v, f.Value())
	}
}

func TestFields(t *testing.T) {
	fields := logger.NewFields()
	assert.NotNil(t, fields)
	assert.Empty(t, fields)

	expected := []*logger.Field{logger.NewField("a", "string"), logger.NewField("b", 2)}

	fields = logger.NewFields(expected...)
	assert.EqualValues(t, expected, fields)
}

func TestFieldsAddValue(t *testing.T) {
	f1 := logger.NewField("a", "a")

	var fields logger.Fields

	fields = fields.AddValue(f1.Key(), f1.Value())

	assert.Equal(t, logger.NewFields(f1), fields)

	f2 := logger.NewField("b", "b")
	newFields := fields.AddValue(f2.Key(), f2.Value())

	assert.Equal(t, 1, len(fields))
	assert.Equal(t, logger.NewFields(f1, f2), newFields)

	f3 := logger.NewField("b", "c")
	newFields = fields.AddValue(f3.Key(), f3.Value())

	assert.Equal(t, logger.NewFields(f1, f3), newFields)
	assert.Equal(t, "b", f2.Value())

	f4 := logger.NewField("a", "d")
	newFields = newFields.AddValue(f4.Key(), f4.Value())

	assert.Equal(t, logger.NewFields(f3, f4), newFields)
	assert.Equal(t, "a", f1.Value())
}

func TestFieldsGetValue(t *testing.T) {
	expected := logger.NewField("a", "value")
	fields := logger.NewFields()

	actualValue, actualExists := fields.GetValue(expected.Key())

	assert.False(t, actualExists)
	assert.Nil(t, actualValue)

	fields = logger.NewFields(expected)
	actualValue, actualExists = fields.GetValue(expected.Key())

	assert.True(t, actualExists)
	assert.Equal(t, expected.Value(), actualValue)
}

func TestFieldsMap(t *testing.T) {
	expected := logger.NewField("a", "value")

	fields := logger.NewFields(expected)

	assert.Equal(t, map[string]interface{}{expected.Key(): expected.Value()}, fields.Map())
}
