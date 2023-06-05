package cerror_test

import (
	"errors"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/tj/assert"
)

type expectedRedisKind struct {
	kind         cerror.Kind
	kindString   string
	kindHTTPCode int
}

var expectedKindRedisTable = []expectedRedisKind{
	{cerror.KindRedisOther, "redis_other_error", http.StatusInternalServerError},
	{cerror.KindRedisNil, "redis_nil_error", http.StatusInternalServerError},
	{cerror.KindRedisClose, "redis_close_error", http.StatusInternalServerError},
	{cerror.KindRedisPermission, "redis_permission_denied", http.StatusInternalServerError},
	{cerror.KindRedisIO, "redis_io_error", http.StatusInternalServerError},
}

var errRedisList = []struct {
	kind cerror.Kind
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

func TestKindRedisHTTPCode(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindRedisTable {
		assert.Equal(t, v.kindHTTPCode, v.kind.HTTPCode())
	}
}

func TestRedisToKind(t *testing.T) {
	t.Parallel()

	for _, e := range errRedisList {
		errKind := cerror.RedisToKind(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
