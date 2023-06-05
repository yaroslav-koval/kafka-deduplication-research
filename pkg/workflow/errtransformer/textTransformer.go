package errtransformer

import (
	"context"
	"kafka-polygon/pkg/workflow/entity"
	"strings"

	"github.com/samber/lo"
)

type ProcessingErrorByMessageTransformer struct {
	customTexts []string
}

func NewErrorByMessageTransformer(customTexts []string) *ProcessingErrorByMessageTransformer {
	return &ProcessingErrorByMessageTransformer{
		customTexts: lo.Map(customTexts, func(t string, _ int) string {
			return strings.ToLower(t)
		}),
	}
}

func (c *ProcessingErrorByMessageTransformer) Transform(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	msg := strings.ToLower(err.Error())

	if lo.Contains(c.customTexts, msg) {
		return entity.NewProcessingError(err).SetRetry(true)
	}

	return err
}
