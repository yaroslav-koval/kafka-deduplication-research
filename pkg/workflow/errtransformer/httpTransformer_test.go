package errtransformer

import (
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestHTTPCodeTransform(t *testing.T) {
	t.Parallel()

	retryCodes := []int{400, 500}

	transformer := NewErrorByHTTPCodeTransformer(retryCodes)

	type testData struct {
		inputErr error
		expErr   error
	}

	td := []testData{{nil, nil}}

	for _, c := range retryCodes {
		cerrRetry := cerror.NewF(_cb, cerror.KindFromHTTPCode(c), "err")
		td = append(td, testData{cerrRetry, entity.NewProcessingError(cerrRetry).SetRetry(true)})

		cerrNoRetry := cerror.NewF(_cb, cerror.KindFromHTTPCode(c+1), "err")
		td = append(td, testData{cerrNoRetry, cerrNoRetry})
	}

	for _, td := range td {
		assert.Equal(t, td.expErr, transformer.Transform(_cb, td.inputErr))
	}
}
