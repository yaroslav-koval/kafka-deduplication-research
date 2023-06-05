package notify

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"strings"
)

func CheckChannelAndProvider(ctx context.Context, channel, provider string) error {
	switch TypeNotify(channel) {
	case TypeEmail:
		return checkEmailByProvider(ctx, provider)
	case TypeSMS:
		return checkSMSByProvider(ctx, provider)
	}

	return cerror.NewF(ctx, cerror.KindBadValidation, "not supported channel %s", channel).
		LogError()
}

func checkEmailByProvider(ctx context.Context, name string) error {
	currProvName := strings.ToLower(name)

	supportedProviders := []string{
		ProviderPostmark,
		ProviderMailgun,
	}

	for _, provName := range supportedProviders {
		if strings.EqualFold(currProvName, provName) {
			return nil
		}
	}

	return cerror.NewF(ctx,
		cerror.KindBadValidation, "unknown %s notifier %s", TypeEmail, name).LogError()
}

func checkSMSByProvider(ctx context.Context, name string) error {
	currProvName := strings.ToLower(name)

	supportedProviders := []string{
		ProviderTiniyo,
		ProviderTwilio,
	}

	for _, provName := range supportedProviders {
		if strings.EqualFold(currProvName, provName) {
			return nil
		}
	}

	return cerror.NewF(ctx,
		cerror.KindBadValidation, "unknown %s notifier %s", TypeSMS, name).LogError()
}
