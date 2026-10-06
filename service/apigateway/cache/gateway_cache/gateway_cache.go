package gateway_cache

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type GatewayCache struct {
	Store *cache.CacheStore
}

func NewGatewayCache(store *cache.CacheStore) *GatewayCache {
	return &GatewayCache{Store: store}
}

func Get[T any](ctx context.Context, c *GatewayCache, key string) (*T, bool) {
	if c == nil || c.Store == nil {
		return nil, false
	}
	return cache.GetFromCache[T](ctx, c.Store, key)
}

func Set[T any](ctx context.Context, c *GatewayCache, key string, data *T, ttl time.Duration) {
	if c == nil || c.Store == nil {
		return
	}
	cache.SetToCache[T](ctx, c.Store, key, data, ttl)
}

func Delete(ctx context.Context, c *GatewayCache, key string) {
	if c == nil || c.Store == nil {
		return
	}
	cache.DeleteFromCache(ctx, c.Store, key)
}

func InvalidatePattern(ctx context.Context, c *GatewayCache, pattern string) {
	if c == nil || c.Store == nil {
		return
	}
	c.Store.InvalidateCache(ctx, pattern)
}
