package cerror

import (
	"net/http"

	"github.com/go-redis/redis/v8"
)

type RedisKind uint8

const (
	_ RedisKind = iota
	KindRedisOther
	KindRedisNil
	KindRedisClose
	KindRedisPermission // KindPermission denied
	KindRedisIO         // External I/O error such as network failure

	redisClose            = "redis_close_error"
	redisUnknown          = "redis_unknown_error"
	redisOther            = "redis_other_error"
	redisPermissionDenied = "redis_permission_denied"
	redisIO               = "redis_io_error"
	redisNil              = "redis_nil_error"
)

var (
	_redisPermissionCodes = []string{
		"ERR AUTH",
	}

	_redisIOCodes = []string{
		"connection refused",           // dial tcp [::1]:6379: connect: connection refused
		"ERR DB index is out of range", // ERR DB index is out of range
		"READONLY",
	}
)

func (k RedisKind) String() string {
	switch k {
	case KindRedisNil:
		return redisNil
	case KindRedisPermission:
		return redisPermissionDenied
	case KindRedisIO:
		return redisIO
	case KindRedisClose:
		return redisClose
	case KindRedisOther:
		return redisOther
	}

	return redisUnknown
}

func (k RedisKind) HTTPCode() int {
	return http.StatusInternalServerError
}

func (k RedisKind) Group() KindGroup {
	return GroupRedis
}

func RedisToKind(err error) RedisKind {
	switch {
	case err == redis.Nil:
		return KindRedisNil
	case err == redis.ErrClosed:
		return KindRedisClose
	case checkCode(err.Error(), _redisPermissionCodes):
		return KindRedisPermission
	case checkCode(err.Error(), _redisIOCodes):
		return KindRedisIO
	default:
		return KindRedisOther
	}
}
