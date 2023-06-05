package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"testing"

	"github.com/tj/assert"
)

type expectedMinioKind struct {
	kind         cerror.Kind
	kindString   string
	kindHTTPCode int
}

var expectedKindMinioTable = []expectedMinioKind{
	{cerror.KindMinioOther, "minio_other_error", http.StatusInternalServerError},
	{cerror.KindMinioNotExist, "minio_not_exist", http.StatusNotFound},
	{cerror.KindMinioConflict, "minio_conflict", http.StatusInternalServerError},
	{cerror.KindMinioPreconditionFailed, "minio_precondition_failed", http.StatusInternalServerError},
	{cerror.KindMinioPermission, "minio_permission_denied", http.StatusInternalServerError},
	{cerror.KindMinioIO, "minio_io_error", http.StatusInternalServerError},
}

var errMinioList = []struct {
	kind cerror.Kind
	err  error
}{
	{
		kind: cerror.KindMinioOther,
		err:  errors.New("other error"),
	},

	{
		kind: cerror.KindMinioNotExist,
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

func TestKindMinioHTTPCode(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindMinioTable {
		assert.Equal(t, v.kindHTTPCode, v.kind.HTTPCode())
	}
}

func TestMinioToKind(t *testing.T) {
	t.Parallel()

	for _, e := range errMinioList {
		errKind := cerror.MinioToKind(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
