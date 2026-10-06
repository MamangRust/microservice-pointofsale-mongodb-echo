package mencache

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

const (
	cashierAllCacheKey     = "cashier:all:page:%d:pageSize:%d:search:%s"
	cashierByIdCacheKey    = "cashier:id:%d"
	cashierActiveCacheKey  = "cashier:active:page:%d:pageSize:%d:search:%s"
	cashierTrashedCacheKey = "cashier:trashed:page:%d:pageSize:%d:search:%s"

	cashierByMerchantCacheKey = "cashier:merchant:%d:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)

type cashierCacheResponse struct {
	Data         []*repository.CashierResult `json:"data"`
	TotalRecords *int                        `json:"totalRecords"`
}

type cashierCacheResponseActive struct {
	Data         []*repository.CashierResultDeleteAt `json:"data"`
	TotalRecords *int                                `json:"totalRecords"`
}

type cashierCacheResponseTrashed struct {
	Data         []*repository.CashierResultDeleteAt `json:"data"`
	TotalRecords *int                                `json:"totalRecords"`
}

type cashierCacheResponseMerchant struct {
	Data         []*repository.CashierResult `json:"data"`
	TotalRecords *int                        `json:"totalRecords"`
}

type cashierQueryCache struct {
	store *cache.CacheStore
}

func NewCashierQueryCache(store *cache.CacheStore) CashierQueryCache {
	return &cashierQueryCache{store: store}
}

func (s *cashierQueryCache) GetCachedCashiersCache(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResult, *int, bool) {
	key := fmt.Sprintf(cashierAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[cashierCacheResponse](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (s *cashierQueryCache) SetCachedCashiersCache(ctx context.Context, req *requests.FindAllCashiers, data []*repository.CashierResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.CashierResult{}
	}

	key := fmt.Sprintf(cashierAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &cashierCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *cashierQueryCache) GetCachedCashiersByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*repository.CashierResult, *int, bool) {
	key := fmt.Sprintf(cashierByMerchantCacheKey, req.MerchantID, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[cashierCacheResponseMerchant](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (s *cashierQueryCache) SetCachedCashiersByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant, data []*repository.CashierResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.CashierResult{}
	}

	key := fmt.Sprintf(cashierByMerchantCacheKey, req.MerchantID, req.Page, req.PageSize, req.Search)
	payload := &cashierCacheResponseMerchant{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *cashierQueryCache) GetCachedCashiersActive(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(cashierActiveCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[cashierCacheResponseActive](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (s *cashierQueryCache) SetCachedCashiersActive(ctx context.Context, req *requests.FindAllCashiers, data []*repository.CashierResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.CashierResultDeleteAt{}
	}

	key := fmt.Sprintf(cashierActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &cashierCacheResponseActive{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *cashierQueryCache) GetCachedCashiersTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(cashierTrashedCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[cashierCacheResponseTrashed](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (s *cashierQueryCache) SetCachedCashiersTrashed(ctx context.Context, req *requests.FindAllCashiers, data []*repository.CashierResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.CashierResultDeleteAt{}
	}

	key := fmt.Sprintf(cashierTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &cashierCacheResponseTrashed{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *cashierQueryCache) GetCachedCashier(ctx context.Context, id int) (*models.Cashier, bool) {
	key := fmt.Sprintf(cashierByIdCacheKey, id)
	result, found := cache.GetFromCache[*models.Cashier](ctx, s.store, key)
	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (s *cashierQueryCache) SetCachedCashier(ctx context.Context, data *models.Cashier) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cashierByIdCacheKey, data.CashierID)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
