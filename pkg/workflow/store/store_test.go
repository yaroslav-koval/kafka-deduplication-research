package store_test

import (
	"kafka-polygon/pkg/workflow/store"
	"testing"

	"github.com/tj/assert"
)

func TestConst(t *testing.T) {
	t.Parallel()

	sConst := map[string]string{
		store.StoreTypeMongo:    "mongo",
		store.StoreTypePostgres: "postgres",
	}

	for actual, expected := range sConst {
		assert.Equal(t, actual, expected)
	}
}
