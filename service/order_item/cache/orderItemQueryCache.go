package mencache

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

const (
	orderItemAllCacheKey     = "order_item:all:page:%d:pageSize:%d:search:%s"
	orderItemActiveCacheKey  = "order_item:active:page:%d:pageSize:%d:search:%s"
	orderItemTrashedCacheKey = "order_item:trashed:page:%d:pageSize:%d:search:%s"

	orderItemByIdCacheKey = "order_item:id:%d"

	ttlDefault = 5 * time.Minute
)

type orderItemQueryCacheResponse struct {
	Data         []*repository.OrderItemResult `json:"data"`
	TotalRecords *int                          `json:"total_records"`
}

type orderItemQueryCacheResponseActive struct {
	Data         []*repository.OrderItemResultDeleteAt `json:"data"`
	TotalRecords *int                                  `json:"total_records"`
}

type orderItemQueryCacheResponseTrashed struct {
	Data         []*repository.OrderItemResultDeleteAt `json:"data"`
	TotalRecords *int                                  `json:"total_records"`
}

type orderItemQueryCache struct {
	store *cache.CacheStore
}

func NewOrderItemQueryCache(store *cache.CacheStore) OrderItemQueryCache {
	return &orderItemQueryCache{store: store}
}

func (o *orderItemQueryCache) GetCachedOrderItemsAll(ctx context.Context, req *requests.FindAllOrderItems) ([]*repository.OrderItemResult, *int, bool) {
	key := fmt.Sprintf(orderItemAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[orderItemQueryCacheResponse](ctx, o.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (o *orderItemQueryCache) SetCachedOrderItemsAll(ctx context.Context, req *requests.FindAllOrderItems, data []*repository.OrderItemResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.OrderItemResult{}
	}

	key := fmt.Sprintf(orderItemAllCacheKey, req.Page, req.PageSize, req.Search)

	payload := &orderItemQueryCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, o.store, key, payload, ttlDefault)
}

func (o *orderItemQueryCache) GetCachedOrderItemActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*repository.OrderItemResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(orderItemActiveCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[orderItemQueryCacheResponseActive](ctx, o.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (o *orderItemQueryCache) SetCachedOrderItemActive(ctx context.Context, req *requests.FindAllOrderItems, data []*repository.OrderItemResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.OrderItemResultDeleteAt{}
	}

	key := fmt.Sprintf(orderItemActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &orderItemQueryCacheResponseActive{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, o.store, key, payload, ttlDefault)
}

func (o *orderItemQueryCache) GetCachedOrderItemTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*repository.OrderItemResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(orderItemTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[orderItemQueryCacheResponseTrashed](ctx, o.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (o *orderItemQueryCache) SetCachedOrderItemTrashed(ctx context.Context, req *requests.FindAllOrderItems, data []*repository.OrderItemResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.OrderItemResultDeleteAt{}
	}

	key := fmt.Sprintf(orderItemTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &orderItemQueryCacheResponseTrashed{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, o.store, key, payload, ttlDefault)
}

func (o *orderItemQueryCache) GetCachedOrderItems(ctx context.Context, order_id int) ([]*models.OrderItem, bool) {
	key := fmt.Sprintf(orderItemByIdCacheKey, order_id)
	result, found := cache.GetFromCache[[]*models.OrderItem](ctx, o.store, key)
	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (o *orderItemQueryCache) SetCachedOrderItems(ctx context.Context, data []*models.OrderItem) {
	if len(data) == 0 {
		return
	}

	key := fmt.Sprintf(orderItemByIdCacheKey, data[0].OrderID)
	cache.SetToCache(ctx, o.store, key, &data, ttlDefault)
}
