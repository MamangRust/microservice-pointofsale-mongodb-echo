package mencache

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

const (
	orderAllCacheKey      = "order:all:page:%d:pageSize:%d:search:%s"
	orderByIdCacheKey     = "order:id:%d"
	orderMerchantCacheKey = "order:merchant:%d:page:%d:pageSize:%d:search:%s"
	orderActiveCacheKey   = "order:active:page:%d:pageSize:%d:search:%s"
	orderTrashedCacheKey  = "order:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)

type orderCacheResponse struct {
	Data         []*repository.OrderResult `json:"data"`
	TotalRecords *int                      `json:"total_records"`
}

type orderCacheResponseByMerchant struct {
	Data         []*repository.OrderResult `json:"data"`
	TotalRecords *int                      `json:"total_records"`
}

type orderCacheResponseActive struct {
	Data         []*repository.OrderResultDeleteAt `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type orderCacheResponseTrashed struct {
	Data         []*repository.OrderResultDeleteAt `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type orderQueryCache struct {
	store *cache.CacheStore
}

func NewOrderQueryCache(store *cache.CacheStore) OrderQueryCache {
	return &orderQueryCache{store: store}
}

func (s *orderQueryCache) GetOrderAllCache(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResult, *int, bool) {
	key := fmt.Sprintf(orderAllCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[orderCacheResponse](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (s *orderQueryCache) SetOrderAllCache(ctx context.Context, req *requests.FindAllOrders, data []*repository.OrderResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.OrderResult{}
	}

	key := fmt.Sprintf(orderAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &orderCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *orderQueryCache) GetCachedOrderMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*repository.OrderResult, *int, bool) {
	key := fmt.Sprintf(orderMerchantCacheKey, req.MerchantID, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[orderCacheResponseByMerchant](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (s *orderQueryCache) SetCachedOrderMerchant(ctx context.Context, req *requests.FindAllOrderMerchant, res []*repository.OrderResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if res == nil {
		res = []*repository.OrderResult{}
	}

	key := fmt.Sprintf(orderMerchantCacheKey, req.MerchantID, req.Page, req.PageSize, req.Search)
	payload := &orderCacheResponseByMerchant{Data: res, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *orderQueryCache) GetOrderActiveCache(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(orderActiveCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[orderCacheResponseActive](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (s *orderQueryCache) SetOrderActiveCache(ctx context.Context, req *requests.FindAllOrders, data []*repository.OrderResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.OrderResultDeleteAt{}
	}

	key := fmt.Sprintf(orderActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &orderCacheResponseActive{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *orderQueryCache) GetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(orderTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[orderCacheResponseTrashed](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (s *orderQueryCache) SetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrders, data []*repository.OrderResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.OrderResultDeleteAt{}
	}

	key := fmt.Sprintf(orderTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &orderCacheResponseTrashed{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *orderQueryCache) GetCachedOrderCache(ctx context.Context, orderID int) (*models.Order, bool) {
	key := fmt.Sprintf(orderByIdCacheKey, orderID)
	result, found := cache.GetFromCache[models.Order](ctx, s.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *orderQueryCache) SetCachedOrderCache(ctx context.Context, data *models.Order) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(orderByIdCacheKey, data.OrderID)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
