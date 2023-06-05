package factory_test

import (
	"kafka-polygon/pkg/log/factory"
	"kafka-polygon/pkg/log/zero"
	"testing"

	"github.com/tj/assert"
)

func TestGetLogger(t *testing.T) {
	assert.IsType(t, &zero.Logger{}, factory.GetLogger())
}
