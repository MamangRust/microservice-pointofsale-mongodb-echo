package handler

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

// StatsCache wraps the shared CacheStore with a 5-minute TTL, per the F4
// spec (§7.4 "Redis cache TTL 5 menit di layer reader"). Keys are namespaced
// with "stats:reader:" so they can never collide with gateway ("apigw:") or
// domain-service cache keys.
type StatsCache struct {
	store *cache.CacheStore
	ttl   time.Duration
}

func NewStatsCache(store *cache.CacheStore) *StatsCache {
	return &StatsCache{store: store, ttl: 5 * time.Minute}
}

// CacheGet returns a cached value; nil/false when the cache is disabled or the
// key is absent (callers fall back to ClickHouse).
func CacheGet[T any](ctx context.Context, c *StatsCache, key string) (*T, bool) {
	if c == nil || c.store == nil {
		return nil, false
	}
	return cache.GetFromCache[T](ctx, c.store, key)
}

func CacheSet[T any](ctx context.Context, c *StatsCache, key string, data *T) {
	if c == nil || c.store == nil || data == nil {
		return
	}
	cache.SetToCache[T](ctx, c.store, key, data, c.ttl)
}
