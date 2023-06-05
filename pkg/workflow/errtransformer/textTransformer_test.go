package errtransformer

import (
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestTextTransform(t *testing.T) {
	t.Parallel()

	customTexts := []string{"1", "2"}

	transformer := NewErrorByMessageTransformer(customTexts)

	type testData struct {
		inputErr error
		expErr   error
	}

	td := []testData{{nil, nil}}

	for _, t := range customTexts {
		cerrRetry := cerror.NewF(_cb, cerror.KindBadParams, t)
		td = append(td, testData{cerrRetry, entity.NewProcessingError(cerrRetry).SetRetry(true)})

		cerrNoRetry := cerror.NewF(_cb, cerror.KindBadParams, "3")
		td = append(td, testData{cerrNoRetry, cerrNoRetry})
	}

	for _, td := range td {
		assert.Equal(t, td.expErr, transformer.Transform(_cb, td.inputErr))
	}
}
