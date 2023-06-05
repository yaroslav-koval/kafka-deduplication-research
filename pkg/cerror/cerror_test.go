package cerror_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
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
	err := cerror.New(ctx, kind, originalErr).WithPayload(fields).LogError()

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
			ErrorCode:  kind.HTTPCode(),
			Payload:    fields,
			Operations: err.Ops(),
		}

		var actualMsg testLogMsg

		_ = json.Unmarshal(buf.Bytes(), &actualMsg)

		assert.Equal(t, expectedMsg, actualMsg)
	}
}
