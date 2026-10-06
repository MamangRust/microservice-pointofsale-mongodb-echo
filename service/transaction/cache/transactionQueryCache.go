package mencache

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-transacton/repository"
)

const (
	transactionAllCacheKey  = "transaction:all:page:%d:pageSize:%d:search:%s"
	transactionByIdCacheKey = "transaction:id:%d"

	transactionByMerchantCacheKey = "transaction:merchant:%d:page:%d:pageSize:%d:search:%s"

	transactionActiveCacheKey  = "transaction:active:page:%d:pageSize:%d:search:%s"
	transactionTrashedCacheKey = "transaction:trashed:page:%d:pageSize:%d:search:%s"

	transactionByOrderCacheKey = "transaction:order:%d"

	ttlDefault = 5 * time.Minute
)

type transactionCacheResponse struct {
	Data         []*repository.TransactionResult `json:"data"`
	TotalRecords *int                            `json:"totalRecords"`
}

type transactionMerchantCacheResponse struct {
	Data         []*repository.TransactionByMerchantResult `json:"data"`
	TotalRecords *int                                      `json:"totalRecords"`
}

type transactionCacheResponseActive struct {
	Data         []*repository.TransactionResultDeleteAt `json:"data"`
	TotalRecords *int                                    `json:"totalRecords"`
}

type transactionCacheResponseTrashed struct {
	Data         []*repository.TransactionResultDeleteAt `json:"data"`
	TotalRecords *int                                    `json:"totalRecords"`
}

type transactionQueryCache struct {
	store *cache.CacheStore
}

func NewTransactionQueryCache(store *cache.CacheStore) TransactionQueryCache {
	return &transactionQueryCache{store: store}
}

func (t *transactionQueryCache) GetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransaction) ([]*repository.TransactionResult, *int, bool) {
	key := fmt.Sprintf(transactionAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[transactionCacheResponse](ctx, t.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (t *transactionQueryCache) SetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransaction, data []*repository.TransactionResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.TransactionResult{}
	}

	key := fmt.Sprintf(transactionAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &transactionCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, t.store, key, payload, ttlDefault)
}

func (t *transactionQueryCache) GetCachedTransactionByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*repository.TransactionByMerchantResult, *int, bool) {
	key := fmt.Sprintf(transactionByMerchantCacheKey, req.MerchantID, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[transactionMerchantCacheResponse](ctx, t.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (t *transactionQueryCache) SetCachedTransactionByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant, data []*repository.TransactionByMerchantResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.TransactionByMerchantResult{}
	}

	key := fmt.Sprintf(transactionByMerchantCacheKey, req.MerchantID, req.Page, req.PageSize, req.Search)
	payload := &transactionMerchantCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, t.store, key, payload, ttlDefault)
}

func (t *transactionQueryCache) GetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransaction) ([]*repository.TransactionResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(transactionActiveCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[transactionCacheResponseActive](ctx, t.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (t *transactionQueryCache) SetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransaction, data []*repository.TransactionResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.TransactionResultDeleteAt{}
	}

	key := fmt.Sprintf(transactionActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &transactionCacheResponseActive{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, t.store, key, payload, ttlDefault)
}

func (t *transactionQueryCache) GetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransaction) ([]*repository.TransactionResultDeleteAt, *int, bool) {
	key := fmt.Sprintf(transactionTrashedCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[transactionCacheResponseTrashed](ctx, t.store, key)
	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (t *transactionQueryCache) SetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransaction, data []*repository.TransactionResultDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*repository.TransactionResultDeleteAt{}
	}

	key := fmt.Sprintf(transactionTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &transactionCacheResponseTrashed{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, t.store, key, payload, ttlDefault)
}

func (t *transactionQueryCache) GetCachedTransactionCache(ctx context.Context, id int) (*models.Transaction, bool) {
	key := fmt.Sprintf(transactionByIdCacheKey, id)

	result, found := cache.GetFromCache[models.Transaction](ctx, t.store, key)
	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (t *transactionQueryCache) SetCachedTransactionCache(ctx context.Context, data *models.Transaction) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transactionByIdCacheKey, data.TransactionID)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionQueryCache) GetCachedTransactionByOrderId(ctx context.Context, orderID int) (*models.Transaction, bool) {
	key := fmt.Sprintf(transactionByOrderCacheKey, orderID)

	result, found := cache.GetFromCache[models.Transaction](ctx, t.store, key)
	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (t *transactionQueryCache) SetCachedTransactionByOrderId(ctx context.Context, orderID int, data *models.Transaction) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transactionByOrderCacheKey, orderID)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
