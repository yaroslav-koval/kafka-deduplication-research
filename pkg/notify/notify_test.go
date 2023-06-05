package notify_test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/notify"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	bgCtx = context.Background()
)

func TestCheckChannelAndProvider(t *testing.T) {
	t.Parallel()

	notifyList := map[notify.TypeNotify][]string{
		notify.TypeEmail: {
			notify.ProviderPostmark,
		},
		notify.TypeSMS: {
			notify.ProviderTwilio,
		},
	}

	for channel, providers := range notifyList {
		for _, provider := range providers {
			err := notify.CheckChannelAndProvider(bgCtx, channel.ToString(), provider)
			require.NoError(t, err)
		}
	}
}

func TestCheckChannelAndProviderError(t *testing.T) {
	t.Parallel()

	notifyList := map[notify.TypeNotify][]string{
		notify.TypeEmail: {
			"not-supported-email-provider",
		},
		notify.TypeSMS: {
			"not-supported-sms-provider",
		},
	}

	for channel, providers := range notifyList {
		for _, provider := range providers {
			err := notify.CheckChannelAndProvider(bgCtx, channel.ToString(), provider)
			require.Error(t, err)
			assert.Equal(t,
				fmt.Sprintf("unknown %s notifier %s", channel.ToString(), provider),
				err.Error())
		}
	}
}
