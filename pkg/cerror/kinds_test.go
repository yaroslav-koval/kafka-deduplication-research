package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"testing"

	"github.com/tj/assert"
)

type expectedKind struct {
	kind         cerror.Kind
	kindString   string
	kindHTTPCode int
}

var expectedKindTable = []expectedKind{
	{cerror.KindOther, "other_error", http.StatusInternalServerError},
	{cerror.KindInvalid, "invalid_operation", http.StatusBadRequest},
	{cerror.KindPermission, "permission_denied", http.StatusUnauthorized},
	{cerror.KindExist, "already_exists", http.StatusConflict},
	{cerror.KindNotExist, "not_exist", http.StatusNotFound},
	{cerror.KindPrivate, "this_is_private", http.StatusInternalServerError},
	{cerror.KindInternal, "internal_error", http.StatusInternalServerError},
	{cerror.KindTransient, "transient_error", http.StatusInternalServerError},
	{cerror.KindBadParams, "bad_params", http.StatusBadRequest},
	{cerror.KindBadValidation, "validation_error", 422},
	{cerror.KindInvalidState, "invalid_state", http.StatusConflict},
	{cerror.KindConflict, "conflict", http.StatusConflict},
	{cerror.KindForbidden, "forbidden", http.StatusForbidden},
	{cerror.KindBadContentType, "invalid_content_type", http.StatusUnsupportedMediaType},
	{cerror.KindMethodNotAllowed, "method_not_allowed", http.StatusMethodNotAllowed},
	{cerror.KindNotAcceptable, "not_acceptable", http.StatusNotAcceptable},
	{cerror.KindProxyAuthRequired, "proxy_auth_required", http.StatusProxyAuthRequired},
	{cerror.KindRequestTimeout, "request_timeout", http.StatusRequestTimeout},
	{cerror.KindGone, "gone", http.StatusGone},
	{cerror.KindLengthRequired, "length_required", http.StatusLengthRequired},
	{cerror.KindPreconditionFailed, "precondition_failed", http.StatusPreconditionFailed},
	{cerror.KindRequestEntityTooLarge, "request_entity_too_large", http.StatusRequestEntityTooLarge},
	{cerror.KindRequestURITooLong, "request_uri_too_long", http.StatusRequestURITooLong},
	{cerror.KindUnsupportedMediaType, "unsupported_media_type", http.StatusUnsupportedMediaType},
	{cerror.KindRequestedRangeNotSatisfiable, "requested_range_not_satisfiable",
		http.StatusRequestedRangeNotSatisfiable},
	{cerror.KindExpectationFailed, "expectation_failed", http.StatusExpectationFailed},
	{cerror.KindTeapot, "teapot", http.StatusTeapot},
	{cerror.KindMisdirectedRequest, "misdirected_request", http.StatusMisdirectedRequest},
	{cerror.KindLocked, "locked", http.StatusLocked},
	{cerror.KindFailedDependency, "failed_dependency", http.StatusFailedDependency},
	{cerror.KindTooEarly, "too_early", http.StatusTooEarly},
	{cerror.KindUpgradeRequired, "upgrade_required", http.StatusUpgradeRequired},
	{cerror.KindPreconditionRequired, "precondition_required", http.StatusPreconditionRequired},
	{cerror.KindTooManyRequests, "too_many_requests", http.StatusTooManyRequests},
	{cerror.KindRequestHeaderFieldsTooLarge, "request_header_fields_too_large",
		http.StatusRequestHeaderFieldsTooLarge},
	{cerror.KindUnavailableForLegalReasons, "unavailable_for_legal_reasons",
		http.StatusUnavailableForLegalReasons},
	{cerror.KindNotImplemented, "not_implemented", http.StatusNotImplemented},
	{cerror.KindBadGateway, "bad_gateway", http.StatusBadGateway},
	{cerror.KindServiceUnavailable, "service_unavailable", http.StatusServiceUnavailable},
	{cerror.KindGatewayTimeout, "gateway_timeout", http.StatusGatewayTimeout},
	{cerror.KindHTTPVersionNotSupported, "http_version_not_supported",
		http.StatusHTTPVersionNotSupported},
	{cerror.KindVariantAlsoNegotiates, "variant_also_negotiates", http.StatusVariantAlsoNegotiates},
	{cerror.KindInsufficientStorage, "insufficient_storage", http.StatusInsufficientStorage},
	{cerror.KindLoopDetected, "loop_detected", http.StatusLoopDetected},
	{cerror.KindNotExtended, "not_extended", http.StatusNotExtended},
	{cerror.KindNetworkAuthenticationRequired, "network_authentication_required",
		http.StatusNetworkAuthenticationRequired},
	{cerror.KindWebServerIsDown, "web_server_is_down", cerror.StatusWebServerIsDown},
	{cerror.KindConnectionTimedOut, "connection_timed_out", cerror.StatusConnectionTimedOut},
	{cerror.KindOriginIsUnreachable, "origin_is_unreachable", cerror.StatusOriginIsUnreachable},
	{cerror.KindTimeoutOccurred, "timeout_occurred", cerror.StatusTimeoutOccurred},
}

func TestKindString(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}

func TestKindHTTPCode(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindTable {
		assert.Equal(t, v.kindHTTPCode, v.kind.HTTPCode())
	}
}

type kindErr struct {
	kind cerror.Kind
}

func (ke *kindErr) Kind() cerror.Kind {
	return ke.kind
}

func (ke *kindErr) Error() string {
	return ke.kind.String()
}

func TestErrKind(t *testing.T) {
	t.Parallel()

	errKind := cerror.ErrKind(errors.New("other error"))
	assert.NotNil(t, errKind)
	assert.Equal(t, cerror.KindOther, errKind)

	expErrKind := &kindErr{
		kind: cerror.KindInternal,
	}

	errKind = cerror.ErrKind(expErrKind)
	assert.NotNil(t, errKind)
	assert.Equal(t, cerror.KindInternal, errKind)
}

var expErrFromHTTPCode = []struct {
	code int
	kind cerror.Kind
}{
	{
		code: http.StatusBadRequest,
		kind: cerror.KindBadParams,
	},
	{
		code: http.StatusUnauthorized,
		kind: cerror.KindPermission,
	},
	{
		code: http.StatusNotFound,
		kind: cerror.KindNotExist,
	},
	{
		code: http.StatusUnprocessableEntity,
		kind: cerror.KindBadValidation,
	},
	{
		code: http.StatusConflict,
		kind: cerror.KindConflict,
	},
	{
		code: http.StatusForbidden,
		kind: cerror.KindForbidden,
	},
	{
		code: http.StatusUnsupportedMediaType,
		kind: cerror.KindBadContentType,
	},
	{
		code: http.StatusMethodNotAllowed,
		kind: cerror.KindMethodNotAllowed,
	},
	{
		code: http.StatusNotAcceptable,
		kind: cerror.KindNotAcceptable,
	},
	{
		code: http.StatusProxyAuthRequired,
		kind: cerror.KindProxyAuthRequired,
	},
	{
		code: http.StatusRequestTimeout,
		kind: cerror.KindRequestTimeout,
	},
	{
		code: http.StatusGone,
		kind: cerror.KindGone,
	},
	{
		code: http.StatusLengthRequired,
		kind: cerror.KindLengthRequired,
	},
	{
		code: http.StatusPreconditionFailed,
		kind: cerror.KindPreconditionFailed,
	},
	{
		code: http.StatusRequestEntityTooLarge,
		kind: cerror.KindRequestEntityTooLarge,
	},
	{
		code: http.StatusRequestURITooLong,
		kind: cerror.KindRequestURITooLong,
	},
	{
		code: http.StatusRequestedRangeNotSatisfiable,
		kind: cerror.KindRequestedRangeNotSatisfiable,
	},
	{
		code: http.StatusExpectationFailed,
		kind: cerror.KindExpectationFailed,
	},
	{
		code: http.StatusTeapot,
		kind: cerror.KindTeapot,
	},
	{
		code: http.StatusMisdirectedRequest,
		kind: cerror.KindMisdirectedRequest,
	},
	{
		code: http.StatusLocked,
		kind: cerror.KindLocked,
	},
	{
		code: http.StatusFailedDependency,
		kind: cerror.KindFailedDependency,
	},
	{
		code: http.StatusTooEarly,
		kind: cerror.KindTooEarly,
	},
	{
		code: http.StatusUpgradeRequired,
		kind: cerror.KindUpgradeRequired,
	},
	{
		code: http.StatusPreconditionRequired,
		kind: cerror.KindPreconditionRequired,
	},
	{
		code: http.StatusTooManyRequests,
		kind: cerror.KindTooManyRequests,
	},
	{
		code: http.StatusRequestHeaderFieldsTooLarge,
		kind: cerror.KindRequestHeaderFieldsTooLarge,
	},
	{
		code: http.StatusUnavailableForLegalReasons,
		kind: cerror.KindUnavailableForLegalReasons,
	},
	{
		code: http.StatusNotImplemented,
		kind: cerror.KindNotImplemented,
	},
	{
		code: http.StatusBadGateway,
		kind: cerror.KindBadGateway,
	},
	{
		code: http.StatusServiceUnavailable,
		kind: cerror.KindServiceUnavailable,
	},
	{
		code: http.StatusGatewayTimeout,
		kind: cerror.KindGatewayTimeout,
	},
	{
		code: http.StatusHTTPVersionNotSupported,
		kind: cerror.KindHTTPVersionNotSupported,
	},
	{
		code: http.StatusVariantAlsoNegotiates,
		kind: cerror.KindVariantAlsoNegotiates,
	},
	{
		code: http.StatusInsufficientStorage,
		kind: cerror.KindInsufficientStorage,
	},
	{
		code: http.StatusLoopDetected,
		kind: cerror.KindLoopDetected,
	},
	{
		code: http.StatusNotExtended,
		kind: cerror.KindNotExtended,
	},
	{
		code: http.StatusNetworkAuthenticationRequired,
		kind: cerror.KindNetworkAuthenticationRequired,
	},
	{
		code: cerror.StatusWebServerIsDown,
		kind: cerror.KindWebServerIsDown,
	},
	{
		code: cerror.StatusConnectionTimedOut,
		kind: cerror.KindConnectionTimedOut,
	},
	{
		code: cerror.StatusOriginIsUnreachable,
		kind: cerror.KindOriginIsUnreachable,
	},
	{
		code: cerror.StatusTimeoutOccurred,
		kind: cerror.KindTimeoutOccurred,
	},
	{
		code: http.StatusInternalServerError,
		kind: cerror.KindOther,
	},
}

func TestKindFromHTTPCode(t *testing.T) {
	t.Parallel()

	for _, item := range expErrFromHTTPCode {
		actualKind := cerror.KindFromHTTPCode(item.code)
		assert.Equal(t, item.kind, actualKind)
	}
}
