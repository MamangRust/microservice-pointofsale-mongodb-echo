package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type orderItemCommandCache struct {
	store *cache.CacheStore
}

func NewOrderItemCommandCache(store *cache.CacheStore) OrderItemCommandCache {
	return &orderItemCommandCache{store: store}
}

func (s *orderItemCommandCache) DeleteOrderItemCache(ctx context.Context, orderID int) {
	cache.DeleteFromCache(ctx, s.store, fmt.Sprintf(orderItemByIdCacheKey, orderID))
}

func (s *orderItemCommandCache) DeleteOrderItemAllCache(ctx context.Context) {
	s.store.InvalidateCache(ctx, "order_item:*")
}
