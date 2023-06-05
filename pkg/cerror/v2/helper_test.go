package cerror_test

import (
	"fmt"
	"kafka-polygon/pkg/cerror/v2"
	"net/http"
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
	Level      string                `json:"level"`
	Message    string                `json:"message"`
	Error      string                `json:"error"`
	RequestID  string                `json:"requestID"`
	ErrorKind  string                `json:"errorKind"`
	ErrorCode  int                   `json:"errorCode"`
	Attributes *cerror.ErrAttributes `json:"errorAttributes"`
	Payload    interface{}           `json:"payload"`
	Operations []string
}

type testError struct {
	err      error
	kind     cerror.Kind
	ops      string
	fields   map[string]interface{}
	httpCode int
	attrs    *cerror.ErrAttributes
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

func (e *testError) HTTPCode() (int, bool) {
	return e.httpCode, e.httpCode > 0
}

func (e *testError) Attributes() *cerror.ErrAttributes {
	return e.attrs
}

func TestErrKind(t *testing.T) {
	terr := &testError{kind: cerror.KindConflict}
	assert.Equal(t, terr.kind, cerror.ErrKind(terr))

	assert.Equal(t, cerror.KindOther, cerror.ErrKind(fmt.Errorf("err")))
}

func TestIsNotExists(t *testing.T) {
	terr := &testError{kind: cerror.KindConflict}
	assert.False(t, cerror.IsNotExists(terr))

	terr.kind = cerror.KindNotExists
	assert.True(t, cerror.IsNotExists(terr))

	assert.False(t, cerror.IsNotExists(fmt.Errorf("err")))
}

func TestHTTPCode(t *testing.T) {
	terr := &testError{httpCode: 400}
	actCode := cerror.HTTPCode(terr)

	assert.Equal(t, terr.httpCode, actCode)

	terr.httpCode = 0

	for _, v := range []struct {
		kind   cerror.Kind
		status int
	}{
		{cerror.KindValidation, http.StatusUnprocessableEntity},
		{cerror.KindUnauthenticated, http.StatusUnauthorized},
		{cerror.KindForbidden, http.StatusForbidden},
		{cerror.KindExists, http.StatusConflict},
		{cerror.KindConflict, http.StatusConflict},
		{cerror.KindNotExists, http.StatusNotFound},
		{cerror.KindSyntax, http.StatusBadRequest},
		{cerror.KindTimeout, http.StatusGatewayTimeout},
		{cerror.KindTimeout, http.StatusGatewayTimeout},
		{cerror.KindOther, http.StatusInternalServerError},
	} {
		terr.kind = v.kind
		assert.Equal(t, v.status, cerror.HTTPCode(terr))
	}

	assert.Equal(t, http.StatusInternalServerError, cerror.HTTPCode(fmt.Errorf("err")))
}
