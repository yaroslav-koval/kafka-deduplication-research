package tiniyo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/notify"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	mimeApplicationJSON = "application/json"
	FormatDateTime      = "2006-01-02 15:04:05"
)

type HTTPClient interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
	DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
}

type Request struct {
	URL    string
	Method string
	Body   interface{}
}

type Client struct {
	httpClient HTTPClient
	opt        *Options
}

func New(opt *Options) *Client {
	return &Client{
		httpClient: new(fasthttp.Client),
		opt:        opt,
	}
}

func (c *Client) WithHTTPClient(h HTTPClient) *Client {
	c.httpClient = h
	return c
}

func (c *Client) WithBaseURL(hostAPI string) {
	c.opt.HostAPI = hostAPI
}

func (c *Client) UID() string {
	return notify.ProviderTiniyo
}

func (c *Client) From(from string) notify.BySMS {
	c.opt.From = from
	return c
}

func (c *Client) WithBody(body string) notify.BySMS {
	c.opt.Body = body
	return c
}

func (c *Client) To(to string, cc ...string) notify.BySMS {
	c.opt.To = to
	return c
}

func (c *Client) Send(ctx context.Context) (*notify.MessageResponse, error) {
	if c.opt.To == "" {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty To").LogError()
	}

	if c.opt.From == "" {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty From").LogError()
	}

	req := MessageRequestBody{
		From: c.opt.From,
		To:   c.opt.To,
		Body: c.opt.Body,
	}

	r := &Request{
		URL:    fmt.Sprintf(c.opt.HostAPI, c.opt.AuthID),
		Method: fasthttp.MethodPost,
		Body:   req,
	}

	msgResp := &notify.MessageResponse{}

	resp, errResp, err := c.sendRequest(ctx, r)
	if err != nil {
		if data, ok := err.(*cerror.CError); ok {
			msgResp.ResponseStatusCode = data.Kind().HTTPCode()
			msgResp.Message = data.Error()
		}

		return msgResp, err
	}

	if errResp != nil {
		msgResp.ErrorCode = int64(errResp.Code)
		msgResp.Message = errResp.Message
		msgResp.ResponseStatusCode = errResp.Status

		return msgResp, nil
	}

	submittedAt, _ := time.Parse(FormatDateTime, resp.DateCreated)
	if submittedAt.IsZero() {
		submittedAt = time.Now().UTC()
	}

	return &notify.MessageResponse{
		SubmittedAt: submittedAt,
		MessageID:   resp.SID,
	}, nil
}

func (c *Client) GetStatus(ctx context.Context, req notify.MessageStatusRequest) (*notify.StatusResponse, error) {
	return &notify.StatusResponse{
		Status:  notify.StatusDelivered,
		Message: "success",
	}, nil
}

// SendRequest is method to send request to Tiniyo
func (c *Client) sendRequest(ctx context.Context, r *Request) (
	*MessageResponseBody, *MessageErrorResponseBody, error) {
	var err error

	req := fasthttp.AcquireRequest()
	req.Header.Add("Authorization", "Basic "+basicAuth(c.opt.AuthID, c.opt.AuthToken))

	defer fasthttp.ReleaseRequest(req)
	req.Header.SetContentType(mimeApplicationJSON)
	req.Header.SetMethod(r.Method)
	req.SetRequestURI(r.URL)

	if r.Body != nil {
		b, mErr := json.Marshal(r.Body)
		if mErr != nil {
			return nil, nil, cerror.New(ctx, cerror.KindInternal, err).LogError()
		}

		req.SetBody(b)
	}

	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseResponse(resp)

	if c.opt.RequestTimeout != 0 {
		err = c.httpClient.DoTimeout(req, resp, c.opt.RequestTimeout)
	} else {
		err = c.httpClient.Do(req, resp)
	}

	if err != nil {
		return nil, nil, cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	respStatus := resp.StatusCode()
	respBody := resp.Body()

	if !(respStatus >= 200 && respStatus < 300) {
		errResp := new(MessageErrorResponseBody)

		err := json.Unmarshal(respBody, errResp)
		if err != nil {
			return nil, nil, cerror.NewF(ctx, cerror.KindInternal, "unmarshal error body").LogError()
		}

		return nil, errResp, nil
	}

	result := new(MessageResponseBody)
	if err := json.Unmarshal(respBody, result); err != nil {
		return nil, nil, cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	return result, nil, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
