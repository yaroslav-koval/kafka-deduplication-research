package cerror

import (
	"kafka-polygon/pkg/log/logger"
	"net/http"
	"strings"

	"github.com/samber/lo"
)

// ErrKind defines kind based on given error.
func ErrKind(err error) Kind {
	if e, ok := err.(KindError); ok {
		return e.Kind()
	}

	return KindOther
}

// HTTPCode detects appropriate http code for the given error.
func HTTPCode(err error) int {
	if herr, ok := err.(HTTPCodeError); ok {
		if c, ok := herr.HTTPCode(); ok {
			return c
		}
	}

	if kerr, ok := err.(KindError); ok {
		switch kerr.Kind() {
		case KindValidation:
			return http.StatusUnprocessableEntity
		case KindUnauthenticated:
			return http.StatusUnauthorized
		case KindForbidden:
			return http.StatusForbidden
		case KindExists, KindConflict:
			return http.StatusConflict
		case KindNotExists:
			return http.StatusNotFound
		case KindSyntax:
			return http.StatusBadRequest
		case KindTimeout:
			return http.StatusGatewayTimeout
		default:
			return http.StatusInternalServerError
		}
	}

	return http.StatusInternalServerError
}

// IsNotExists detects whether the error is from "not found" family.
func IsNotExists(err error) bool {
	kerr, ok := err.(KindError)

	return ok && kerr.Kind() == KindNotExists
}

func logLevelByKind(kind Kind) logger.Level {
	if kind == KindNotExists {
		return logger.LevelWarn
	}

	return logger.LevelError
}

func isStrContainsInsensitive(s string, values []string) bool {
	sLower := strings.ToLower(s)

	return lo.ContainsBy(values, func(v string) bool {
		return strings.Contains(sLower, strings.ToLower(v))
	})
}
