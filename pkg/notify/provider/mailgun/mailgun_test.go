package mailgun_test

import (
	"context"
	"errors"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/notify"
	pkgMailgun "kafka-polygon/pkg/notify/provider/mailgun"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
)

type mockClient struct {
	DoSend       func(ctx context.Context, message *mailgun.Message) (mes, id string, err error)
	DoSetAPIBase func(address string)
}

func (m *mockClient) NewMessage(from, subject, text string, to ...string) *mailgun.Message {
	cl := mailgun.NewMailgun("", "")
	return cl.NewMessage(from, subject, text, to...)
}

func (m *mockClient) Send(ctx context.Context, message *mailgun.Message) (mes, id string, err error) {
	return m.DoSend(ctx, message)
}

func (m *mockClient) SetAPIBase(address string) {
	m.DoSetAPIBase(address)
}

type clientTestSuite struct {
	suite.Suite
	mailgunProvClient *pkgMailgun.Client
	mailgunOptions    *pkgMailgun.Options
	mailgunMockClient *mockClient
}

func TestClientTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(clientTestSuite))
}

func (s *clientTestSuite) SetupSuite() {
	s.mailgunMockClient = new(mockClient)
	s.mailgunOptions = &pkgMailgun.Options{
		Domain: "domain",
		APIKey: "apiKey",
	}
	s.mailgunProvClient = pkgMailgun.New(*s.mailgunOptions).WithMailgunClient(s.mailgunMockClient)
}

func (s *clientTestSuite) TearDownTest() {
}

func (s *clientTestSuite) TearDownSuite() {
}

func (s *clientTestSuite) TestSend() {
	ctx := context.Background()
	_, err := s.mailgunProvClient.Send(ctx)
	s.Error(err)
	s.Contains(err.Error(), "empty To")

	s.mailgunProvClient.To("ashyshlakov@edenlab.com.ua")

	_, err = s.mailgunProvClient.Send(ctx)
	s.Error(err)
	s.Contains(err.Error(), "empty From")

	s.mailgunProvClient.From("shyshlakov@icloud.com")

	s.mailgunMockClient.DoSend = func(ctx context.Context, message *mailgun.Message) (mes, id string, err error) {
		return "", "", errors.New("failed to send message")
	}

	msgResp, err := s.mailgunProvClient.Send(ctx)
	s.Error(err)
	s.Contains(err.Error(), "send email via mailgun: failed")
	s.Contains(msgResp.Message, "send message")
	s.EqualValues(500, msgResp.ErrorCode)

	msgID := uuid.NewV4().String()
	s.mailgunMockClient.DoSend = func(ctx context.Context, message *mailgun.Message) (mes, id string, err error) {
		return "Queued. Thank you.", msgID, nil
	}

	msgResp, err = s.mailgunProvClient.Send(ctx)
	s.NoError(err)
	s.Contains(msgResp.Message, "Queued.")
	s.Equal(msgID, msgResp.MessageID)
}

func (s *clientTestSuite) TestUUID() {
	s.Equal(notify.ProviderMailgun, s.mailgunProvClient.UID())
}

func (s *clientTestSuite) TestSetBody() {
	ctx := context.Background()

	msgMailgun := s.mailgunMockClient.NewMessage("from@gmail.com", "subject", "text", "to@gmail.com")

	s.mailgunProvClient.UseHTML()
	err := s.mailgunProvClient.SetBody(ctx, msgMailgun)
	s.NoError(err)
}
