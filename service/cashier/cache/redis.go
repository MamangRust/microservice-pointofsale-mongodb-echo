package mencache

import (
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type Mencache interface {
	CashierQueryCache
	CashierCommandCache
}

type mencache struct {
	CashierQueryCache
	CashierCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		CashierQueryCache:   NewCashierQueryCache(cacheStore),
		CashierCommandCache: NewCashierCommandCache(cacheStore),
	}
}
