package entity_test

import (
	"fmt"
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestNewProcessingError(t *testing.T) {
	err := fmt.Errorf("err")

	actErr := entity.NewProcessingError(err)

	assert.Equal(t, err, actErr.OriginalError())
	assert.False(t, actErr.Retry())
}

func TestProcessingErrorError(t *testing.T) {
	err := fmt.Errorf("err")

	actErr := entity.NewProcessingError(err)

	assert.Equal(t, err.Error(), actErr.Error())
}

func TestProcessingErrorRetry(t *testing.T) {
	err := fmt.Errorf("err")

	actErr := entity.NewProcessingError(err)

	assert.False(t, actErr.Retry())
	_ = actErr.SetRetry(true)
	assert.True(t, actErr.Retry())
	_ = actErr.SetRetry(false)
	assert.False(t, actErr.Retry())
}
