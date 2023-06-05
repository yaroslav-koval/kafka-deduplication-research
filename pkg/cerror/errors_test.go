package cerror_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log"
	"testing"

	"github.com/tj/assert"
)

const (
	errText        = "error"
	opsText        = "ops"
	requestIDText  = "123"
	levelErrorText = "error"
	levelWarnText  = "warn"
	levelFatalText = "fatal"
)

var (
	fields = map[string]interface{}{"name": "John"}
)

type testLogMsg struct {
	Level      string      `json:"level"`
	Message    string      `json:"message"`
	Error      string      `json:"error"`
	RequestID  string      `json:"requestID"`
	ErrorKind  string      `json:"errorKind"`
	ErrorCode  int         `json:"errorCode"`
	Payload    interface{} `json:"payload"`
	Operations []string
}

type testError struct {
	err    error
	kind   cerror.Kind
	ops    string
	fields map[string]interface{}
}

func (e *testError) Error() string {
	return e.err.Error()
}

func (e *testError) Op() string {
	return e.ops
}

func (e *testError) Kind() cerror.Kind {
	return e.kind
}

func (e *testError) Fields() map[string]interface{} {
	return e.fields
}

func TestBuildErrorResponse(t *testing.T) {
	cErr := &testError{
		err:    fmt.Errorf(errText),
		kind:   cerror.KindConflict,
		ops:    opsText,
		fields: fields,
	}
	resp := cerror.BuildErrorResponse(cErr)
	assert.Equal(t, &cerror.ResponseErrorWrap{
		Error: cerror.ResponseError{
			Message: cErr.Error(),
			Type:    cErr.kind.String(),
			Group:   cErr.kind.Group().String(),
			Errors:  cErr.fields,
		},
	}, resp, "compare with cerror")

	err := fmt.Errorf(errText)
	resp = cerror.BuildErrorResponse(err)
	assert.Equal(t, &cerror.ResponseErrorWrap{
		Error: cerror.ResponseError{
			Message: err.Error(),
			Type:    cerror.KindOther.String(),
			Group:   cerror.KindOther.Group().String(),
		},
	}, resp, "compare with not cerror")
}

func TestLogHTTPHandlerError(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetGlobalOutput(buf)

	cErr := &testError{
		err:    fmt.Errorf(errText),
		ops:    opsText,
		fields: fields,
	}

	for k, v := range map[cerror.Kind]string{
		cerror.KindOther:         levelErrorText,
		cerror.KindInvalid:       levelErrorText,
		cerror.KindPermission:    levelErrorText,
		cerror.KindExist:         levelErrorText,
		cerror.KindNotExist:      levelWarnText,
		cerror.KindPrivate:       levelErrorText,
		cerror.KindInternal:      levelErrorText,
		cerror.KindTransient:     levelErrorText,
		cerror.KindBadParams:     levelErrorText,
		cerror.KindBadValidation: levelErrorText,
		cerror.KindInvalidState:  levelErrorText,
		cerror.KindConflict:      levelErrorText,
		cerror.KindForbidden:     levelErrorText,
	} {
		cErr.kind = k
		expectedMsg := testLogMsg{
			Level:     v,
			Message:   "http request was processed with error",
			Error:     cErr.Error(),
			ErrorKind: cErr.kind.String(),
			ErrorCode: cErr.kind.HTTPCode(),
		}

		buf.Reset()
		cerror.LogHTTPHandlerError(cErr)

		var actualMsg testLogMsg

		_ = json.Unmarshal(buf.Bytes(), &actualMsg)

		assert.Equal(t, expectedMsg, actualMsg)
	}
}

func TestLogHTTPHandlerErrorCtx(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetGlobalOutput(buf)

	ctx := context.WithValue(context.Background(), consts.HeaderXRequestID, requestIDText) //nolint:staticcheck
	cErr := &testError{
		err:    fmt.Errorf(errText),
		ops:    opsText,
		fields: fields,
	}

	for k, v := range map[cerror.Kind]string{
		cerror.KindOther:         levelErrorText,
		cerror.KindInvalid:       levelErrorText,
		cerror.KindPermission:    levelErrorText,
		cerror.KindExist:         levelErrorText,
		cerror.KindNotExist:      levelWarnText,
		cerror.KindPrivate:       levelErrorText,
		cerror.KindInternal:      levelErrorText,
		cerror.KindTransient:     levelErrorText,
		cerror.KindBadParams:     levelErrorText,
		cerror.KindBadValidation: levelErrorText,
		cerror.KindInvalidState:  levelErrorText,
		cerror.KindConflict:      levelErrorText,
		cerror.KindForbidden:     levelErrorText,
	} {
		cErr.kind = k
		expectedMsg := testLogMsg{
			Level:     v,
			Message:   "http request was processed with error",
			Error:     cErr.Error(),
			RequestID: fmt.Sprintf("%s", ctx.Value(consts.HeaderXRequestID)),
			ErrorKind: cErr.kind.String(),
			ErrorCode: cErr.kind.HTTPCode(),
		}

		buf.Reset()
		cerror.LogHTTPHandlerErrorCtx(ctx, cErr)

		var actualMsg testLogMsg

		_ = json.Unmarshal(buf.Bytes(), &actualMsg)

		assert.Equal(t, expectedMsg, actualMsg)
	}
}
