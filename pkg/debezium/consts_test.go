package debezium_test

import (
	"kafka-polygon/pkg/debezium"
	"testing"

	"github.com/tj/assert"
)

func TestDebeziumConst(t *testing.T) {
	t.Parallel()

	sConst := map[string]string{
		debezium.OpCreate: "c",
		debezium.OpUpdate: "u",
		debezium.OpDelete: "d",
		debezium.OpRead:   "r",
	}

	for actual, expected := range sConst {
		assert.Equal(t, expected, actual)
	}
}
