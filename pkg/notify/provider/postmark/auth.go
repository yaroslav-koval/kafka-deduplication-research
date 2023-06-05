package postmark

import (
	"kafka-polygon/pkg/converto"
	"net/http"
)

// AuthTransport holds an authentication information for Postmark API.
type AuthTransport struct {
	Transport    http.RoundTripper
	ServerToken  string
	AccountToken *string
}

// RoundTrip implements the RoundTripper interface.
func (t *AuthTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Postmark-Server-Token", t.ServerToken)

	if t.AccountToken != nil {
		req.Header.Add("X-Postmark-Account-Token", converto.StringValue(t.AccountToken))
	}

	return transport.RoundTrip(req)
}
