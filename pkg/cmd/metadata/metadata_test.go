package metadata_test

import (
	"kafka-polygon/pkg/cmd/metadata"
	"testing"

	"github.com/tj/assert"
)

func TestMeta(t *testing.T) {
	expModule := "notification"
	expVer := "4d7f728-dirty"
	expBuildDate := "20230124.125009.N.+0200"
	m := metadata.Meta{
		Version:   expVer,
		Module:    expModule,
		BuildDate: expBuildDate,
	}

	assert.Equal(t, expVer, m.Version)
	assert.Equal(t, expModule, m.Module)
	assert.Equal(t, expBuildDate, m.BuildDate)
}
