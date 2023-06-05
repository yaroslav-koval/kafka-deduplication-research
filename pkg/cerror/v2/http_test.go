package cerror_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror/v2"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log"
	"testing"

	"github.com/tj/assert"
)

func TestBuildErrorHTTPResponse(t *testing.T) {
	cErr := &testError{
		err:      fmt.Errorf(errText),
		kind:     cerror.KindConflict,
		ops:      opsText,
		fields:   fields,
		httpCode: 400,
		attrs:    &cerror.ErrAttributes{Type: "type"},
	}
	resp := cerror.BuildErrorHTTPResponse(cErr)
	assert.Equal(t, &cerror.HTTPResponseErrorWrap{
		Error: cerror.HTTPResponseError{
			Message:    cErr.Error(),
			Kind:       cErr.kind.String(),
			Errors:     cErr.fields,
			Attributes: cErr.attrs,
		},
	}, resp, "compare with cerror")

	err := fmt.Errorf(errText)
	resp = cerror.BuildErrorHTTPResponse(err)
	assert.Equal(t, &cerror.HTTPResponseErrorWrap{
		Error: cerror.HTTPResponseError{
			Message: err.Error(),
			Kind:    cerror.KindOther.String(),
		},
	}, resp, "compare with not cerror")
}

func TestLogHTTPHandlerError(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetGlobalOutput(buf)

	cErr := &testError{
		err:      fmt.Errorf(errText),
		ops:      opsText,
		fields:   fields,
		httpCode: 400,
		attrs:    &cerror.ErrAttributes{Type: "type"},
	}

	for k, v := range map[cerror.Kind]string{
		cerror.KindOther:           levelErrorText,
		cerror.KindUnauthenticated: levelErrorText,
		cerror.KindNotExists:       levelWarnText,
		cerror.KindConflict:        levelErrorText,
		cerror.KindSyntax:          levelErrorText,
		cerror.KindValidation:      levelErrorText,
		cerror.KindForbidden:       levelErrorText,
	} {
		cErr.kind = k
		expectedMsg := testLogMsg{
			Level:     v,
			Message:   "http request was processed with error",
			Error:     cErr.Error(),
			ErrorKind: cErr.kind.String(),
			ErrorCode: cErr.httpCode,
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
		err:      fmt.Errorf(errText),
		ops:      opsText,
		fields:   fields,
		httpCode: 400,
		attrs:    &cerror.ErrAttributes{Type: "type"},
	}

	for k, v := range map[cerror.Kind]string{
		cerror.KindOther:           levelErrorText,
		cerror.KindUnauthenticated: levelErrorText,
		cerror.KindNotExists:       levelWarnText,
		cerror.KindConflict:        levelErrorText,
		cerror.KindSyntax:          levelErrorText,
		cerror.KindValidation:      levelErrorText,
		cerror.KindForbidden:       levelErrorText,
	} {
		cErr.kind = k
		expectedMsg := testLogMsg{
			Level:     v,
			Message:   "http request was processed with error",
			Error:     cErr.Error(),
			RequestID: fmt.Sprintf("%s", ctx.Value(consts.HeaderXRequestID)),
			ErrorKind: cErr.kind.String(),
			ErrorCode: cErr.httpCode,
		}

		buf.Reset()
		cerror.LogHTTPHandlerErrorCtx(ctx, cErr)

		var actualMsg testLogMsg

		_ = json.Unmarshal(buf.Bytes(), &actualMsg)

		assert.Equal(t, expectedMsg, actualMsg)
	}
}
