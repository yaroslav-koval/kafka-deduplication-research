package notify_test

import (
	"context"
	"kafka-polygon/pkg/notify"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/tj/assert"
)

type MockedByEmail struct {
	mock.Mock
}

func (mbe *MockedByEmail) UID() string {
	args := mbe.Called()

	return args.String(0)
}

func (mbe *MockedByEmail) WithBaseURL(baseURL string) {
	_ = mbe.Called(baseURL)
}

func (mbe *MockedByEmail) GetStatus(ctx context.Context, req notify.MessageStatusRequest) (*notify.StatusResponse, error) {
	args := mbe.Called(ctx, req)

	return args.Get(0).(*notify.StatusResponse), args.Error(1)
}

func (mbe *MockedByEmail) From(from string) notify.ByEmail {
	args := mbe.Called(from)

	return args.Get(0).(notify.ByEmail)
}

func (mbe *MockedByEmail) WithSubject(subject string) notify.ByEmail {
	args := mbe.Called(subject)

	return args.Get(0).(notify.ByEmail)
}

func (mbe *MockedByEmail) UseHTML() notify.ByEmail {
	args := mbe.Called()

	return args.Get(0).(notify.ByEmail)
}

func (mbe *MockedByEmail) WithBody(body string) notify.ByEmail {
	args := mbe.Called(body)

	return args.Get(0).(notify.ByEmail)
}

func (mbe *MockedByEmail) WithTag(tag string) notify.ByEmail {
	args := mbe.Called(tag)

	return args.Get(0).(notify.ByEmail)
}

func (mbe *MockedByEmail) WithTrackOpens() notify.ByEmail {
	args := mbe.Called()

	return args.Get(0).(notify.ByEmail)
}

func (mbe *MockedByEmail) To(to string, cc ...string) notify.ByEmail {
	args := mbe.Called(to, cc)

	return args.Get(0).(notify.ByEmail)
}

func (mbe *MockedByEmail) Send(ctx context.Context) (*notify.MessageResponse, error) {
	args := mbe.Called(ctx)

	return args.Get(0).(*notify.MessageResponse), args.Error(1)
}

type MockedBySMS struct {
	mock.Mock
}

func (mbs *MockedBySMS) UID() string {
	args := mbs.Called()

	return args.String(0)
}

func (mbs *MockedBySMS) GetStatus(ctx context.Context, req notify.MessageStatusRequest) (*notify.StatusResponse, error) {
	args := mbs.Called(ctx, req)

	return args.Get(0).(*notify.StatusResponse), args.Error(1)
}

func (mbs *MockedBySMS) From(from string) notify.BySMS {
	args := mbs.Called(from)

	return args.Get(0).(notify.BySMS)
}

func (mbs *MockedBySMS) WithBody(body string) notify.BySMS {
	args := mbs.Called(body)

	return args.Get(0).(notify.BySMS)
}

func (mbs *MockedBySMS) To(to string, cc ...string) notify.BySMS {
	args := mbs.Called(to, cc)

	return args.Get(0).(notify.BySMS)
}

func (mbs *MockedBySMS) Send(ctx context.Context) (*notify.MessageResponse, error) {
	args := mbs.Called(ctx)

	return args.Get(0).(*notify.MessageResponse), args.Error(1)
}

func TestProvider(t *testing.T) {
	t.Parallel()

	expEmailNameProv := "test-email-prov"
	expSMSNameProv := "test-sms-prov"

	mockEmailProv := &MockedByEmail{}
	mockEmailProv.On("UID").Return(expEmailNameProv)

	mockSMSProv := &MockedBySMS{}
	mockSMSProv.On("UID").Return(expSMSNameProv)

	prov := notify.New()
	prov.AddEmailProvider(mockEmailProv)
	prov.AddSMSProvider(mockSMSProv)

	providers := map[notify.TypeNotify]string{
		notify.TypeEmail: expEmailNameProv,
		notify.TypeSMS:   expSMSNameProv,
	}

	for channel, expProv := range providers {
		checkProv := prov.HasProvider(channel, expProv)
		assert.True(t, checkProv)

		actualProv := prov.GetProvider(channel, expProv)

		switch data := actualProv.(type) {
		case notify.ByEmail:
			assert.IsType(t, mockEmailProv, data)
			assert.Equal(t, expProv, data.UID())
		case notify.BySMS:
			assert.IsType(t, mockSMSProv, data)
			assert.Equal(t, expProv, data.UID())
		}

		provList := prov.GetProviders(channel)
		assert.True(t, len(provList) > 0)
	}
}

func TestMultiProvider(t *testing.T) {
	t.Parallel()

	expProviders := map[notify.TypeNotify]map[string]interface{}{
		notify.TypeEmail: {
			"test-email-prov-1": &MockedByEmail{},
			"test-email-prov-2": &MockedByEmail{},
		},
		notify.TypeSMS: {
			"test-sms-prov-1": &MockedBySMS{},
			"test-sms-prov-2": &MockedBySMS{},
		},
	}

	prov := notify.New()

	for nType, mockProviders := range expProviders {
		for uid, mockProvider := range mockProviders {
			switch nType {
			case notify.TypeSMS:
				mockSmsProv := mockProvider.(*MockedBySMS)
				mockSmsProv.On("UID").Return(uid)
				prov.AddSMSProvider(mockSmsProv)
			case notify.TypeEmail:
				mockEmailProv := mockProvider.(*MockedByEmail)
				mockEmailProv.On("UID").Return(uid)
				prov.AddEmailProvider(mockEmailProv)
			}
		}
	}

	for nType, mockProviders := range expProviders {
		actualProviders := prov.GetProviders(nType)
		assert.Equal(t, len(mockProviders), len(actualProviders))

		for pName := range mockProviders {
			assert.True(t, prov.HasProvider(nType, pName))
		}
	}
}
