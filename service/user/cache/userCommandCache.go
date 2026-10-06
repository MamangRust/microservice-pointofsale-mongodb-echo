package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type userCommandCache struct {
	store *cache.CacheStore
}

func NewUserCommandCache(store *cache.CacheStore) UserCommandCache {
	return &userCommandCache{store: store}
}

func (c *userCommandCache) DeleteUserCache(ctx context.Context, id int) {
	key := fmt.Sprintf(userByIdCacheKey, id)
	cache.DeleteFromCache(ctx, c.store, key)
}

func (c *userCommandCache) DeleteUserAllCache(ctx context.Context) {
	c.store.InvalidateCache(ctx, "user:*")
}
