package cerror_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror/v2"
	"kafka-polygon/pkg/http/consts"
	logs "kafka-polygon/pkg/log"
	"testing"

	"github.com/tj/assert"
)

func TestNew(t *testing.T) {
	ctx := context.WithValue(context.Background(), consts.HeaderXRequestID, requestIDText) //nolint:staticcheck
	kind := cerror.KindConflict
	originalErr := fmt.Errorf(errText)
	err := cerror.New(ctx, kind, originalErr)
	assert.Equal(t, ctx.Value(consts.HeaderXRequestID), err.Ctx().Value(consts.HeaderXRequestID))
	assert.Equal(t, originalErr, err.Err())
	assert.Equal(t, originalErr.Error(), err.Error())
	assert.Equal(t, kind, err.Kind())
	assert.Contains(t, err.Ops()[0], "TestNew")
	assert.Nil(t, err.Payload())
}

func TestOpThirdNestLevel(t *testing.T) {
	err := OpSecondNestLevel()
	assert.Equal(t, 5, len(err.Ops()))
	assert.Contains(t, err.Ops()[0], "OpFirstNestLevel")
	assert.Contains(t, err.Ops()[1], "OpSecondNestLevel")
	assert.Contains(t, err.Ops()[2], "TestOpThirdNestLevel")
}

func OpSecondNestLevel() *cerror.CError {
	return OpFirstNestLevel()
}

func OpFirstNestLevel() *cerror.CError {
	ctx := context.WithValue(context.Background(), consts.HeaderXRequestID, requestIDText) //nolint:staticcheck
	kind := cerror.KindConflict
	originalErr := fmt.Errorf(errText)

	return cerror.New(ctx, kind, originalErr)
}

func TestCErrorWithPayload(t *testing.T) {
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).WithPayload(fields)
	assert.Equal(t, fields, err.Payload())
}

func TestCErrorLog(t *testing.T) {
	buf := new(bytes.Buffer)
	logs.SetGlobalOutput(buf)
	logs.SetGlobalLogLevel("debug")

	ctx := context.WithValue(context.Background(), consts.HeaderXRequestID, requestIDText) //nolint:staticcheck
	kind := cerror.KindConflict
	originalErr := fmt.Errorf(errText)
	attrs := &cerror.ErrAttributes{Code: 422}
	err := cerror.New(ctx, kind, originalErr).WithPayload(fields).WithAttributes(attrs).LogError()

	for level, logFn := range map[string]func() *cerror.CError{
		levelErrorText: err.LogError,
		levelWarnText:  err.LogWarn,
		levelFatalText: err.LogFatal,
	} {
		buf.Reset()

		_ = logFn()

		expectedMsg := testLogMsg{
			Level:      level,
			Message:    "",
			Error:      originalErr.Error(),
			RequestID:  fmt.Sprintf("%s", ctx.Value(consts.HeaderXRequestID)),
			ErrorKind:  kind.String(),
			ErrorCode:  cerror.HTTPCode(err),
			Payload:    fields,
			Operations: err.Ops(),
			Attributes: attrs,
		}

		var actualMsg testLogMsg

		_ = json.Unmarshal(buf.Bytes(), &actualMsg)

		assert.Equal(t, expectedMsg, actualMsg)
	}
}

func TestCErrorWithHTTPCode(t *testing.T) {
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText))
	for i := 0; i <= 600; i++ {
		c, ok := err.WithHTTPCode(i).HTTPCode()

		if i >= 400 && i < 600 {
			assert.True(t, ok)
			assert.Equal(t, i, c)
		} else {
			assert.False(t, ok)
		}
	}
}

func TestCErrorWithAttributes(t *testing.T) {
	attrs := &cerror.ErrAttributes{Type: "t"}
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).WithAttributes(attrs)
	assert.Equal(t, attrs, err.Attributes())
}

func TestCErrorWithAttributesHTTP(t *testing.T) {
	attrs := &cerror.ErrAttributes{
		Type:    "api",
		Subtype: "http",
		Code:    1,
	}
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).
		WithAttributesHTTP(cerror.ErrAttributesValuesHTTP{Code: attrs.Code})

	assert.Equal(t, attrs, err.Attributes())
}

func TestCErrorWithAttributesRedis(t *testing.T) {
	srcErr := fmt.Errorf("err")
	attrs := &cerror.ErrAttributes{
		Type:    "db",
		Subtype: "redis",
		CodeStr: cerror.KindFromRedis(srcErr).String(),
	}
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).
		WithAttributesRedis(cerror.ErrAttributesValuesRedis{Err: srcErr})

	assert.Equal(t, attrs, err.Attributes())
}

func TestCErrorWithAttributesMinio(t *testing.T) {
	srcErr := fmt.Errorf("err")
	attrs := &cerror.ErrAttributes{
		Type:    "mediastorage",
		Subtype: "minio",
		CodeStr: cerror.KindFromMinio(srcErr).String(),
	}
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).
		WithAttributesMinio(cerror.ErrAttributesValuesMinio{Err: srcErr})

	assert.Equal(t, attrs, err.Attributes())
}

func TestCErrorWithAttributesKafka(t *testing.T) {
	srcErr := fmt.Errorf("err")
	attrs := &cerror.ErrAttributes{
		Type:    "msgbroker",
		Subtype: "kafka",
		CodeStr: cerror.KindFromKafka(srcErr).String(),
	}
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).
		WithAttributesKafka(cerror.ErrAttributesValuesKafka{Err: srcErr})

	assert.Equal(t, attrs, err.Attributes())
}

func TestCErrorWithAttributesElastic(t *testing.T) {
	srcErr := fmt.Errorf("err")
	attrs := &cerror.ErrAttributes{
		Type:    "db",
		Subtype: "elastic",
		CodeStr: cerror.KindFromElastic(srcErr).String(),
	}
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).
		WithAttributesElastic(cerror.ErrAttributesValuesElastic{Err: srcErr})

	assert.Equal(t, attrs, err.Attributes())
}

func TestCErrorWithAttributesPostgres(t *testing.T) {
	srcErr := fmt.Errorf("err")
	attrs := &cerror.ErrAttributes{
		Type:    "db",
		Subtype: "postgres",
		CodeStr: cerror.KindFromPostgres(srcErr).String(),
	}
	err := cerror.New(context.Background(), cerror.KindConflict, fmt.Errorf(errText)).
		WithAttributesPostgres(cerror.ErrAttributesValuesPostgres{Err: srcErr})

	assert.Equal(t, attrs, err.Attributes())
}
