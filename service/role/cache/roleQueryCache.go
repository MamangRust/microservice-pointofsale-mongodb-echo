package mencache

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-role/repository"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

const (
	roleAllCacheKey      = "role:all:page:%d:pageSize:%d:search:%s"
	roleByIdCacheKey     = "role:id:%d"
	roleByUserIdCacheKey = "role:user:%d"
	roleActiveCacheKey   = "role:active:page:%d:pageSize:%d:search:%s"
	roleTrashedCacheKey  = "role:trashed:page:%d:pageSize:%d:search:%s"
	ttlDefault           = 5 * time.Minute
)

type roleCachedResponse struct {
	Data         []*repository.RoleResult `json:"data"`
	TotalRecords *int                     `json:"total_records"`
}

type roleQueryCache struct {
	store *cache.CacheStore
}

func NewRoleQueryCache(store *cache.CacheStore) RoleQueryCache {
	return &roleQueryCache{store: store}
}

func (m *roleQueryCache) SetCachedRoles(ctx context.Context, req *requests.FindAllRoles, data []*repository.RoleResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.RoleResult{}
	}
	key := fmt.Sprintf(roleAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &roleCachedResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleById(ctx context.Context, data *models.Role) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(roleByIdCacheKey, data.RoleID)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleByUserId(ctx context.Context, userId int, data []*models.Role) {
	if data == nil {
		data = []*models.Role{}
	}
	key := fmt.Sprintf(roleByUserIdCacheKey, userId)
	cache.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles, data []*repository.RoleResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.RoleResult{}
	}
	key := fmt.Sprintf(roleActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &roleCachedResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles, data []*repository.RoleResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.RoleResult{}
	}
	key := fmt.Sprintf(roleTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &roleCachedResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *roleQueryCache) GetCachedRoles(ctx context.Context, req *requests.FindAllRoles) ([]*repository.RoleResult, *int, bool) {
	key := fmt.Sprintf(roleAllCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[roleCachedResponse](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (m *roleQueryCache) GetCachedRoleById(ctx context.Context, id int) (*models.Role, bool) {
	key := fmt.Sprintf(roleByIdCacheKey, id)
	result, found := cache.GetFromCache[models.Role](ctx, m.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *roleQueryCache) GetCachedRoleByUserId(ctx context.Context, userId int) ([]*models.Role, bool) {
	key := fmt.Sprintf(roleByUserIdCacheKey, userId)
	result, found := cache.GetFromCache[[]*models.Role](ctx, m.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (m *roleQueryCache) GetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles) ([]*repository.RoleResult, *int, bool) {
	key := fmt.Sprintf(roleActiveCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[roleCachedResponse](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (m *roleQueryCache) GetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles) ([]*repository.RoleResult, *int, bool) {
	key := fmt.Sprintf(roleTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[roleCachedResponse](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}
