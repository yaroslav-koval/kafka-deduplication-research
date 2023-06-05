package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror/v2"
	"testing"

	"github.com/tj/assert"
)

type expectedMinioKind struct {
	kind       cerror.KindMinio
	kindString string
}

var expectedKindMinioTable = []expectedMinioKind{
	{cerror.KindMinioOther, "other"},
	{cerror.KindMinioNotExists, "not_exists"},
	{cerror.KindMinioConflict, "conflict"},
	{cerror.KindMinioPreconditionFailed, "precondition_failed"},
	{cerror.KindMinioPermission, "permission"},
	{cerror.KindMinioIO, "io"},
}

var errMinioList = []struct {
	kind cerror.KindMinio
	err  error
}{
	{
		kind: cerror.KindMinioOther,
		err:  errors.New("other error"),
	},

	{
		kind: cerror.KindMinioNotExists,
		err:  errors.New("the specified bucket does not exist"),
	},
	{
		kind: cerror.KindMinioPreconditionFailed,
		err:  errors.New("at least one of the pre-conditions you specified did not hold"),
	},

	{
		kind: cerror.KindMinioConflict,
		err:  errors.New("err: bucket not empty"),
	},
	{
		kind: cerror.KindMinioPermission,
		err:  errors.New("err: Access Denied"),
	},
	{
		kind: cerror.KindMinioIO,
		err:  errors.New("err: connection refused"),
	},
}

func TestKindMinioString(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindMinioTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}

func TestKindFromMinio(t *testing.T) {
	t.Parallel()

	for _, e := range errMinioList {
		errKind := cerror.KindFromMinio(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
