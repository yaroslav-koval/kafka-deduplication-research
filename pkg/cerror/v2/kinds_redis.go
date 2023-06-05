package cerror

import (
	"github.com/go-redis/redis/v8"
)

type KindRedis uint8

const (
	_ KindRedis = iota
	KindRedisOther
	KindRedisNil
	KindRedisClose
	KindRedisPermission
	KindRedisIO

	_kindRedisClose      = "close"
	_kindRedisOther      = "other"
	_kindRedisPermission = "permission"
	_kindRedisIO         = "io"
	_kindRedisNil        = "nil"
)

var (
	_codesRedisPermission = []string{
		"ERR AUTH",
	}

	_codesRedisIO = []string{
		"connection refused",           // dial tcp [::1]:6379: connect: connection refused
		"ERR DB index is out of range", // ERR DB index is out of range
		"READONLY",
	}
)

func (k KindRedis) String() string {
	switch k {
	case KindRedisNil:
		return _kindRedisNil
	case KindRedisPermission:
		return _kindRedisPermission
	case KindRedisIO:
		return _kindRedisIO
	case KindRedisClose:
		return _kindRedisClose
	default:
		return _kindRedisOther
	}
}

func KindFromRedis(err error) KindRedis {
	switch {
	case err == redis.Nil:
		return KindRedisNil
	case err == redis.ErrClosed:
		return KindRedisClose
	case isStrContainsInsensitive(err.Error(), _codesRedisPermission):
		return KindRedisPermission
	case isStrContainsInsensitive(err.Error(), _codesRedisIO):
		return KindRedisIO
	}

	return KindRedisOther
}
