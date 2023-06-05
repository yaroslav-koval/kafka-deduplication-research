package errtransformer

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/workflow/entity"

	"github.com/samber/lo"
)

type ProcessingErrorByHTTPCodeTransformer struct {
	retryCodes []int
}

func NewErrorByHTTPCodeTransformer(retryCodes []int) *ProcessingErrorByHTTPCodeTransformer {
	return &ProcessingErrorByHTTPCodeTransformer{
		retryCodes: retryCodes,
	}
}

func (c *ProcessingErrorByHTTPCodeTransformer) Transform(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	if lo.Contains(c.retryCodes, cerror.ErrKind(err).HTTPCode()) {
		return entity.NewProcessingError(err).SetRetry(true)
	}

	return err
}
