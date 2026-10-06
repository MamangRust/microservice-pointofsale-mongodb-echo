package handler

import (
	"context"
	"fmt"

	orderpb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-stats-reader/repository"
	"go.uber.org/zap"
)

// OrderStatsHandler serves revenue + order aggregates from ClickHouse
// (order_daily + order_item_daily). Results are cached in Redis for 5 minutes.
// One handler struct registers all three order stats services (main,
// ByMerchant, ById).
type OrderStatsHandler struct {
	statspb.UnimplementedOrderStatsServiceServer
	statspb.UnimplementedOrderStatsByMerchantServiceServer
	statspb.UnimplementedOrderStatsByIdServiceServer
	repo  repository.Repository
	cache *StatsCache
	log   logger.LoggerInterface
}

func NewOrderStatsHandler(repo repository.Repository, cache *StatsCache, log logger.LoggerInterface) *OrderStatsHandler {
	return &OrderStatsHandler{repo: repo, cache: cache, log: log}
}

func (h *OrderStatsHandler) FindMonthlyTotalRevenue(ctx context.Context, req *orderpb.FindYearMonthTotalRevenue) (*orderpb.ApiResponseOrderMonthlyTotalRevenue, error) {
	key := fmt.Sprintf("stats:reader:order:monthly-total-revenue:%d:%d", req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[orderpb.ApiResponseOrderMonthlyTotalRevenue](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalRevenue(ctx, int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		h.log.Error("FindMonthlyTotalRevenue failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderMonthlyTotalRevenue{
		Status:  "success",
		Message: "Monthly total revenue retrieved successfully",
		Data:    mapOrderMonthlyTotalRevenue(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindYearlyTotalRevenue(ctx context.Context, req *orderpb.FindYearTotalRevenue) (*orderpb.ApiResponseOrderYearlyTotalRevenue, error) {
	key := fmt.Sprintf("stats:reader:order:yearly-total-revenue:%d", req.GetYear())
	if cached, found := CacheGet[orderpb.ApiResponseOrderYearlyTotalRevenue](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalRevenue(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearlyTotalRevenue failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderYearlyTotalRevenue{
		Status:  "success",
		Message: "Yearly total revenue retrieved successfully",
		Data:    mapOrderYearlyTotalRevenue(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindMonthlyRevenue(ctx context.Context, req *orderpb.FindYearOrder) (*orderpb.ApiResponseOrderMonthly, error) {
	key := fmt.Sprintf("stats:reader:order:monthly-revenue:%d", req.GetYear())
	if cached, found := CacheGet[orderpb.ApiResponseOrderMonthly](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyRevenue(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindMonthlyRevenue failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderMonthly{
		Status:  "success",
		Message: "Monthly revenue retrieved successfully",
		Data:    mapOrderMonthly(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindYearlyRevenue(ctx context.Context, req *orderpb.FindYearOrder) (*orderpb.ApiResponseOrderYearly, error) {
	key := fmt.Sprintf("stats:reader:order:yearly-revenue:%d", req.GetYear())
	if cached, found := CacheGet[orderpb.ApiResponseOrderYearly](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyRevenue(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearlyRevenue failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderYearly{
		Status:  "success",
		Message: "Yearly revenue retrieved successfully",
		Data:    mapOrderYearly(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindMonthlyTotalRevenueByMerchant(ctx context.Context, req *orderpb.FindYearMonthTotalRevenueByMerchant) (*orderpb.ApiResponseOrderMonthlyTotalRevenue, error) {
	key := fmt.Sprintf("stats:reader:order:monthly-total-revenue:merchant:%d:%d:%d", req.GetMerchantId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[orderpb.ApiResponseOrderMonthlyTotalRevenue](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalRevenueByMerchant(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthlyTotalRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderMonthlyTotalRevenue{
		Status:  "success",
		Message: "Monthly total revenue by merchant retrieved successfully",
		Data:    mapOrderMonthlyTotalRevenue(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindYearlyTotalRevenueByMerchant(ctx context.Context, req *orderpb.FindYearTotalRevenueByMerchant) (*orderpb.ApiResponseOrderYearlyTotalRevenue, error) {
	key := fmt.Sprintf("stats:reader:order:yearly-total-revenue:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[orderpb.ApiResponseOrderYearlyTotalRevenue](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalRevenueByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearlyTotalRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderYearlyTotalRevenue{
		Status:  "success",
		Message: "Yearly total revenue by merchant retrieved successfully",
		Data:    mapOrderYearlyTotalRevenue(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindMonthlyRevenueByMerchant(ctx context.Context, req *orderpb.FindYearOrderByMerchant) (*orderpb.ApiResponseOrderMonthly, error) {
	key := fmt.Sprintf("stats:reader:order:monthly-revenue:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[orderpb.ApiResponseOrderMonthly](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyRevenueByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthlyRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderMonthly{
		Status:  "success",
		Message: "Monthly revenue by merchant retrieved successfully",
		Data:    mapOrderMonthly(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindYearlyRevenueByMerchant(ctx context.Context, req *orderpb.FindYearOrderByMerchant) (*orderpb.ApiResponseOrderYearly, error) {
	key := fmt.Sprintf("stats:reader:order:yearly-revenue:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[orderpb.ApiResponseOrderYearly](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyRevenueByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearlyRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderYearly{
		Status:  "success",
		Message: "Yearly revenue by merchant retrieved successfully",
		Data:    mapOrderYearly(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindMonthlyTotalRevenueById(ctx context.Context, req *orderpb.FindYearMonthTotalRevenueById) (*orderpb.ApiResponseOrderMonthlyTotalRevenue, error) {
	key := fmt.Sprintf("stats:reader:order:monthly-total-revenue:id:%d:%d:%d", req.GetOrderId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[orderpb.ApiResponseOrderMonthlyTotalRevenue](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalRevenueById(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetOrderId()))
	if err != nil {
		h.log.Error("FindMonthlyTotalRevenueById failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderMonthlyTotalRevenue{
		Status:  "success",
		Message: "Monthly total revenue by order retrieved successfully",
		Data:    mapOrderMonthlyTotalRevenue(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *OrderStatsHandler) FindYearlyTotalRevenueById(ctx context.Context, req *orderpb.FindYearTotalRevenueById) (*orderpb.ApiResponseOrderYearlyTotalRevenue, error) {
	key := fmt.Sprintf("stats:reader:order:yearly-total-revenue:id:%d:%d", req.GetOrderId(), req.GetYear())
	if cached, found := CacheGet[orderpb.ApiResponseOrderYearlyTotalRevenue](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalRevenueById(ctx, int(req.GetYear()), int(req.GetOrderId()))
	if err != nil {
		h.log.Error("FindYearlyTotalRevenueById failed", zap.Error(err))
		return nil, err
	}

	resp := &orderpb.ApiResponseOrderYearlyTotalRevenue{
		Status:  "success",
		Message: "Yearly total revenue by order retrieved successfully",
		Data:    mapOrderYearlyTotalRevenue(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

// --- Mappers ---

func mapOrderMonthlyTotalRevenue(data []repository.OrderMonthlyTotalRevenue) []*orderpb.OrderMonthlyTotalRevenueResponse {
	var out []*orderpb.OrderMonthlyTotalRevenueResponse
	for _, d := range data {
		out = append(out, &orderpb.OrderMonthlyTotalRevenueResponse{
			Year:           d.Year,
			Month:          d.Month,
			OrderCount:     int32(d.OrderCount),
			TotalRevenue:   int32(d.TotalRevenue),
			TotalItemsSold: int32(d.TotalItemsSold),
		})
	}
	return out
}

func mapOrderYearlyTotalRevenue(data []repository.OrderYearlyTotalRevenue) []*orderpb.OrderYearlyTotalRevenueResponse {
	var out []*orderpb.OrderYearlyTotalRevenueResponse
	for _, d := range data {
		out = append(out, &orderpb.OrderYearlyTotalRevenueResponse{
			Year:               d.Year,
			OrderCount:         int32(d.OrderCount),
			TotalRevenue:       int32(d.TotalRevenue),
			TotalItemsSold:     int32(d.TotalItemsSold),
			ActiveCashiers:     int32(d.ActiveCashiers),
			UniqueProductsSold: int32(d.UniqueProductsSold),
		})
	}
	return out
}

func mapOrderMonthly(data []repository.OrderMonthly) []*orderpb.OrderMonthlyResponse {
	var out []*orderpb.OrderMonthlyResponse
	for _, d := range data {
		out = append(out, &orderpb.OrderMonthlyResponse{
			Month:          d.Month,
			OrderCount:     int32(d.OrderCount),
			TotalRevenue:   int32(d.TotalRevenue),
			TotalItemsSold: int32(d.TotalItemsSold),
		})
	}
	return out
}

func mapOrderYearly(data []repository.OrderYearly) []*orderpb.OrderYearlyResponse {
	var out []*orderpb.OrderYearlyResponse
	for _, d := range data {
		out = append(out, &orderpb.OrderYearlyResponse{
			Year:               d.Year,
			OrderCount:         int32(d.OrderCount),
			TotalRevenue:       int32(d.TotalRevenue),
			TotalItemsSold:     int32(d.TotalItemsSold),
			ActiveCashiers:     int32(d.ActiveCashiers),
			UniqueProductsSold: int32(d.UniqueProductsSold),
		})
	}
	return out
}
