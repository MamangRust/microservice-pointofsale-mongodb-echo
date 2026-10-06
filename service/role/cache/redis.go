package mencache

import (
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type Mencache interface {
	RoleQueryCache
	RoleCommandCache
}

type mencache struct {
	RoleQueryCache
	RoleCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		RoleQueryCache:   NewRoleQueryCache(cacheStore),
		RoleCommandCache: NewRoleCommandCache(cacheStore),
	}
}
