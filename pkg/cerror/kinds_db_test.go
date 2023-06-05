package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/tj/assert"
)

type expectedDBKind struct {
	kind         cerror.Kind
	kindString   string
	kindHTTPCode int
}

var expectedKindDBTable = []expectedDBKind{
	{cerror.KindDBOther, "db_other_error", http.StatusInternalServerError},
	{cerror.KindDBNoRows, "db_no_rows_error", http.StatusNotFound},
	{cerror.KindDBMultiRows, "db_multi_rows_error", http.StatusInternalServerError},
	{cerror.KindDBPermission, "db_permission_denied", http.StatusInternalServerError},
	{cerror.KindDBIO, "db_io_error", http.StatusInternalServerError},
}

var errDBList = []struct {
	kind cerror.Kind
	err  error
}{
	{
		kind: cerror.KindDBNoRows,
		err:  pg.ErrNoRows,
	},
	{
		kind: cerror.KindDBMultiRows,
		err:  pg.ErrMultiRows,
	},
	{
		kind: cerror.KindDBPermission,
		err:  errors.New("0LP01: invalid_grant_operation"),
	},
	{
		kind: cerror.KindDBIO,
		err:  errors.New("25P03: idle_in_transaction_session_timeout"),
	},
}

func TestKindDBString(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindDBTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}

func TestKindDBHTTPCode(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindDBTable {
		assert.Equal(t, v.kindHTTPCode, v.kind.HTTPCode())
	}
}

func TestDBToKind(t *testing.T) {
	t.Parallel()

	for _, e := range errDBList {
		errKind := cerror.DBToKind(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
