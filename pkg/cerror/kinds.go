package cerror

import (
	"net/http"
)

type Kind interface {
	HTTPCode() int
	String() string
	Group() KindGroup
}

type HTTPKind uint8

const (
	_                  HTTPKind = iota
	KindOther                   // Unclassified error
	KindInvalid                 // KindInvalid operation for this type of item
	KindPermission              // KindPermission denied
	KindExist                   // Item already exists
	KindNotExist                // Item does not exist
	KindPrivate                 // Information withheld
	KindInternal                // KindInternal error or inconsistency
	KindTransient               // A transient error
	KindBadParams               // Data has invalid syntax
	KindBadValidation           // Data is well-formed but some validations failed
	KindInvalidState            // Object has invalid state
	KindConflict                // State or logic conflict
	KindForbidden               // Resource is forbidden
	KindBadContentType          // Invalid content type

	// 4xx http error
	KindMethodNotAllowed             // Method Not Allowed
	KindNotAcceptable                // Not Acceptable
	KindProxyAuthRequired            // Proxy Authentication Required
	KindRequestTimeout               // Request Timeout
	KindGone                         // Gone
	KindLengthRequired               // Length Required
	KindPreconditionFailed           // Precondition Failed
	KindRequestEntityTooLarge        // Request Entity Too Large
	KindRequestURITooLong            // Request URI Too Long
	KindUnsupportedMediaType         // Unsupported Media Type
	KindRequestedRangeNotSatisfiable // Requested Range Not Satisfiable
	KindExpectationFailed            // Expectation Failed
	KindTeapot                       // Teapot
	KindMisdirectedRequest           // Misdirected Request
	KindLocked                       // Locked
	KindFailedDependency             // Failed Dependency
	KindTooEarly                     // Too Early
	KindUpgradeRequired              // Upgrade Required
	KindPreconditionRequired         // Precondition Required
	KindTooManyRequests              // Too Many Requests
	KindRequestHeaderFieldsTooLarge  // Request Header Fields Too Large
	KindUnavailableForLegalReasons   // Unavailable For Legal Reasons

	// 5xx http error
	KindNotImplemented                // Not Implemented
	KindBadGateway                    // Bad Gateway
	KindServiceUnavailable            // Service Unavailable
	KindGatewayTimeout                // Gateway Timeout
	KindHTTPVersionNotSupported       // HTTP Version Not Supported
	KindVariantAlsoNegotiates         // Variant Also Negotiates
	KindInsufficientStorage           // Insufficient Storage
	KindLoopDetected                  // Loop Detected
	KindNotExtended                   // Not Extended
	KindNetworkAuthenticationRequired // Network Authentication Required
	KindWebServerIsDown               // Web Server Is Down
	KindConnectionTimedOut            // 522 Connection Timed Out
	KindOriginIsUnreachable           // 523 Origin Is Unreachable
	KindTimeoutOccurred               // 524 A Timeout Occurred

	StatusWebServerIsDown     = 521
	StatusConnectionTimedOut  = 522
	StatusOriginIsUnreachable = 523
	StatusTimeoutOccurred     = 524

	other                         = "other_error" // king message
	invalidOperation              = "invalid_operation"
	permissionDenied              = "permission_denied"
	forbidden                     = "forbidden"
	alreadyExists                 = "already_exists"
	notExist                      = "not_exist"
	thisIsPrivate                 = "this_is_private"
	internal                      = "internal_error"
	transient                     = "transient_error"
	badParams                     = "bad_params"
	validation                    = "validation_error"
	conflict                      = "conflict"
	invalidState                  = "invalid_state"
	badContentType                = "invalid_content_type"
	unknown                       = "unknown_error_kind"
	methodNotAllowed              = "method_not_allowed"
	notAcceptable                 = "not_acceptable"
	proxyAuthRequired             = "proxy_auth_required"
	requestTimeout                = "request_timeout"
	gone                          = "gone"
	lengthRequired                = "length_required"
	preconditionFailed            = "precondition_failed"
	requestEntityTooLarge         = "request_entity_too_large"
	requestURITooLong             = "request_uri_too_long"
	unsupportedMediaType          = "unsupported_media_type"
	requestedRangeNotSatisfiable  = "requested_range_not_satisfiable"
	expectationFailed             = "expectation_failed"
	teapot                        = "teapot"
	notImplemented                = "not_implemented"
	misdirectedRequest            = "misdirected_request"
	locked                        = "locked"
	failedDependency              = "failed_dependency"
	tooEarly                      = "too_early"
	upgradeRequired               = "upgrade_required"
	preconditionRequired          = "precondition_required"
	tooManyRequests               = "too_many_requests"
	requestHeaderFieldsTooLarge   = "request_header_fields_too_large"
	unavailableForLegalReasons    = "unavailable_for_legal_reasons"
	badGateway                    = "bad_gateway"
	serviceUnavailable            = "service_unavailable"
	gatewayTimeout                = "gateway_timeout"
	httpVersionNotSupported       = "http_version_not_supported"
	variantAlsoNegotiates         = "variant_also_negotiates"
	insufficientStorage           = "insufficient_storage"
	loopDetected                  = "loop_detected"
	notExtended                   = "not_extended"
	networkAuthenticationRequired = "network_authentication_required"
	webServerIsDown               = "web_server_is_down"
	connectionTimedOut            = "connection_timed_out"
	originIsUnreachable           = "origin_is_unreachable"
	timeoutOccurred               = "timeout_occurred"
)

// String returns a string representation of the underlying kind.
//
//nolint:gocyclo,funlen
func (k HTTPKind) String() string {
	switch k {
	case KindOther:
		return other
	case KindInvalid:
		return invalidOperation
	case KindPermission:
		return permissionDenied
	case KindForbidden:
		return forbidden
	case KindExist:
		return alreadyExists
	case KindNotExist:
		return notExist
	case KindPrivate:
		return thisIsPrivate
	case KindInternal:
		return internal
	case KindTransient:
		return transient
	case KindBadParams:
		return badParams
	case KindBadValidation:
		return validation
	case KindConflict:
		return conflict
	case KindBadContentType:
		return badContentType
	case KindInvalidState:
		return invalidState
	case KindMethodNotAllowed:
		return methodNotAllowed
	case KindNotAcceptable:
		return notAcceptable
	case KindProxyAuthRequired:
		return proxyAuthRequired
	case KindRequestTimeout:
		return requestTimeout
	case KindGone:
		return gone
	case KindLengthRequired:
		return lengthRequired
	case KindPreconditionFailed:
		return preconditionFailed
	case KindRequestEntityTooLarge:
		return requestEntityTooLarge
	case KindRequestURITooLong:
		return requestURITooLong
	case KindUnsupportedMediaType:
		return unsupportedMediaType
	case KindRequestedRangeNotSatisfiable:
		return requestedRangeNotSatisfiable
	case KindExpectationFailed:
		return expectationFailed
	case KindTeapot:
		return teapot
	case KindNotImplemented:
		return notImplemented
	case KindMisdirectedRequest:
		return misdirectedRequest
	case KindLocked:
		return locked
	case KindFailedDependency:
		return failedDependency
	case KindTooEarly:
		return tooEarly
	case KindUpgradeRequired:
		return upgradeRequired
	case KindPreconditionRequired:
		return preconditionRequired
	case KindTooManyRequests:
		return tooManyRequests
	case KindRequestHeaderFieldsTooLarge:
		return requestHeaderFieldsTooLarge
	case KindUnavailableForLegalReasons:
		return unavailableForLegalReasons
	case KindBadGateway:
		return badGateway
	case KindServiceUnavailable:
		return serviceUnavailable
	case KindGatewayTimeout:
		return gatewayTimeout
	case KindHTTPVersionNotSupported:
		return httpVersionNotSupported
	case KindVariantAlsoNegotiates:
		return variantAlsoNegotiates
	case KindInsufficientStorage:
		return insufficientStorage
	case KindLoopDetected:
		return loopDetected
	case KindNotExtended:
		return notExtended
	case KindNetworkAuthenticationRequired:
		return networkAuthenticationRequired
	case KindWebServerIsDown:
		return webServerIsDown
	case KindConnectionTimedOut:
		return connectionTimedOut
	case KindOriginIsUnreachable:
		return originIsUnreachable
	case KindTimeoutOccurred:
		return timeoutOccurred
	}

	return unknown
}

// HTTPCode returns http code representation of the underlying kind.
func (k HTTPKind) HTTPCode() int {
	var code int

	if k == KindOther {
		return http.StatusInternalServerError
	} else if code = kindTo4xxHTTPCode(k); code != -1 {
		return code
	} else if code = kindTo5xxHTTPCode(k); code != -1 {
		return code
	}

	return http.StatusInternalServerError
}

// Group returns kind group
func (k HTTPKind) Group() KindGroup {
	return GroupHTTP
}

// KindFromHTTPCode converts HTTP code to kind.
//
//nolint:funlen,gocyclo
func KindFromHTTPCode(code int) HTTPKind {
	switch code {
	case http.StatusBadRequest:
		return KindBadParams
	case http.StatusUnauthorized:
		return KindPermission
	case http.StatusNotFound:
		return KindNotExist
	case http.StatusUnprocessableEntity:
		return KindBadValidation
	case http.StatusConflict:
		return KindConflict
	case http.StatusForbidden:
		return KindForbidden
	case http.StatusUnsupportedMediaType:
		return KindBadContentType
	case http.StatusMethodNotAllowed:
		return KindMethodNotAllowed
	case http.StatusNotAcceptable:
		return KindNotAcceptable
	case http.StatusProxyAuthRequired:
		return KindProxyAuthRequired
	case http.StatusRequestTimeout:
		return KindRequestTimeout
	case http.StatusGone:
		return KindGone
	case http.StatusLengthRequired:
		return KindLengthRequired
	case http.StatusPreconditionFailed:
		return KindPreconditionFailed
	case http.StatusRequestEntityTooLarge:
		return KindRequestEntityTooLarge
	case http.StatusRequestURITooLong:
		return KindRequestURITooLong
	case http.StatusRequestedRangeNotSatisfiable:
		return KindRequestedRangeNotSatisfiable
	case http.StatusExpectationFailed:
		return KindExpectationFailed
	case http.StatusTeapot:
		return KindTeapot
	case http.StatusMisdirectedRequest:
		return KindMisdirectedRequest
	case http.StatusLocked:
		return KindLocked
	case http.StatusFailedDependency:
		return KindFailedDependency
	case http.StatusTooEarly:
		return KindTooEarly
	case http.StatusUpgradeRequired:
		return KindUpgradeRequired
	case http.StatusPreconditionRequired:
		return KindPreconditionRequired
	case http.StatusTooManyRequests:
		return KindTooManyRequests
	case http.StatusRequestHeaderFieldsTooLarge:
		return KindRequestHeaderFieldsTooLarge
	case http.StatusUnavailableForLegalReasons:
		return KindUnavailableForLegalReasons
	case http.StatusNotImplemented:
		return KindNotImplemented
	case http.StatusBadGateway:
		return KindBadGateway
	case http.StatusServiceUnavailable:
		return KindServiceUnavailable
	case http.StatusGatewayTimeout:
		return KindGatewayTimeout
	case http.StatusHTTPVersionNotSupported:
		return KindHTTPVersionNotSupported
	case http.StatusVariantAlsoNegotiates:
		return KindVariantAlsoNegotiates
	case http.StatusInsufficientStorage:
		return KindInsufficientStorage
	case http.StatusLoopDetected:
		return KindLoopDetected
	case http.StatusNotExtended:
		return KindNotExtended
	case http.StatusNetworkAuthenticationRequired:
		return KindNetworkAuthenticationRequired
	case StatusWebServerIsDown:
		return KindWebServerIsDown
	case StatusConnectionTimedOut:
		return KindConnectionTimedOut
	case StatusOriginIsUnreachable:
		return KindOriginIsUnreachable
	case StatusTimeoutOccurred:
		return KindTimeoutOccurred
	default:
		return KindOther
	}
}

// ErrKind defines kind based on given error.
func ErrKind(err error) Kind {
	if e, ok := err.(KindError); ok {
		return e.Kind()
	}

	return KindOther
}

//nolint:funlen,gocyclo
func kindTo4xxHTTPCode(k HTTPKind) int {
	switch k {
	case KindInvalid, KindBadParams:
		return http.StatusBadRequest
	case KindPermission:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindExist, KindInvalidState, KindConflict:
		return http.StatusConflict
	case KindNotExist:
		return http.StatusNotFound
	case KindBadValidation:
		return http.StatusUnprocessableEntity
	case KindMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case KindNotAcceptable:
		return http.StatusNotAcceptable
	case KindProxyAuthRequired:
		return http.StatusProxyAuthRequired
	case KindRequestTimeout:
		return http.StatusRequestTimeout
	case KindGone:
		return http.StatusGone
	case KindLengthRequired:
		return http.StatusLengthRequired
	case KindPreconditionFailed:
		return http.StatusPreconditionFailed
	case KindRequestEntityTooLarge:
		return http.StatusRequestEntityTooLarge
	case KindRequestURITooLong:
		return http.StatusRequestURITooLong
	case KindUnsupportedMediaType, KindBadContentType:
		return http.StatusUnsupportedMediaType
	case KindRequestedRangeNotSatisfiable:
		return http.StatusRequestedRangeNotSatisfiable
	case KindExpectationFailed:
		return http.StatusExpectationFailed
	case KindTeapot:
		return http.StatusTeapot
	case KindMisdirectedRequest:
		return http.StatusMisdirectedRequest
	case KindLocked:
		return http.StatusLocked
	case KindFailedDependency:
		return http.StatusFailedDependency
	case KindTooEarly:
		return http.StatusTooEarly
	case KindUpgradeRequired:
		return http.StatusUpgradeRequired
	case KindPreconditionRequired:
		return http.StatusPreconditionRequired
	case KindTooManyRequests:
		return http.StatusTooManyRequests
	case KindRequestHeaderFieldsTooLarge:
		return http.StatusRequestHeaderFieldsTooLarge
	case KindUnavailableForLegalReasons:
		return http.StatusUnavailableForLegalReasons
	default:
		return -1
	}
}

func kindTo5xxHTTPCode(k HTTPKind) int {
	switch k {
	case KindNotImplemented:
		return http.StatusNotImplemented
	case KindBadGateway:
		return http.StatusBadGateway
	case KindServiceUnavailable:
		return http.StatusServiceUnavailable
	case KindGatewayTimeout:
		return http.StatusGatewayTimeout
	case KindHTTPVersionNotSupported:
		return http.StatusHTTPVersionNotSupported
	case KindVariantAlsoNegotiates:
		return http.StatusVariantAlsoNegotiates
	case KindInsufficientStorage:
		return http.StatusInsufficientStorage
	case KindLoopDetected:
		return http.StatusLoopDetected
	case KindNotExtended:
		return http.StatusNotExtended
	case KindNetworkAuthenticationRequired:
		return http.StatusNetworkAuthenticationRequired
	case KindWebServerIsDown:
		return StatusWebServerIsDown
	case KindConnectionTimedOut:
		return StatusConnectionTimedOut
	case KindOriginIsUnreachable:
		return StatusOriginIsUnreachable
	case KindTimeoutOccurred:
		return StatusTimeoutOccurred
	default:
		return -1
	}
}
