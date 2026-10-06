package mencache

import (
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type Mencache interface {
	CategoryQueryCache
	CategoryCommandCache
}

type mencache struct {
	CategoryQueryCache
	CategoryCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		CategoryQueryCache:   NewCategoryQueryCache(cacheStore),
		CategoryCommandCache: NewCategoryCommandCache(cacheStore),
	}
}
