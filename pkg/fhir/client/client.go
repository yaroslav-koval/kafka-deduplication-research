package client

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	pkgFHIR "kafka-polygon/pkg/fhir"
	"time"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/edenlabllc/go-fhir-adapter/model/fhir/consts"
	"github.com/valyala/fasthttp"
)

const (
	headerXConsumerID               = "x-consumer-id"
	headerPreferHandling            = "prefer_handling"
	headerPreferHandlingValueStrict = "strict"
	headerPrefer                    = "prefer"
	headerPreferValueRespondSync    = "respond-sync"

	mimeApplicationJSON = "application/json"
)

type PatchBody struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type Request struct {
	URL         string
	Method      string
	QueryParams []*QParam
	Headers     map[string]string
	Body        interface{}
	Destination interface{}
}

type QParam struct {
	Key   string
	Value string
}

type HTTPClient interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
	DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
}

type Client struct {
	cfg        *env.FHIR
	httpClient HTTPClient
}

func New(cfg *env.FHIR) *Client {
	return &Client{cfg: cfg, httpClient: new(fasthttp.Client)}
}

func (c *Client) WithHTTPClient(h HTTPClient) *Client {
	c.httpClient = h
	return c
}

// GetResourceByID searches resource with given resource name and id
// and saves it into dst if resource was found.
// If resource was not found, returns not found error.
func (c *Client) GetResourceByID(ctx context.Context, resName string, id fhir.ID, dst any) error {
	r := &Request{
		URL:         fmt.Sprintf("%s/%s/%s", c.cfg.HostSearch, resName, id),
		Method:      fasthttp.MethodGet,
		Destination: dst,
	}

	if err := c.SendRequest(ctx, r); err != nil {
		return err
	}

	return nil
}

// SearchResourceByParams searches resources by given query parameters
// and returns result as bundle with found entries.
// If resources were not found, doesn't return error.
// Bundle with empty entries list will be returned instead.
func (c *Client) SearchResourceByParams(ctx context.Context, resName string, params []*QParam) (
	*fhir.Bundle, error) {
	dst := new(fhir.Bundle)
	r := &Request{
		URL:         fmt.Sprintf("%s/%s", c.cfg.HostSearch, resName),
		Method:      fasthttp.MethodGet,
		QueryParams: params,
		Headers:     map[string]string{headerPreferHandling: headerPreferHandlingValueStrict},
		Destination: dst,
	}

	if err := c.SendRequest(ctx, r); err != nil {
		return nil, err
	}

	return dst, nil
}

// CreateBundle creates bundle of given resources.
// It uses root FHIR route instead of /Bundle.
// This means it creates bundle and each of given resources
// rather than just Bundle resource.
func (c *Client) CreateBundle(ctx context.Context, b *fhir.Bundle) (*fhir.Bundle, error) {
	dst := new(fhir.Bundle)
	r := &Request{
		URL:         c.cfg.HostAPI,
		Method:      fasthttp.MethodPost,
		Headers:     map[string]string{headerPrefer: headerPreferValueRespondSync},
		Body:        b,
		Destination: dst,
	}

	if err := c.SendRequest(ctx, r); err != nil {
		return nil, err
	}

	return dst, nil
}

// CreateResource creates resource by resName
// with given body
func (c *Client) CreateResource(ctx context.Context, res pkgFHIR.Resource) error {
	req := &Request{
		URL:    fmt.Sprintf("%s/%s", c.cfg.HostAPI, res.ResourceName()),
		Method: fasthttp.MethodPost,
		Body:   res,
	}

	if err := c.SendRequest(ctx, req); err != nil {
		return err
	}

	return nil
}

// PatchResource updates resource by id and resName
// with given body
func (c *Client) PatchResource(ctx context.Context, resName string, id fhir.ID, body []*PatchBody) error {
	r := &Request{
		URL:    fmt.Sprintf("%s/%s/%s", c.cfg.HostAPI, resName, id),
		Method: fasthttp.MethodPatch,
		Body:   body,
		Headers: map[string]string{
			"Content-Type": "application/json-patch+json",
		},
	}

	if err := c.SendRequest(ctx, r); err != nil {
		return err
	}

	return nil
}

// PutResource updates resource by id and resName
// with given body
func (c *Client) PutResource(ctx context.Context, body pkgFHIR.Resource, dst interface{}) error {
	r := &Request{
		URL:         fmt.Sprintf("%s/%s/%s", c.cfg.HostAPI, body.ResourceName(), body.ResourceID()),
		Method:      fasthttp.MethodPut,
		Body:        body,
		Destination: dst,
		Headers: map[string]string{
			"Content-Type": mimeApplicationJSON,
		},
	}

	return c.SendRequest(ctx, r)
}

// ValidateResource validates a given FHIR resource.
// Returns validation error when at least one issue with
// error severity was found.
func (c *Client) ValidateResource(ctx context.Context, res pkgFHIR.Resource) (*fhir.OperationOutcome, error) {
	dst := new(fhir.OperationOutcome)
	r := &Request{
		URL:         fmt.Sprintf("%s/%s/$validate", c.cfg.HostAPI, res.ResourceName()),
		Method:      fasthttp.MethodPost,
		Body:        res,
		Destination: dst,
	}

	if err := c.SendRequest(ctx, r); err != nil {
		return nil, err
	}

	for _, i := range dst.Issue {
		if i.Severity != nil && i.Severity.String() == consts.IssueSeverityError {
			return nil, cerror.NewF(ctx, cerror.KindBadValidation, "FHIR validation issue").
				WithPayload(dst).
				LogError()
		}
	}

	return dst, nil
}

// SendRequest is generic method to send custom request to FHIR.
// It adds only FHIR required options (such as content-type, user-id headers).
// Other request parts can be set by passing them in Request parameter.
func (c *Client) SendRequest(ctx context.Context, r *Request) error {
	var err error

	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseRequest(req)
	req.Header.SetContentType(mimeApplicationJSON)
	req.Header.SetMethod(r.Method)
	req.Header.Set(headerXConsumerID, c.cfg.User)

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	req.SetRequestURI(r.URL)

	for _, p := range r.QueryParams {
		req.URI().QueryArgs().Add(p.Key, p.Value)
	}

	if r.Body != nil {
		b, mErr := json.Marshal(r.Body)
		if mErr != nil {
			return cerror.New(ctx, cerror.KindInternal, err).LogError()
		}

		req.SetBody(b)
	}

	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseResponse(resp)

	if c.cfg.RequestTimeout != 0 {
		err = c.httpClient.DoTimeout(req, resp, c.cfg.RequestTimeout)
	} else {
		err = c.httpClient.Do(req, resp)
	}

	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	respStatus := resp.StatusCode()
	respBody := resp.Body()

	if !(respStatus >= 200 && respStatus < 300) {
		var payload interface{}

		ar := new(fhir.OperationOutcome)

		if err := json.Unmarshal(respBody, ar); err == nil {
			payload = ar
		} else {
			payload = string(respBody)
		}

		return cerror.NewF(ctx, cerror.KindFromHTTPCode(respStatus), "FHIR unsuccessful response. Code: %d", respStatus).
			WithPayload(payload).
			LogError()
	}

	if r.Destination != nil {
		if err := json.Unmarshal(respBody, r.Destination); err != nil {
			return cerror.New(ctx, cerror.KindInternal, err).LogError()
		}
	}

	return nil
}
