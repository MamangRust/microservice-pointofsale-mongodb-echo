package mencache

import (
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type Mencache interface {
	TransactionQueryCache
	TransactionCommandCache
}

type mencache struct {
	TransactionQueryCache
	TransactionCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		TransactionQueryCache:   NewTransactionQueryCache(cacheStore),
		TransactionCommandCache: NewTransactionCommandCache(cacheStore),
	}
}
