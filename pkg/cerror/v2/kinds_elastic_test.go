package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror/v2"
	"testing"

	"github.com/tj/assert"
)

type expectedElasticKind struct {
	kind       cerror.KindElastic
	kindString string
}

var expectedKindElasticTable = []expectedElasticKind{
	{cerror.KindElasticOther, "other"},
	{cerror.KindElasticPermission, "permission"},
	{cerror.KindElasticIO, "io"},
}

var errElasticList = []struct {
	kind cerror.KindElastic
	err  error
}{
	{
		kind: cerror.KindElasticOther,
		err:  errors.New("other error"),
	},
	{
		kind: cerror.KindElasticPermission,
		err:  errors.New("client Access Denied"),
	},
	{
		kind: cerror.KindElasticIO,
		err:  errors.New("failed to compress request body"),
	},
}

func TestKindElasticString(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindElasticTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}

func TestKindFromElastic(t *testing.T) {
	t.Parallel()

	for _, e := range errElasticList {
		errKind := cerror.KindFromElastic(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
