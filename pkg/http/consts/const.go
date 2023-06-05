package consts

const (
	// HTTP headers
	HeaderXRequestID = "X-Request-ID"
	HeaderXClientID  = "X-Client-ID"
	HeaderXUserID    = "X-User-ID"

	// BackendAppName application name
	BackendAppName = "wasfaty"

	WorkflowRoutesPrefix = "/workflows"
)

// RequestHeadersToSave returns list of request headers that should be saved to context.
func RequestHeadersToSave() []string {
	return []string{
		HeaderXRequestID,
		HeaderXClientID,
		HeaderXUserID,
	}
}

// ResponseHeadersToSend returns list of response headers that should be set.
func ResponseHeadersToSend() []string {
	return []string{HeaderXRequestID}
}
