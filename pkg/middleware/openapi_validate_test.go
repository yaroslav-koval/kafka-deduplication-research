package middleware_test

import (
	"context"
	"errors"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/middleware"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	bgCtx       = context.Background()
	manifestErr = "read REST Manifest to understand this error"
)

type testModelSchema struct {
	reversePath []string
	reason      string
}

func (tms *testModelSchema) JSONPointer() []string {
	return tms.reversePath
}

func (tms *testModelSchema) GetReason() string {
	return tms.reason
}

func (tms *testModelSchema) Error() string {
	return ""
}

func TestCValidateRequestError(t *testing.T) {
	tests := []struct {
		reqErr     *openapi3filter.RequestError
		expFields  map[string]interface{}
		expPayload map[string]string
		expErr     string
	}{
		{
			reqErr: &openapi3filter.RequestError{
				Parameter: &openapi3.Parameter{
					In:   "test-in",
					Name: "test-name",
				},
				Reason: "test-reason",
				Err: &openapi3.SchemaError{
					Reason: "test-reason-schema",
				},
			},
			expFields: map[string]interface{}{
				"test-in.test-name": map[string]string{
					"message": "test-reason-schema",
				},
			},
			expPayload: map[string]string{
				"test-in.test-name": "test-reason-schema",
			},
			expErr: "test-in.test-name:test-reason-schema",
		},
		{
			reqErr: &openapi3filter.RequestError{
				Parameter: &openapi3.Parameter{
					In:   "test-in",
					Name: "test-name",
				},
				Reason: "test-reason",
				Err: &openapi3filter.ParseError{
					Reason: "test-reason-parse",
				},
			},
			expFields: map[string]interface{}{
				"path": map[string]string{
					"message": "test-reason-parse",
				},
			},
			expPayload: map[string]string{"path": "test-reason-parse"},
			expErr:     "path:test-reason-parse",
		},
		{
			reqErr: &openapi3filter.RequestError{
				Parameter: &openapi3.Parameter{
					In:   "test-in",
					Name: "test-name",
				},
				Reason: "test-reason",
				Err: openapi3.MultiError{
					&testModelSchema{
						reversePath: []string{"test-key", "test-field"},
						reason:      "test-reason-multi-error",
					},
				},
			},
			expFields: map[string]interface{}{
				"$[test-key].test-field": map[string]string{
					"message": "test-reason-multi-error",
				},
			},
			expPayload: map[string]string{
				"$[test-key].test-field": "test-reason-multi-error",
			},
			expErr: "$[test-key].test-field:test-reason-multi-error",
		},
	}

	for _, item := range tests {
		err := middleware.CValidateRequestError(bgCtx, item.reqErr)
		require.Error(t, err)
		assert.Equal(t, manifestErr, err.Error())
		assert.IsType(t, &cerror.ValidationError{}, err)

		e := err.(*cerror.ValidationError)

		assert.Equal(t, http.StatusUnprocessableEntity, e.Kind().HTTPCode())
		assert.Equal(t, "validation_error", e.Kind().String())
		assert.Equal(t, item.expFields, e.Fields())
		assert.Equal(t, item.expPayload, e.Payload())
		assert.Equal(t, item.expErr, e.Err().Error())
	}
}

func TestCValidateRequestDefaultError(t *testing.T) {
	expDefErr := errors.New("cValidateRequestError err")
	err := middleware.CValidateRequestError(bgCtx, &openapi3filter.RequestError{
		Err: expDefErr,
	})
	require.Error(t, err)
	assert.Equal(t, expDefErr.Error(), err.Error())
}

func TestHandleSchemaError(t *testing.T) {
	he := &middleware.HandlerErr{
		Reason: "test-reason",
	}

	errFields := make(map[string]string)

	he.ReversePath = []string{"test-first"}

	middleware.HandleSchemaError(he, errFields)
	assert.Equal(t, map[string]string{"$.test-first": "test-reason"}, errFields)

	he.ReversePath = []string{"test-first", "test-second"}

	middleware.HandleSchemaError(he, errFields)
	assert.Equal(t, map[string]string{"$.test-first": "test-reason", "$[test-first].test-second": "test-reason"}, errFields)
}

func TestCValidateMultiError(t *testing.T) {
	expFields := map[string]interface{}{
		"$[test-key].test-field": map[string]string{"message": "test-reason-multi-error"},
	}
	expPayload := map[string]string{"$[test-key].test-field": "test-reason-multi-error"}
	err := middleware.CValidateMultiError(bgCtx, openapi3.MultiError{
		&openapi3filter.RequestError{
			Parameter: &openapi3.Parameter{
				In:   "test-in",
				Name: "test-name",
			},
			Err: openapi3.MultiError{
				&testModelSchema{
					reversePath: []string{"test-key", "test-field"},
					reason:      "test-reason-multi-error",
				},
			},
		},
	}, func(ctx context.Context, e error) error {
		return nil
	})

	require.Error(t, err)
	assert.Equal(t, "read REST Manifest to understand this error", err.Error())
	assert.IsType(t, &cerror.ValidationError{}, err)

	e := err.(*cerror.ValidationError)

	assert.Equal(t, http.StatusUnprocessableEntity, e.Kind().HTTPCode())
	assert.Equal(t, "validation_error", e.Kind().String())
	assert.Equal(t, expFields, e.Fields())
	assert.Equal(t, expPayload, e.Payload())
	assert.Equal(t, "$[test-key].test-field:test-reason-multi-error", e.Err().Error())
}

func TestCValidateMultiErrorUseSecurityRequirementsError(t *testing.T) {
	err := middleware.CValidateMultiError(bgCtx, openapi3.MultiError{
		&openapi3filter.SecurityRequirementsError{
			SecurityRequirements: openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"field": []string{
						"test-sec-val",
					},
				},
			},
			Errors: []error{
				errors.New("first err"),
			},
		},
	}, func(ctx context.Context, e error) error {
		return nil
	})
	require.Error(t, err)

	e := err.(*cerror.CError)

	assert.Equal(t, http.StatusForbidden, e.Kind().HTTPCode())
	assert.Equal(t, "forbidden", e.Kind().String())
	assert.Equal(t, "security requirements failed: first err", e.Err().Error())
}

func TestCValidateMultiErrorUseDefaultErr(t *testing.T) {
	expErr := errors.New("exp default multi err")
	err := middleware.CValidateMultiError(bgCtx, openapi3.MultiError{
		expErr,
		expErr,
	}, func(ctx context.Context, e error) error {
		return nil
	})
	require.Error(t, err)

	e := err.(*cerror.CError)

	assert.Equal(t, http.StatusInternalServerError, e.Kind().HTTPCode())
	assert.Equal(t, "internal_error", e.Kind().String())
	assert.Equal(t, "error validating request: exp default multi err | exp default multi err", e.Err().Error())
}
