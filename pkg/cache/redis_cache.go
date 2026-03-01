package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache using Redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a Redis-backed cache client.
func NewRedisCache(addr, password string, db int) *RedisCache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	}
	if err != nil {
		return fmt.Errorf("get %s: %w", key, err)
	}
	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return fmt.Errorf("unmarshal %s: %w", key, err)
	}
	return nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", key, err)
	}
	if err := r.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("set %s: %w", key, err)
	}
	return nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}

func (r *RedisCache) InvalidateByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan %s: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("invalidate %s: %w", pattern, err)
			}
		}
		if nextCursor == 0 {
			return nil
		}
		cursor = nextCursor
	}
}
