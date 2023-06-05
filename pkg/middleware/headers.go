package middleware

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func DefaultHeaderXRequestID(key, value string) string {
	if key == consts.HeaderXRequestID && value == "" {
		value = uuid.NewV4().String()
	}

	return value
}

func ValidateContentTypeJSON(ctx context.Context,
	method, contentType string, excludeMethods ...string) *cerror.CError {
	if !contains(excludeMethods, method) &&
		!strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return cerror.NewF(ctx, cerror.KindBadContentType,
			"content-type header should be %v. Current value: %v",
			"application/json", contentType).LogError()
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
