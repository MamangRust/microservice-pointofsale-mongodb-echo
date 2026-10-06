package handler

import (
	"context"
	"fmt"

	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-stats-reader/repository"
	"go.uber.org/zap"
)

// CashierStatsHandler serves sales aggregates from ClickHouse (order_daily).
// One handler struct registers all three cashier stats services (main,
// ByMerchant, ById).
type CashierStatsHandler struct {
	statspb.UnimplementedCashierStatsServiceServer
	statspb.UnimplementedCashierStatsByMerchantServiceServer
	statspb.UnimplementedCashierStatsByIdServiceServer
	repo  repository.Repository
	cache *StatsCache
	log   logger.LoggerInterface
}

func NewCashierStatsHandler(repo repository.Repository, cache *StatsCache, log logger.LoggerInterface) *CashierStatsHandler {
	return &CashierStatsHandler{repo: repo, cache: cache, log: log}
}

func (h *CashierStatsHandler) FindMonthlyTotalSales(ctx context.Context, req *cashierpb.FindYearMonthTotalSales) (*cashierpb.ApiResponseCashierMonthlyTotalSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:monthly-total-sales:%d:%d", req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierMonthlyTotalSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalSales(ctx, int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		h.log.Error("FindMonthlyTotalSales failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierMonthlyTotalSales{
		Status:  "success",
		Message: "Monthly total sales retrieved successfully",
		Data:    mapCashierMonthTotalSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindYearlyTotalSales(ctx context.Context, req *cashierpb.FindYearTotalSales) (*cashierpb.ApiResponseCashierYearlyTotalSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:yearly-total-sales:%d", req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierYearlyTotalSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalSales(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearlyTotalSales failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierYearlyTotalSales{
		Status:  "success",
		Message: "Yearly total sales retrieved successfully",
		Data:    mapCashierYearTotalSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindMonthSales(ctx context.Context, req *cashierpb.FindYearCashier) (*cashierpb.ApiResponseCashierMonthSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:month-sales:%d", req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierMonthSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthSales(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindMonthSales failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierMonthSales{
		Status:  "success",
		Message: "Monthly cashier sales retrieved successfully",
		Data:    mapCashierMonthSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindYearSales(ctx context.Context, req *cashierpb.FindYearCashier) (*cashierpb.ApiResponseCashierYearSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:year-sales:%d", req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierYearSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearSales(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearSales failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierYearSales{
		Status:  "success",
		Message: "Yearly cashier sales retrieved successfully",
		Data:    mapCashierYearSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindMonthlyTotalSalesByMerchant(ctx context.Context, req *cashierpb.FindYearMonthTotalSalesByMerchant) (*cashierpb.ApiResponseCashierMonthlyTotalSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:monthly-total-sales:merchant:%d:%d:%d", req.GetMerchantId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierMonthlyTotalSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalSalesByMerchant(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthlyTotalSalesByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierMonthlyTotalSales{
		Status:  "success",
		Message: "Monthly total sales by merchant retrieved successfully",
		Data:    mapCashierMonthTotalSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindYearlyTotalSalesByMerchant(ctx context.Context, req *cashierpb.FindYearTotalSalesByMerchant) (*cashierpb.ApiResponseCashierYearlyTotalSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:yearly-total-sales:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierYearlyTotalSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalSalesByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearlyTotalSalesByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierYearlyTotalSales{
		Status:  "success",
		Message: "Yearly total sales by merchant retrieved successfully",
		Data:    mapCashierYearTotalSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindMonthSalesByMerchant(ctx context.Context, req *cashierpb.FindYearCashierByMerchant) (*cashierpb.ApiResponseCashierMonthSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:month-sales:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierMonthSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthSalesByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthSalesByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierMonthSales{
		Status:  "success",
		Message: "Monthly cashier sales by merchant retrieved successfully",
		Data:    mapCashierMonthSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindYearSalesByMerchant(ctx context.Context, req *cashierpb.FindYearCashierByMerchant) (*cashierpb.ApiResponseCashierYearSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:year-sales:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierYearSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearSalesByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearSalesByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierYearSales{
		Status:  "success",
		Message: "Yearly cashier sales by merchant retrieved successfully",
		Data:    mapCashierYearSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindMonthlyTotalSalesById(ctx context.Context, req *cashierpb.FindYearMonthTotalSalesById) (*cashierpb.ApiResponseCashierMonthlyTotalSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:monthly-total-sales:id:%d:%d:%d", req.GetCashierId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierMonthlyTotalSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthlyTotalSalesById(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetCashierId()))
	if err != nil {
		h.log.Error("FindMonthlyTotalSalesById failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierMonthlyTotalSales{
		Status:  "success",
		Message: "Monthly total sales by cashier retrieved successfully",
		Data:    mapCashierMonthTotalSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindYearlyTotalSalesById(ctx context.Context, req *cashierpb.FindYearTotalSalesById) (*cashierpb.ApiResponseCashierYearlyTotalSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:yearly-total-sales:id:%d:%d", req.GetCashierId(), req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierYearlyTotalSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearlyTotalSalesById(ctx, int(req.GetYear()), int(req.GetCashierId()))
	if err != nil {
		h.log.Error("FindYearlyTotalSalesById failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierYearlyTotalSales{
		Status:  "success",
		Message: "Yearly total sales by cashier retrieved successfully",
		Data:    mapCashierYearTotalSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindMonthSalesById(ctx context.Context, req *cashierpb.FindYearCashierById) (*cashierpb.ApiResponseCashierMonthSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:month-sales:id:%d:%d", req.GetCashierId(), req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierMonthSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthSalesById(ctx, int(req.GetYear()), int(req.GetCashierId()))
	if err != nil {
		h.log.Error("FindMonthSalesById failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierMonthSales{
		Status:  "success",
		Message: "Monthly cashier sales by cashier retrieved successfully",
		Data:    mapCashierMonthSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *CashierStatsHandler) FindYearSalesById(ctx context.Context, req *cashierpb.FindYearCashierById) (*cashierpb.ApiResponseCashierYearSales, error) {
	key := fmt.Sprintf("stats:reader:cashier:year-sales:id:%d:%d", req.GetCashierId(), req.GetYear())
	if cached, found := CacheGet[cashierpb.ApiResponseCashierYearSales](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearSalesById(ctx, int(req.GetYear()), int(req.GetCashierId()))
	if err != nil {
		h.log.Error("FindYearSalesById failed", zap.Error(err))
		return nil, err
	}

	resp := &cashierpb.ApiResponseCashierYearSales{
		Status:  "success",
		Message: "Yearly cashier sales by cashier retrieved successfully",
		Data:    mapCashierYearSales(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

// --- Mappers ---
//
// cashier_name is always empty: ClickHouse does not store cashier names; the
// front end resolves them via the cashier domain service.

func mapCashierMonthTotalSales(data []repository.CashierMonthTotalSales) []*cashierpb.CashierResponseMonthTotalSales {
	var out []*cashierpb.CashierResponseMonthTotalSales
	for _, d := range data {
		out = append(out, &cashierpb.CashierResponseMonthTotalSales{
			Year:       d.Year,
			Month:      d.Month,
			TotalSales: int32(d.TotalSales),
		})
	}
	return out
}

func mapCashierYearTotalSales(data []repository.CashierYearTotalSales) []*cashierpb.CashierResponseYearTotalSales {
	var out []*cashierpb.CashierResponseYearTotalSales
	for _, d := range data {
		out = append(out, &cashierpb.CashierResponseYearTotalSales{
			Year:       d.Year,
			TotalSales: int32(d.TotalSales),
		})
	}
	return out
}

func mapCashierMonthSales(data []repository.CashierMonthSales) []*cashierpb.CashierResponseMonthSales {
	var out []*cashierpb.CashierResponseMonthSales
	for _, d := range data {
		out = append(out, &cashierpb.CashierResponseMonthSales{
			Month:       d.Month,
			CashierId:   int32(d.CashierID),
			CashierName: "",
			OrderCount:  int32(d.OrderCount),
			TotalSales:  int32(d.TotalSales),
		})
	}
	return out
}

func mapCashierYearSales(data []repository.CashierYearSales) []*cashierpb.CashierResponseYearSales {
	var out []*cashierpb.CashierResponseYearSales
	for _, d := range data {
		out = append(out, &cashierpb.CashierResponseYearSales{
			Year:        d.Year,
			CashierId:   int32(d.CashierID),
			CashierName: "",
			OrderCount:  int32(d.OrderCount),
			TotalSales:  int32(d.TotalSales),
		})
	}
	return out
}
