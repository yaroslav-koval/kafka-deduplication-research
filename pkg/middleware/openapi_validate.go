package middleware

import (
	"context"
	"errors"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

const splitNumber = 2

type HandlerSchemaError interface {
	JSONPointer() []string
	Error() string
}

type HandlerSchema interface {
	HandlerSchemaError
	GetReason() string
}

type HandlerErr struct {
	ReversePath []string
	Reason      string
}

func NewHandlerError(obj interface{}) *HandlerErr {
	he := &HandlerErr{}
	switch v := obj.(type) {
	case *openapi3.SchemaError:
		he.ReversePath = v.JSONPointer()
		he.Reason = v.Reason
	case HandlerSchema:
		he.ReversePath = v.JSONPointer()
		he.Reason = v.GetReason()
	}

	return he
}

type MultiErrFn func(ctx context.Context, e error) error

func CValidateRequestError(c context.Context, reqErr *openapi3filter.RequestError) error {
	errFields, err := collectErrFieldsByType(c, reqErr)
	if err != nil {
		return defaultValidateRequestError(c, err)
	}

	return cerror.NewValidationError(c, errFields).LogError()
}

func collectErrFieldsByType(_ context.Context, reqErr *openapi3filter.RequestError) (
	map[string]string, *openapi3filter.RequestError) {
	unwrap := reqErr.Unwrap()
	errFields := make(map[string]string)

	switch err := unwrap.(type) {
	case *openapi3.SchemaError:
		he := NewHandlerError(unwrap)
		HandleSchemaError(he, errFields)

		if len(he.ReversePath) == 0 {
			if reqErr.Parameter != nil {
				errFields[fmt.Sprintf("%s.%s", reqErr.Parameter.In, reqErr.Parameter.Name)] = he.Reason
			}

			if reqErr.RequestBody != nil {
				errFields["body"] = he.Reason
			}
		}
	case *openapi3filter.ParseError:
		errFields["path"] = err.Reason
	case openapi3.MultiError:
		for _, e := range err {
			he := NewHandlerError(e)
			if he != nil {
				HandleSchemaError(he, errFields)
			} else {
				return nil, reqErr
			}
		}
	default:
		return nil, reqErr
	}

	return errFields, nil
}

func HandleSchemaError(schErr *HandlerErr, errFields map[string]string) {
	if len(schErr.ReversePath) > 0 {
		errKey := fmt.Sprintf("$.%s", schErr.ReversePath[0])
		if len(schErr.ReversePath) > 1 {
			errKey = fmt.Sprintf("$[%s].%s", schErr.ReversePath[0], schErr.ReversePath[1])
		}

		errFields[errKey] = schErr.Reason
	}
}

func defaultValidateRequestError(c context.Context, err *openapi3filter.RequestError) error {
	kind := cerror.KindBadParams
	errorLines := strings.SplitN(err.Error(), "\n", splitNumber)

	if strings.Contains(err.Error(), "header Content-Type has unexpected value") {
		kind = cerror.KindBadContentType
	}

	return cerror.New(c, kind, errors.New(errorLines[0])).LogError()
}

func CValidateMultiError(ctx context.Context, errs openapi3.MultiError, fn MultiErrFn) error {
	errFields := make(map[string]string)

	for i := range errs {
		if fn != nil {
			errFn := fn(ctx, errs[i])
			if errFn != nil {
				return errFn
			}
		}

		switch e := errs[i].(type) {
		case *openapi3filter.RequestError:
			errList, reqErr := collectErrFieldsByType(ctx, e)
			if reqErr != nil {
				return defaultValidateRequestError(ctx, reqErr)
			}

			for j := range errList {
				errFields[j] = errList[j]
			}
		case *openapi3filter.SecurityRequirementsError:
			return cerror.NewF(
				ctx,
				cerror.KindForbidden,
				"%s", e.Error()).LogError()
		default:
			return cerror.NewF(
				ctx,
				cerror.KindInternal,
				"error validating request: %s", errs).LogError()
		}
	}

	return cerror.NewValidationError(ctx, errFields).LogError()
}
