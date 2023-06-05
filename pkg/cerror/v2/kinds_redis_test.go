package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror/v2"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/tj/assert"
)

type expectedRedisKind struct {
	kind       cerror.KindRedis
	kindString string
}

var expectedKindRedisTable = []expectedRedisKind{
	{cerror.KindRedisOther, "other"},
	{cerror.KindRedisNil, "nil"},
	{cerror.KindRedisClose, "close"},
	{cerror.KindRedisPermission, "permission"},
	{cerror.KindRedisIO, "io"},
}

var errRedisList = []struct {
	kind cerror.KindRedis
	err  error
}{
	{
		kind: cerror.KindRedisOther,
		err:  errors.New("other error"),
	},

	{
		kind: cerror.KindRedisNil,
		err:  redis.Nil,
	},
	{
		kind: cerror.KindRedisClose,
		err:  redis.ErrClosed,
	},
	{
		kind: cerror.KindRedisPermission,
		err:  errors.New("ERR AUTH: <access_denied>"),
	},
	{
		kind: cerror.KindRedisIO,
		err:  errors.New("connection refused"),
	},
}

func TestKindRedisString(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindRedisTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}

func TestKindFromRedis(t *testing.T) {
	t.Parallel()

	for _, e := range errRedisList {
		errKind := cerror.KindFromRedis(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
