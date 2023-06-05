package redis

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cerror"
	"time"

	"github.com/go-redis/redis/v8"
)

type Settings struct {
	KeyPrefix string
	TTL       time.Duration
}

type Store struct {
	rc       *redis.Client
	settings Settings
}

var _ store.Store = (*Store)(nil)

func NewStore(rc *redis.Client, settings Settings) *Store {
	return &Store{
		rc:       rc,
		settings: settings,
	}
}

func (s *Store) GetEventInfoByID(ctx context.Context, id string) (store.EventProcessData, error) {
	key := s.getKey(id)

	val, err := s.rc.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return store.EventProcessData{}, cerror.New(ctx, cerror.KindNotExist, err) //nolint:cerrl
		}

		return store.EventProcessData{}, cerror.New(ctx, cerror.RedisToKind(err), err).LogError()
	}

	return store.EventProcessData{Status: val}, nil
}

func (s *Store) PutEventInfo(ctx context.Context, id string, data store.EventProcessData) error {
	key := s.getKey(id)

	err := s.rc.Set(ctx, key, []byte(data.Status), s.settings.TTL).Err()
	if err != nil {
		return cerror.New(ctx, cerror.RedisToKind(err), err).LogError()
	}

	return nil
}

func (s *Store) DeleteEventInfoByID(ctx context.Context, id string) error {
	key := s.getKey(id)

	err := s.rc.Del(ctx, key).Err()
	if err != nil {
		return cerror.New(ctx, cerror.RedisToKind(err), err).LogError()
	}

	return nil
}

func (s *Store) getKey(id string) string {
	key := id
	if s.settings.KeyPrefix != "" {
		key = fmt.Sprintf("%s_%s", s.settings.KeyPrefix, key)
	}

	return key
}
