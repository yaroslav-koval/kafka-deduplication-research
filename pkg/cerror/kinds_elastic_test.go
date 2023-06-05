package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"testing"

	"github.com/tj/assert"
)

type expectedElasticKind struct {
	kind         cerror.Kind
	kindString   string
	kindHTTPCode int
}

var expectedKindElasticTable = []expectedElasticKind{
	{cerror.KindElasticOther, "elastic_other_error", http.StatusInternalServerError},
	{cerror.KindElasticPermission, "elastic_permission_denied", http.StatusInternalServerError},
	{cerror.KindElasticIO, "elastic_io_error", http.StatusInternalServerError},
}

var errElasticList = []struct {
	kind cerror.Kind
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

func TestKindElasticHTTPCode(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindElasticTable {
		assert.Equal(t, v.kindHTTPCode, v.kind.HTTPCode())
	}
}

func TestElasticToKind(t *testing.T) {
	t.Parallel()

	for _, e := range errElasticList {
		errKind := cerror.ElasticToKind(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
