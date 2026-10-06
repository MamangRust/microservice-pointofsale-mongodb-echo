package mencache

import (
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type Mencache interface {
	OrderQueryCache
	OrderCommandCache
}

type mencache struct {
	OrderQueryCache
	OrderCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		OrderQueryCache:   NewOrderQueryCache(cacheStore),
		OrderCommandCache: NewOrderCommandCache(cacheStore),
	}
}
