package mailgun

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/notify"
	"net/http"
	"time"
)

type NotifierClient interface {
	NewMessage(from, subject, text string, to ...string) *mailgun.Message
	Send(ctx context.Context, message *mailgun.Message) (mes string, id string, err error)
	SetAPIBase(address string)
}

type Options struct {
	Domain     string
	APIKey     string
	From       string
	To         []string
	TrackOpens bool
}

type Client struct {
	opt           Options
	mailgunClient NotifierClient

	tag          string
	subject      string
	body         string
	html         bool
	template     string
	templateArgs map[string]string
}

func New(opt Options) *Client {
	return &Client{
		opt:           opt,
		mailgunClient: mailgun.NewMailgun(opt.Domain, opt.APIKey),
	}
}

func (c *Client) WithMailgunClient(mc NotifierClient) *Client {
	c.mailgunClient = mc
	return c
}

func (c *Client) WithBaseURL(baseURL string) {
	c.mailgunClient.SetAPIBase(baseURL)
}

func (c *Client) UID() string {
	return notify.ProviderMailgun
}

func (c *Client) From(from string) notify.ByEmail {
	c.opt.From = from
	return c
}

func (c *Client) WithSubject(subject string) notify.ByEmail {
	c.subject = subject
	return c
}

func (c *Client) UseHTML() notify.ByEmail {
	c.html = true
	return c
}

func (c *Client) WithBody(body string) notify.ByEmail {
	c.body = body
	return c
}

func (c *Client) WithTag(tag string) notify.ByEmail {
	c.tag = tag
	return c
}

func (c *Client) WithTrackOpens() notify.ByEmail {
	c.opt.TrackOpens = true
	return c
}

func (c *Client) To(to string, cc ...string) notify.ByEmail {
	c.opt.To = append([]string{to}, cc...)
	return c
}

func (c *Client) Send(ctx context.Context) (*notify.MessageResponse, error) {
	if len(c.opt.To) == 0 {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty To").LogError()
	}

	if c.opt.From == "" {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty From").LogError()
	}

	message := c.mailgunClient.NewMessage(
		c.opt.From,
		c.subject,
		c.body,
		c.opt.To...,
	)
	message.SetTrackingOpens(c.opt.TrackOpens)

	if c.tag != "" {
		err := message.AddTag(c.tag)
		if err != nil {
			return nil, cerror.NewF(ctx, cerror.KindInternal, "add tag to message: %v", err).LogError()
		}
	}

	err := c.SetBody(ctx, message)
	if err != nil {
		return nil, err
	}

	msgResp := &notify.MessageResponse{}

	respMsg, id, err := c.mailgunClient.Send(ctx, message)
	if err != nil {
		msgResp.Message = err.Error()
		msgResp.ErrorCode = http.StatusInternalServerError

		return msgResp, cerror.NewF(ctx, cerror.KindInternal, "send email via mailgun: %v", err).LogError()
	}

	msgResp.Message = respMsg
	msgResp.MessageID = id
	msgResp.SubmittedAt = time.Now().UTC()

	return msgResp, nil
}

func (c *Client) SetBody(ctx context.Context, msg *mailgun.Message) error {
	if c.html {
		msg.SetHtml(c.body)
	} else if c.template != "" {
		msg.SetTemplate(c.template)

		for i := range c.templateArgs {
			err := msg.AddTemplateVariable(i, c.templateArgs[i])
			if err != nil {
				return cerror.NewF(ctx, cerror.KindInternal, "add variable %s to template %s", i, c.template).LogError()
			}
		}
	}

	return nil
}

func (c *Client) GetStatus(ctx context.Context, req notify.MessageStatusRequest) (*notify.StatusResponse, error) {
	return &notify.StatusResponse{
		Status:  notify.StatusDelivered,
		Message: "success",
	}, nil
}
