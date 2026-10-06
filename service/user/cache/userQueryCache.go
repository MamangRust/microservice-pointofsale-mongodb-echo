package mencache

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-user/repository"
)

const (
	userAllCacheKey     = "user:all:page:%d:pageSize:%d:search:%s"
	userByIdCacheKey    = "user:id:%d"
	userActiveCacheKey  = "user:active:page:%d:pageSize:%d:search:%s"
	userTrashedCacheKey = "user:trashed:page:%d:pageSize:%d:search:%s"
	userTTL             = 5 * time.Minute
)

type userCachedResponse struct {
	Data         []*repository.UserResult `json:"data"`
	TotalRecords *int                     `json:"total_records"`
}

type userQueryCache struct {
	store *cache.CacheStore
}

func NewUserQueryCache(store *cache.CacheStore) UserQueryCache {
	return &userQueryCache{store: store}
}

func (c *userQueryCache) GetCachedUsersCache(ctx context.Context, req *requests.FindAllUsers) ([]*repository.UserResult, *int, bool) {
	key := fmt.Sprintf(userAllCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[userCachedResponse](ctx, c.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (c *userQueryCache) SetCachedUsersCache(ctx context.Context, req *requests.FindAllUsers, data []*repository.UserResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.UserResult{}
	}
	key := fmt.Sprintf(userAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &userCachedResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, c.store, key, payload, userTTL)
}

func (c *userQueryCache) GetCachedUserCache(ctx context.Context, id int) (*models.User, bool) {
	key := fmt.Sprintf(userByIdCacheKey, id)
	result, found := cache.GetFromCache[models.User](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *userQueryCache) SetCachedUserCache(ctx context.Context, data *models.User) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(userByIdCacheKey, data.UserID)
	cache.SetToCache(ctx, c.store, key, data, userTTL)
}

func (c *userQueryCache) GetCachedUserActiveCache(ctx context.Context, req *requests.FindAllUsers) ([]*repository.UserResult, *int, bool) {
	key := fmt.Sprintf(userActiveCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[userCachedResponse](ctx, c.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (c *userQueryCache) SetCachedUserActiveCache(ctx context.Context, req *requests.FindAllUsers, data []*repository.UserResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.UserResult{}
	}
	key := fmt.Sprintf(userActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &userCachedResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, c.store, key, payload, userTTL)
}

func (c *userQueryCache) GetCachedUserTrashedCache(ctx context.Context, req *requests.FindAllUsers) ([]*repository.UserResult, *int, bool) {
	key := fmt.Sprintf(userTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[userCachedResponse](ctx, c.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (c *userQueryCache) SetCachedUserTrashedCache(ctx context.Context, req *requests.FindAllUsers, data []*repository.UserResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.UserResult{}
	}
	key := fmt.Sprintf(userTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &userCachedResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, c.store, key, payload, userTTL)
}
