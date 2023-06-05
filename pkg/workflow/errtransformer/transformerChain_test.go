package errtransformer

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

var _cb = context.Background()

func TestChainTransform(t *testing.T) {
	t.Parallel()

	c := NewChain([]ErrorTransformer{
		NewErrorByHTTPCodeTransformer([]int{400}),
		NewErrorByMessageTransformer([]string{"custom_message"}),
	})

	// http transformer
	err := cerror.NewF(_cb, cerror.KindBadParams, "")
	newErr := c.Transform(_cb, err)
	assert.Equal(t, entity.NewProcessingError(err).SetRetry(true), newErr)

	// message transformer
	err = cerror.NewF(_cb, cerror.KindNotExist, "custom_message")
	newErr = c.Transform(_cb, err)
	assert.Equal(t, entity.NewProcessingError(err).SetRetry(true), newErr)

	// no one transformer
	err = cerror.NewF(_cb, cerror.KindNotExist, "")
	newErr = c.Transform(_cb, err)
	assert.Equal(t, err, newErr)
}
