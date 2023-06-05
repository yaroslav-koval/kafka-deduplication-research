package postmark

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/notify"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Options struct {
	ServerToken  string
	AccountToken string
	From         string
	To           []string
	TrackOpens   bool
}

type client struct {
	opt     Options
	subject string
	body    string
	html    bool
	tag     string
	pm      *postmark.Client
}

func New(opt Options) notify.ByEmail {
	pm := postmark.NewClient(&http.Client{
		Transport: &AuthTransport{
			ServerToken: opt.ServerToken,
		},
	})

	return &client{
		opt: opt,
		pm:  pm,
	}
}

func (c *client) WithBaseURL(baseURL string) {
	c.pm.BackendURL, _ = url.Parse(baseURL)
}

func (c *client) UID() string {
	return notify.ProviderPostmark
}

func (c *client) From(from string) notify.ByEmail {
	c.opt.From = from
	return c
}

func (c *client) WithSubject(subject string) notify.ByEmail {
	c.subject = subject
	return c
}

func (c *client) UseHTML() notify.ByEmail {
	c.html = true
	return c
}

func (c *client) WithBody(body string) notify.ByEmail {
	c.body = body
	return c
}

func (c *client) WithTag(tag string) notify.ByEmail {
	c.tag = tag
	return c
}

func (c *client) WithTrackOpens() notify.ByEmail {
	c.opt.TrackOpens = true
	return c
}

func (c *client) To(to string, cc ...string) notify.ByEmail {
	c.opt.To = append([]string{to}, cc...)
	return c
}

func (c *client) Send(ctx context.Context) (*notify.MessageResponse, error) {
	if len(c.opt.To) == 0 {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty To").LogError()
	}

	if c.opt.From == "" {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty From").LogError()
	}

	msg := &postmark.Email{
		From:       c.opt.From,
		To:         strings.Join(c.opt.To, ","),
		Subject:    c.subject,
		TrackOpens: c.opt.TrackOpens,
	}

	if c.tag != "" {
		msg.Tag = c.tag
	}

	if c.html {
		msg.HTMLBody = c.body
	} else {
		msg.TextBody = c.body
	}

	msgResp := &notify.MessageResponse{
		SubmittedAt: time.Now().UTC(),
	}

	emailResp, dataResp, err := c.pm.Email.Send(msg)

	if dataResp != nil {
		msgResp.ErrorCode = dataResp.ErrorCode
		msgResp.Message = dataResp.Message
		msgResp.ResponseStatusCode = dataResp.StatusCode
	}

	if err != nil {
		if data, ok := err.(*postmark.ErrorResponse); ok {
			msgResp.ErrorCode = data.ErrorCode
			msgResp.Message = data.Message
		}

		return msgResp, err
	}

	defer dataResp.Body.Close()

	if emailResp != nil {
		msgResp.MessageID = emailResp.MessageID
		msgResp.SubmittedAt = emailResp.SubmittedAt
	}

	return msgResp, nil
}

func (c *client) GetStatus(ctx context.Context, req notify.MessageStatusRequest) (*notify.StatusResponse, error) {
	return &notify.StatusResponse{
		Status:  notify.StatusDelivered,
		Message: "success",
	}, nil
}
