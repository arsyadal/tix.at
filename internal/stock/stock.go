package stock

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var reserveScript = redis.NewScript(`
local stock = tonumber(redis.call("GET", KEYS[1]) or "0")
if stock <= 0 then
  return 0
end
redis.call("DECR", KEYS[1])
return 1
`)

type Store struct{ rdb *redis.Client }

func New(addr string) *Store {
	return &Store{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func (s *Store) Close() error { return s.rdb.Close() }

func (s *Store) Ping(ctx context.Context) error { return s.rdb.Ping(ctx).Err() }

func (s *Store) Set(ctx context.Context, eventID string, stock int) error {
	return s.rdb.Set(ctx, key(eventID), stock, 0).Err()
}

func (s *Store) Get(ctx context.Context, eventID string) (int, error) {
	v, err := s.rdb.Get(ctx, key(eventID)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return v, err
}

func (s *Store) Reserve(ctx context.Context, eventID string) (bool, error) {
	v, err := reserveScript.Run(ctx, s.rdb, []string{key(eventID)}).Int64()
	return v == 1, err
}

func (s *Store) Release(ctx context.Context, eventID string) error {
	return s.rdb.Incr(ctx, key(eventID)).Err()
}

func key(eventID string) string { return fmt.Sprintf("event:%s:stock", eventID) }
