package twilio

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/notify"
	"net/http"
	"time"

	twilioClient "github.com/twilio/twilio-go/client"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Options struct {
	AccountSID        string
	AuthToken         string
	MessageServiceSID string
	From              string
	To                []string
	CallbackURL       string
	BaseClient        twilioClient.BaseClient
}

type client struct {
	opt  *Options
	tw   *twilio.RestClient
	body string
}

func New(opt *Options) notify.BySMS {
	cl := &client{
		opt: &Options{},
	}

	if opt == nil {
		return cl
	}

	cl.opt = opt
	cl.tw = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: opt.AccountSID,
		Password: opt.AuthToken,
		Client:   opt.BaseClient,
	})

	return cl
}

func (c *client) WithMessageServiceSID(msSID string) {
	c.opt.MessageServiceSID = msSID
}

func (c *client) UID() string {
	return notify.ProviderTwilio
}

func (c *client) From(from string) notify.BySMS {
	c.opt.From = from

	return c
}

func (c *client) WithBody(body string) notify.BySMS {
	c.body = body
	return c
}

func (c *client) To(to string, cc ...string) notify.BySMS {
	c.opt.To = append([]string{to}, cc...)
	return c
}

func (c *client) Send(ctx context.Context) (*notify.MessageResponse, error) {
	if c.tw == nil {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not set twilio client").LogError()
	}

	if len(c.opt.To) == 0 {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty To").LogError()
	}

	if c.opt.From == "" {
		return nil, cerror.NewF(ctx, cerror.KindBadValidation, "not empty From").LogError()
	}

	msgResp := &notify.MessageResponse{
		SubmittedAt:        time.Now().UTC(),
		ResponseStatusCode: http.StatusOK,
	}

	params := &openapi.CreateMessageParams{}
	params.SetTo(c.opt.To[0])
	params.SetFrom(c.opt.From)
	params.SetBody(c.body)

	if c.opt.CallbackURL != "" {
		params.SetStatusCallback(c.opt.CallbackURL)
	}

	if c.opt.MessageServiceSID != "" {
		params.SetMessagingServiceSid(c.opt.MessageServiceSID)
	}

	resp, err := c.tw.Api.CreateMessage(params)
	if err != nil {
		switch dataErr := err.(type) {
		case *twilioClient.TwilioRestError:
			msgResp.ResponseStatusCode = dataErr.Status
			msgResp.ErrorCode = int64(dataErr.Code)
			msgResp.Message = dataErr.Message
		default:
			msgResp.ResponseStatusCode = http.StatusInternalServerError
		}

		return msgResp, err
	}

	msgResp.MessageID = converto.StringValue(resp.Sid)
	msgResp.ErrorCode = int64(converto.IntValue(resp.ErrorCode))
	msgResp.Message = converto.StringValue(resp.ErrorMessage)

	return msgResp, nil
}

func (c *client) GetStatus(ctx context.Context, req notify.MessageStatusRequest) (*notify.StatusResponse, error) {
	return &notify.StatusResponse{
		Status:  notify.StatusDelivered,
		Message: "success",
	}, nil
}
