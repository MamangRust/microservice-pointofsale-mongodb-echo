package handler

import (
	"context"
	"fmt"

	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-stats-reader/repository"
	"go.uber.org/zap"
)

// CategoryStatsHandler serves category price aggregates from ClickHouse
// (order_item_daily). One handler struct registers all three category stats
// services (main, ByMerchant, ById).
type CategoryStatsHandler struct {
	statspb.UnimplementedCategoryStatsServiceServer
	statspb.UnimplementedCategoryStatsByMerchantServiceServer
	statspb.UnimplementedCategoryStatsByIdServiceServer
	repo  repository.Repository
	cache *StatsCache
	log   logger.LoggerInterface
}

func NewCategoryStatsHandler(repo repository.Repository, cache *StatsCache, log logger.LoggerInterface) *CategoryStatsHandler {
	return &CategoryStatsHandler{repo: repo, cache: cache, log: log}
}

func (h *CategoryStatsHandler) FindMonthlyTotalPrices(ctx context.Context, req *categorypb.FindYearMonthTotalPrices) (*categorypb.ApiResponseCategoryMonthlyTotalPrice, error) {
	key := fmt.Sprintf("stats:reader:category:monthly-total-prices:%d:%d", req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryMonthlyTotalPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalPrices(ctx, int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		h.log.Error("FindMonthlyTotalPrices failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryMonthlyTotalPrice{
		Status:  "success",
		Message: "Monthly total prices retrieved successfully",
		Data:    mapCategoryMonthTotalPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindYearlyTotalPrices(ctx context.Context, req *categorypb.FindYearTotalPrices) (*categorypb.ApiResponseCategoryYearlyTotalPrice, error) {
	key := fmt.Sprintf("stats:reader:category:yearly-total-prices:%d", req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryYearlyTotalPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalPrices(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearlyTotalPrices failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryYearlyTotalPrice{
		Status:  "success",
		Message: "Yearly total prices retrieved successfully",
		Data:    mapCategoryYearTotalPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindMonthPrice(ctx context.Context, req *categorypb.FindYearCategory) (*categorypb.ApiResponseCategoryMonthPrice, error) {
	key := fmt.Sprintf("stats:reader:category:month-price:%d", req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryMonthPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthPrice(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindMonthPrice failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryMonthPrice{
		Status:  "success",
		Message: "Monthly category prices retrieved successfully",
		Data:    mapCategoryMonthPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindYearPrice(ctx context.Context, req *categorypb.FindYearCategory) (*categorypb.ApiResponseCategoryYearPrice, error) {
	key := fmt.Sprintf("stats:reader:category:year-price:%d", req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryYearPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearPrice(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearPrice failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryYearPrice{
		Status:  "success",
		Message: "Yearly category prices retrieved successfully",
		Data:    mapCategoryYearPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindMonthlyTotalPricesByMerchant(ctx context.Context, req *categorypb.FindYearMonthTotalPriceByMerchant) (*categorypb.ApiResponseCategoryMonthlyTotalPrice, error) {
	key := fmt.Sprintf("stats:reader:category:monthly-total-prices:merchant:%d:%d:%d", req.GetMerchantId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryMonthlyTotalPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalPricesByMerchant(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthlyTotalPricesByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryMonthlyTotalPrice{
		Status:  "success",
		Message: "Monthly total prices by merchant retrieved successfully",
		Data:    mapCategoryMonthTotalPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindYearlyTotalPricesByMerchant(ctx context.Context, req *categorypb.FindYearTotalPriceByMerchant) (*categorypb.ApiResponseCategoryYearlyTotalPrice, error) {
	key := fmt.Sprintf("stats:reader:category:yearly-total-prices:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryYearlyTotalPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalPricesByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearlyTotalPricesByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryYearlyTotalPrice{
		Status:  "success",
		Message: "Yearly total prices by merchant retrieved successfully",
		Data:    mapCategoryYearTotalPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindMonthPriceByMerchant(ctx context.Context, req *categorypb.FindYearCategoryByMerchant) (*categorypb.ApiResponseCategoryMonthPrice, error) {
	key := fmt.Sprintf("stats:reader:category:month-price:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryMonthPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthPriceByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthPriceByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryMonthPrice{
		Status:  "success",
		Message: "Monthly category prices by merchant retrieved successfully",
		Data:    mapCategoryMonthPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindYearPriceByMerchant(ctx context.Context, req *categorypb.FindYearCategoryByMerchant) (*categorypb.ApiResponseCategoryYearPrice, error) {
	key := fmt.Sprintf("stats:reader:category:year-price:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryYearPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearPriceByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearPriceByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryYearPrice{
		Status:  "success",
		Message: "Yearly category prices by merchant retrieved successfully",
		Data:    mapCategoryYearPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindMonthlyTotalPricesById(ctx context.Context, req *categorypb.FindYearMonthTotalPriceById) (*categorypb.ApiResponseCategoryMonthlyTotalPrice, error) {
	key := fmt.Sprintf("stats:reader:category:monthly-total-prices:id:%d:%d:%d", req.GetCategoryId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryMonthlyTotalPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalPricesById(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetCategoryId()))
	if err != nil {
		h.log.Error("FindMonthlyTotalPricesById failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryMonthlyTotalPrice{
		Status:  "success",
		Message: "Monthly total prices by category retrieved successfully",
		Data:    mapCategoryMonthTotalPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindYearlyTotalPricesById(ctx context.Context, req *categorypb.FindYearTotalPriceById) (*categorypb.ApiResponseCategoryYearlyTotalPrice, error) {
	key := fmt.Sprintf("stats:reader:category:yearly-total-prices:id:%d:%d", req.GetCategoryId(), req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryYearlyTotalPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalPricesById(ctx, int(req.GetYear()), int(req.GetCategoryId()))
	if err != nil {
		h.log.Error("FindYearlyTotalPricesById failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryYearlyTotalPrice{
		Status:  "success",
		Message: "Yearly total prices by category retrieved successfully",
		Data:    mapCategoryYearTotalPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindMonthPriceById(ctx context.Context, req *categorypb.FindYearCategoryById) (*categorypb.ApiResponseCategoryMonthPrice, error) {
	key := fmt.Sprintf("stats:reader:category:month-price:id:%d:%d", req.GetCategoryId(), req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryMonthPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthPriceById(ctx, int(req.GetYear()), int(req.GetCategoryId()))
	if err != nil {
		h.log.Error("FindMonthPriceById failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryMonthPrice{
		Status:  "success",
		Message: "Monthly category prices by category retrieved successfully",
		Data:    mapCategoryMonthPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CategoryStatsHandler) FindYearPriceById(ctx context.Context, req *categorypb.FindYearCategoryById) (*categorypb.ApiResponseCategoryYearPrice, error) {
	key := fmt.Sprintf("stats:reader:category:year-price:id:%d:%d", req.GetCategoryId(), req.GetYear())
	if cached, found := CacheGet[categorypb.ApiResponseCategoryYearPrice](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearPriceById(ctx, int(req.GetYear()), int(req.GetCategoryId()))
	if err != nil {
		h.log.Error("FindYearPriceById failed", zap.Error(err))
		return nil, err
	}

	resp := &categorypb.ApiResponseCategoryYearPrice{
		Status:  "success",
		Message: "Yearly category prices by category retrieved successfully",
		Data:    mapCategoryYearPrice(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

// --- Mappers ---
//
// category_name is always empty: ClickHouse does not store category names; the
// front end resolves them via the category domain service.

func mapCategoryMonthTotalPrice(data []repository.CategoryMonthTotalPrice) []*categorypb.CategoriesMonthlyTotalPriceResponse {
	var out []*categorypb.CategoriesMonthlyTotalPriceResponse
	for _, d := range data {
		out = append(out, &categorypb.CategoriesMonthlyTotalPriceResponse{
			Year:         d.Year,
			Month:        d.Month,
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapCategoryYearTotalPrice(data []repository.CategoryYearTotalPrice) []*categorypb.CategoriesYearlyTotalPriceResponse {
	var out []*categorypb.CategoriesYearlyTotalPriceResponse
	for _, d := range data {
		out = append(out, &categorypb.CategoriesYearlyTotalPriceResponse{
			Year:         d.Year,
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapCategoryMonthPrice(data []repository.CategoryMonthPrice) []*categorypb.CategoryMonthPriceResponse {
	var out []*categorypb.CategoryMonthPriceResponse
	for _, d := range data {
		out = append(out, &categorypb.CategoryMonthPriceResponse{
			Month:        d.Month,
			CategoryId:   int32(d.CategoryID),
			CategoryName: "",
			OrderCount:   int32(d.OrderCount),
			ItemsSold:    int32(d.ItemsSold),
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapCategoryYearPrice(data []repository.CategoryYearPrice) []*categorypb.CategoryYearPriceResponse {
	var out []*categorypb.CategoryYearPriceResponse
	for _, d := range data {
		out = append(out, &categorypb.CategoryYearPriceResponse{
			Year:               d.Year,
			CategoryId:         int32(d.CategoryID),
			CategoryName:       "",
			OrderCount:         int32(d.OrderCount),
			ItemsSold:          int32(d.ItemsSold),
			TotalRevenue:       int32(d.TotalRevenue),
			UniqueProductsSold: int32(d.UniqueProductsSold),
		})
	}
	return out
}
