package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned when a cache lookup finds no value.
var ErrCacheMiss = errors.New("cache miss")

// Cache is a generic key-value store with TTL support.
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	InvalidateByPattern(ctx context.Context, pattern string) error
}
