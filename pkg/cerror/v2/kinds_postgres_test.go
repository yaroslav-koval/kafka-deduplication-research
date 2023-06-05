package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror/v2"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/tj/assert"
)

type expectedDBKind struct {
	kind       cerror.KindPostgres
	kindString string
}

var expectedKindPostgresTable = []expectedDBKind{
	{cerror.KindPostgresOther, "other"},
	{cerror.KindPostgresNoRows, "no_rows"},
	{cerror.KindPostgresMultiRows, "multi_rows"},
	{cerror.KindPostgresPermission, "permission"},
	{cerror.KindPostgresIO, "io"},
}

var errDBList = []struct {
	kind cerror.KindPostgres
	err  error
}{
	{
		kind: cerror.KindPostgresNoRows,
		err:  pg.ErrNoRows,
	},
	{
		kind: cerror.KindPostgresMultiRows,
		err:  pg.ErrMultiRows,
	},
	{
		kind: cerror.KindPostgresPermission,
		err:  errors.New("0LP01: invalid_grant_operation"),
	},
	{
		kind: cerror.KindPostgresIO,
		err:  errors.New("25P03: idle_in_transaction_session_timeout"),
	},
}

func TestKindPostgresString(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindPostgresTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}

func TestKindFromPostgres(t *testing.T) {
	t.Parallel()

	for _, e := range errDBList {
		errKind := cerror.KindFromPostgres(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
