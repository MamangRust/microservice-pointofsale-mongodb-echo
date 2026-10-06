package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	orderpb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	transactionpb "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
)

// StatsClients bundles the 11 stats-reader gRPC clients (F4 §7.4). The
// transaction ByMerchant rpcs reuse the same Status/Method clients as the main
// service.
type StatsClients struct {
	Order              statspb.OrderStatsServiceClient
	OrderByMerchant    statspb.OrderStatsByMerchantServiceClient
	OrderById          statspb.OrderStatsByIdServiceClient
	Cashier            statspb.CashierStatsServiceClient
	CashierByMerchant  statspb.CashierStatsByMerchantServiceClient
	CashierById        statspb.CashierStatsByIdServiceClient
	Category           statspb.CategoryStatsServiceClient
	CategoryByMerchant statspb.CategoryStatsByMerchantServiceClient
	CategoryById       statspb.CategoryStatsByIdServiceClient
	TransactionStatus  statspb.TransactionStatsStatusServiceClient
	TransactionMethod  statspb.TransactionStatsMethodServiceClient
}

// statsHandleApi exposes the stats-reader (ClickHouse) gRPC contracts as REST
// endpoints (F4 §7.4). All cache keys use the "apigw:" namespace so the gateway
// can never overwrite domain-service or stats-reader cache keys (§8.1).
type statsHandleApi struct {
	clients    *StatsClients
	logger     logger.LoggerInterface
	mapping    response_api.StatsResponseMapper
	apiHandler errors.ApiHandler
	cache      *gateway_cache.GatewayCache
}

func NewHandlerStats(
	router *echo.Echo,
	clients *StatsClients,
	logger logger.LoggerInterface,
	mapping response_api.StatsResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *statsHandleApi {
	statsHandler := &statsHandleApi{
		clients:    clients,
		logger:     logger,
		mapping:    mapping,
		apiHandler: apiHandler,
		cache:      cache,
	}

	// ── Order stats routes ───────────────────────────────────────────────
	orderStats := router.Group("/api/order-stats")
	orderStats.GET("/monthly-total-revenue", apiHandler.Handle("get-order-stats-monthly-total-revenue", statsHandler.FindOrderMonthlyTotalRevenue))
	orderStats.GET("/yearly-total-revenue", apiHandler.Handle("get-order-stats-yearly-total-revenue", statsHandler.FindOrderYearlyTotalRevenue))
	orderStats.GET("/monthly-revenue", apiHandler.Handle("get-order-stats-monthly-revenue", statsHandler.FindOrderMonthlyRevenue))
	orderStats.GET("/yearly-revenue", apiHandler.Handle("get-order-stats-yearly-revenue", statsHandler.FindOrderYearlyRevenue))
	orderStats.GET("/by-merchant/monthly-total-revenue", apiHandler.Handle("get-order-stats-by-merchant-monthly-total-revenue", statsHandler.FindOrderMonthlyTotalRevenueByMerchant))
	orderStats.GET("/by-merchant/yearly-total-revenue", apiHandler.Handle("get-order-stats-by-merchant-yearly-total-revenue", statsHandler.FindOrderYearlyTotalRevenueByMerchant))
	orderStats.GET("/by-merchant/monthly-revenue", apiHandler.Handle("get-order-stats-by-merchant-monthly-revenue", statsHandler.FindOrderMonthlyRevenueByMerchant))
	orderStats.GET("/by-merchant/yearly-revenue", apiHandler.Handle("get-order-stats-by-merchant-yearly-revenue", statsHandler.FindOrderYearlyRevenueByMerchant))
	orderStats.GET("/by-id/monthly-total-revenue", apiHandler.Handle("get-order-stats-by-id-monthly-total-revenue", statsHandler.FindOrderMonthlyTotalRevenueById))
	orderStats.GET("/by-id/yearly-total-revenue", apiHandler.Handle("get-order-stats-by-id-yearly-total-revenue", statsHandler.FindOrderYearlyTotalRevenueById))

	// ── Cashier stats routes ─────────────────────────────────────────────
	cashierStats := router.Group("/api/cashier-stats")
	cashierStats.GET("/monthly-total-sales", apiHandler.Handle("get-cashier-stats-monthly-total-sales", statsHandler.FindCashierMonthlyTotalSales))
	cashierStats.GET("/yearly-total-sales", apiHandler.Handle("get-cashier-stats-yearly-total-sales", statsHandler.FindCashierYearlyTotalSales))
	cashierStats.GET("/month-sales", apiHandler.Handle("get-cashier-stats-month-sales", statsHandler.FindCashierMonthSales))
	cashierStats.GET("/year-sales", apiHandler.Handle("get-cashier-stats-year-sales", statsHandler.FindCashierYearSales))
	cashierStats.GET("/by-merchant/monthly-total-sales", apiHandler.Handle("get-cashier-stats-by-merchant-monthly-total-sales", statsHandler.FindCashierMonthlyTotalSalesByMerchant))
	cashierStats.GET("/by-merchant/yearly-total-sales", apiHandler.Handle("get-cashier-stats-by-merchant-yearly-total-sales", statsHandler.FindCashierYearlyTotalSalesByMerchant))
	cashierStats.GET("/by-merchant/month-sales", apiHandler.Handle("get-cashier-stats-by-merchant-month-sales", statsHandler.FindCashierMonthSalesByMerchant))
	cashierStats.GET("/by-merchant/year-sales", apiHandler.Handle("get-cashier-stats-by-merchant-year-sales", statsHandler.FindCashierYearSalesByMerchant))
	cashierStats.GET("/by-id/monthly-total-sales", apiHandler.Handle("get-cashier-stats-by-id-monthly-total-sales", statsHandler.FindCashierMonthlyTotalSalesById))
	cashierStats.GET("/by-id/yearly-total-sales", apiHandler.Handle("get-cashier-stats-by-id-yearly-total-sales", statsHandler.FindCashierYearlyTotalSalesById))
	cashierStats.GET("/by-id/month-sales", apiHandler.Handle("get-cashier-stats-by-id-month-sales", statsHandler.FindCashierMonthSalesById))
	cashierStats.GET("/by-id/year-sales", apiHandler.Handle("get-cashier-stats-by-id-year-sales", statsHandler.FindCashierYearSalesById))

	// ── Category stats routes ────────────────────────────────────────────
	categoryStats := router.Group("/api/category-stats")
	categoryStats.GET("/monthly-total-prices", apiHandler.Handle("get-category-stats-monthly-total-prices", statsHandler.FindCategoryMonthlyTotalPrices))
	categoryStats.GET("/yearly-total-prices", apiHandler.Handle("get-category-stats-yearly-total-prices", statsHandler.FindCategoryYearlyTotalPrices))
	categoryStats.GET("/month-price", apiHandler.Handle("get-category-stats-month-price", statsHandler.FindCategoryMonthPrice))
	categoryStats.GET("/year-price", apiHandler.Handle("get-category-stats-year-price", statsHandler.FindCategoryYearPrice))
	categoryStats.GET("/by-merchant/monthly-total-prices", apiHandler.Handle("get-category-stats-by-merchant-monthly-total-prices", statsHandler.FindCategoryMonthlyTotalPricesByMerchant))
	categoryStats.GET("/by-merchant/yearly-total-prices", apiHandler.Handle("get-category-stats-by-merchant-yearly-total-prices", statsHandler.FindCategoryYearlyTotalPricesByMerchant))
	categoryStats.GET("/by-merchant/month-price", apiHandler.Handle("get-category-stats-by-merchant-month-price", statsHandler.FindCategoryMonthPriceByMerchant))
	categoryStats.GET("/by-merchant/year-price", apiHandler.Handle("get-category-stats-by-merchant-year-price", statsHandler.FindCategoryYearPriceByMerchant))
	categoryStats.GET("/by-id/monthly-total-prices", apiHandler.Handle("get-category-stats-by-id-monthly-total-prices", statsHandler.FindCategoryMonthlyTotalPricesById))
	categoryStats.GET("/by-id/yearly-total-prices", apiHandler.Handle("get-category-stats-by-id-yearly-total-prices", statsHandler.FindCategoryYearlyTotalPricesById))
	categoryStats.GET("/by-id/month-price", apiHandler.Handle("get-category-stats-by-id-month-price", statsHandler.FindCategoryMonthPriceById))
	categoryStats.GET("/by-id/year-price", apiHandler.Handle("get-category-stats-by-id-year-price", statsHandler.FindCategoryYearPriceById))

	// ── Transaction stats routes ─────────────────────────────────────────
	transactionStats := router.Group("/api/transaction-stats")
	transactionStats.GET("/month-status-success", apiHandler.Handle("get-transaction-stats-month-status-success", statsHandler.FindTransactionMonthStatusSuccess))
	transactionStats.GET("/year-status-success", apiHandler.Handle("get-transaction-stats-year-status-success", statsHandler.FindTransactionYearStatusSuccess))
	transactionStats.GET("/month-status-failed", apiHandler.Handle("get-transaction-stats-month-status-failed", statsHandler.FindTransactionMonthStatusFailed))
	transactionStats.GET("/year-status-failed", apiHandler.Handle("get-transaction-stats-year-status-failed", statsHandler.FindTransactionYearStatusFailed))
	transactionStats.GET("/month-method-success", apiHandler.Handle("get-transaction-stats-month-method-success", statsHandler.FindTransactionMonthMethodSuccess))
	transactionStats.GET("/year-method-success", apiHandler.Handle("get-transaction-stats-year-method-success", statsHandler.FindTransactionYearMethodSuccess))
	transactionStats.GET("/month-method-failed", apiHandler.Handle("get-transaction-stats-month-method-failed", statsHandler.FindTransactionMonthMethodFailed))
	transactionStats.GET("/year-method-failed", apiHandler.Handle("get-transaction-stats-year-method-failed", statsHandler.FindTransactionYearMethodFailed))
	transactionStats.GET("/by-merchant/month-status-success", apiHandler.Handle("get-transaction-stats-by-merchant-month-status-success", statsHandler.FindTransactionMonthStatusSuccessByMerchant))
	transactionStats.GET("/by-merchant/year-status-success", apiHandler.Handle("get-transaction-stats-by-merchant-year-status-success", statsHandler.FindTransactionYearStatusSuccessByMerchant))
	transactionStats.GET("/by-merchant/month-status-failed", apiHandler.Handle("get-transaction-stats-by-merchant-month-status-failed", statsHandler.FindTransactionMonthStatusFailedByMerchant))
	transactionStats.GET("/by-merchant/year-status-failed", apiHandler.Handle("get-transaction-stats-by-merchant-year-status-failed", statsHandler.FindTransactionYearStatusFailedByMerchant))
	transactionStats.GET("/by-merchant/month-method-success", apiHandler.Handle("get-transaction-stats-by-merchant-month-method-success", statsHandler.FindTransactionMonthMethodSuccessByMerchant))
	transactionStats.GET("/by-merchant/year-method-success", apiHandler.Handle("get-transaction-stats-by-merchant-year-method-success", statsHandler.FindTransactionYearMethodSuccessByMerchant))
	transactionStats.GET("/by-merchant/month-method-failed", apiHandler.Handle("get-transaction-stats-by-merchant-month-method-failed", statsHandler.FindTransactionMonthMethodFailedByMerchant))
	transactionStats.GET("/by-merchant/year-method-failed", apiHandler.Handle("get-transaction-stats-by-merchant-year-method-failed", statsHandler.FindTransactionYearMethodFailedByMerchant))

	return statsHandler
}

// ── Param helpers ───────────────────────────────────────────────────────────

func (h *statsHandleApi) parseYear(c echo.Context) (int, error) {
	year, err := strconv.Atoi(c.QueryParam("year"))
	if err != nil || year <= 0 {
		return 0, fmt.Errorf("invalid year")
	}
	return year, nil
}

func (h *statsHandleApi) parseYearMonth(c echo.Context) (int, int, error) {
	year, err := h.parseYear(c)
	if err != nil {
		return 0, 0, err
	}
	month, err := strconv.Atoi(c.QueryParam("month"))
	if err != nil || month <= 0 || month > 12 {
		return 0, 0, fmt.Errorf("invalid month")
	}
	return year, month, nil
}

func (h *statsHandleApi) parseID(c echo.Context, key string) (int, error) {
	id, err := strconv.Atoi(c.QueryParam(key))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return id, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Order stats
// ─────────────────────────────────────────────────────────────────────────────

// Monthly order total revenue
// @Summary Monthly order total revenue
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseOrderMonthlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/monthly-total-revenue [get]
func (h *statsHandleApi) FindOrderMonthlyTotalRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-monthly-total-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-total-revenue:%d:%d", year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Order.FindMonthlyTotalRevenue(ctx, &orderpb.FindYearMonthTotalRevenue{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderMonthlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly order total revenue
// @Summary Yearly order total revenue
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseOrderYearlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/yearly-total-revenue [get]
func (h *statsHandleApi) FindOrderYearlyTotalRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-yearly-total-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-total-revenue:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Order.FindYearlyTotalRevenue(ctx, &orderpb.FindYearTotalRevenue{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderYearlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly order revenue
// @Summary Monthly order revenue
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseOrderMonthly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/monthly-revenue [get]
func (h *statsHandleApi) FindOrderMonthlyRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-monthly-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-revenue:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Order.FindMonthlyRevenue(ctx, &orderpb.FindYearOrder{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderMonthly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly order revenue
// @Summary Yearly order revenue
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseOrderYearly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/yearly-revenue [get]
func (h *statsHandleApi) FindOrderYearlyRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-yearly-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-revenue:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Order.FindYearlyRevenue(ctx, &orderpb.FindYearOrder{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderYearly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly order total revenue by merchant
// @Summary Monthly order total revenue by merchant
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseOrderMonthlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/monthly-total-revenue [get]
func (h *statsHandleApi) FindOrderMonthlyTotalRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-monthly-total-revenue", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-monthly-total-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:by-merchant:monthly-total-revenue:%d:%d:%d", merchantID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.OrderByMerchant.FindMonthlyTotalRevenueByMerchant(ctx, &orderpb.FindYearMonthTotalRevenueByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderMonthlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly order total revenue by merchant
// @Summary Yearly order total revenue by merchant
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseOrderYearlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/yearly-total-revenue [get]
func (h *statsHandleApi) FindOrderYearlyTotalRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-yearly-total-revenue", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-yearly-total-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:by-merchant:yearly-total-revenue:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.OrderByMerchant.FindYearlyTotalRevenueByMerchant(ctx, &orderpb.FindYearTotalRevenueByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderYearlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly order revenue by merchant
// @Summary Monthly order revenue by merchant
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseOrderMonthly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/monthly-revenue [get]
func (h *statsHandleApi) FindOrderMonthlyRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-monthly-revenue", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-monthly-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:by-merchant:monthly-revenue:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.OrderByMerchant.FindMonthlyRevenueByMerchant(ctx, &orderpb.FindYearOrderByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderMonthly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly order revenue by merchant
// @Summary Yearly order revenue by merchant
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseOrderYearly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/yearly-revenue [get]
func (h *statsHandleApi) FindOrderYearlyRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-yearly-revenue", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-merchant-yearly-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:by-merchant:yearly-revenue:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.OrderByMerchant.FindYearlyRevenueByMerchant(ctx, &orderpb.FindYearOrderByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderYearly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly order total revenue by order
// @Summary Monthly order total revenue by order
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param order_id query int true "Order ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseOrderMonthlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-id/monthly-total-revenue [get]
func (h *statsHandleApi) FindOrderMonthlyTotalRevenueById(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-id-monthly-total-revenue", nil)(c)
	}
	orderID, err := h.parseID(c, "order_id")
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-id-monthly-total-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:by-id:monthly-total-revenue:%d:%d:%d", orderID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.OrderById.FindMonthlyTotalRevenueById(ctx, &orderpb.FindYearMonthTotalRevenueById{Year: int32(year), Month: int32(month), OrderId: int32(orderID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderMonthlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly order total revenue by order
// @Summary Yearly order total revenue by order
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param order_id query int true "Order ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseOrderYearlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-id/yearly-total-revenue [get]
func (h *statsHandleApi) FindOrderYearlyTotalRevenueById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-id-yearly-total-revenue", nil)(c)
	}
	orderID, err := h.parseID(c, "order_id")
	if err != nil {
		return h.apiHandler.Handle("get-order-stats-by-id-yearly-total-revenue", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:order-stats:by-id:yearly-total-revenue:%d:%d", orderID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.OrderById.FindYearlyTotalRevenueById(ctx, &orderpb.FindYearTotalRevenueById{Year: int32(year), OrderId: int32(orderID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseOrderYearlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// Cashier stats
// ─────────────────────────────────────────────────────────────────────────────

// Monthly cashier total sales
// @Summary Monthly cashier total sales
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseCashierMonthlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/monthly-total-sales [get]
func (h *statsHandleApi) FindCashierMonthlyTotalSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-monthly-total-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:monthly-total-sales:%d:%d", year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Cashier.FindMonthlyTotalSales(ctx, &cashierpb.FindYearMonthTotalSales{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierMonthlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly cashier total sales
// @Summary Yearly cashier total sales
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierYearlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/yearly-total-sales [get]
func (h *statsHandleApi) FindCashierYearlyTotalSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-yearly-total-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:yearly-total-sales:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Cashier.FindYearlyTotalSales(ctx, &cashierpb.FindYearTotalSales{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierYearlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly sales per cashier
// @Summary Monthly sales per cashier
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierMonthSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/month-sales [get]
func (h *statsHandleApi) FindCashierMonthSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-month-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:month-sales:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Cashier.FindMonthSales(ctx, &cashierpb.FindYearCashier{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierMonthSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly sales per cashier
// @Summary Yearly sales per cashier
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierYearSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/year-sales [get]
func (h *statsHandleApi) FindCashierYearSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-year-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:year-sales:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Cashier.FindYearSales(ctx, &cashierpb.FindYearCashier{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierYearSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly cashier total sales by merchant
// @Summary Monthly cashier total sales by merchant
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseCashierMonthlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/monthly-total-sales [get]
func (h *statsHandleApi) FindCashierMonthlyTotalSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-monthly-total-sales", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-monthly-total-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-merchant:monthly-total-sales:%d:%d:%d", merchantID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierByMerchant.FindMonthlyTotalSalesByMerchant(ctx, &cashierpb.FindYearMonthTotalSalesByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierMonthlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly cashier total sales by merchant
// @Summary Yearly cashier total sales by merchant
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierYearlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/yearly-total-sales [get]
func (h *statsHandleApi) FindCashierYearlyTotalSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-yearly-total-sales", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-yearly-total-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-merchant:yearly-total-sales:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierByMerchant.FindYearlyTotalSalesByMerchant(ctx, &cashierpb.FindYearTotalSalesByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierYearlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly sales per cashier by merchant
// @Summary Monthly sales per cashier by merchant
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierMonthSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/month-sales [get]
func (h *statsHandleApi) FindCashierMonthSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-month-sales", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-month-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-merchant:month-sales:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierByMerchant.FindMonthSalesByMerchant(ctx, &cashierpb.FindYearCashierByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierMonthSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly sales per cashier by merchant
// @Summary Yearly sales per cashier by merchant
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierYearSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/year-sales [get]
func (h *statsHandleApi) FindCashierYearSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-year-sales", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-merchant-year-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-merchant:year-sales:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierByMerchant.FindYearSalesByMerchant(ctx, &cashierpb.FindYearCashierByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierYearSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly cashier total sales by cashier
// @Summary Monthly cashier total sales by cashier
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseCashierMonthlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/monthly-total-sales [get]
func (h *statsHandleApi) FindCashierMonthlyTotalSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-monthly-total-sales", nil)(c)
	}
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-monthly-total-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-id:monthly-total-sales:%d:%d:%d", cashierID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierById.FindMonthlyTotalSalesById(ctx, &cashierpb.FindYearMonthTotalSalesById{Year: int32(year), Month: int32(month), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierMonthlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly cashier total sales by cashier
// @Summary Yearly cashier total sales by cashier
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierYearlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/yearly-total-sales [get]
func (h *statsHandleApi) FindCashierYearlyTotalSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-yearly-total-sales", nil)(c)
	}
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-yearly-total-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-id:yearly-total-sales:%d:%d", cashierID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierById.FindYearlyTotalSalesById(ctx, &cashierpb.FindYearTotalSalesById{Year: int32(year), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierYearlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly sales per cashier by cashier
// @Summary Monthly sales per cashier by cashier
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierMonthSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/month-sales [get]
func (h *statsHandleApi) FindCashierMonthSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-month-sales", nil)(c)
	}
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-month-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-id:month-sales:%d:%d", cashierID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierById.FindMonthSalesById(ctx, &cashierpb.FindYearCashierById{Year: int32(year), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierMonthSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly sales per cashier by cashier
// @Summary Yearly sales per cashier by cashier
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCashierYearSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/year-sales [get]
func (h *statsHandleApi) FindCashierYearSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-year-sales", nil)(c)
	}
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return h.apiHandler.Handle("get-cashier-stats-by-id-year-sales", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:cashier-stats:by-id:year-sales:%d:%d", cashierID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CashierById.FindYearSalesById(ctx, &cashierpb.FindYearCashierById{Year: int32(year), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCashierYearSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// Category stats
// ─────────────────────────────────────────────────────────────────────────────

// Monthly category total prices
// @Summary Monthly category total prices
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/monthly-total-prices [get]
func (h *statsHandleApi) FindCategoryMonthlyTotalPrices(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-monthly-total-prices", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:monthly-total-prices:%d:%d", year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Category.FindMonthlyTotalPrices(ctx, &categorypb.FindYearMonthTotalPrices{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly category total prices
// @Summary Yearly category total prices
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/yearly-total-prices [get]
func (h *statsHandleApi) FindCategoryYearlyTotalPrices(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-yearly-total-prices", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:yearly-total-prices:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Category.FindYearlyTotalPrices(ctx, &categorypb.FindYearTotalPrices{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly prices per category
// @Summary Monthly prices per category
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/month-price [get]
func (h *statsHandleApi) FindCategoryMonthPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-month-price", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:month-price:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Category.FindMonthPrice(ctx, &categorypb.FindYearCategory{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryMonthPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly prices per category
// @Summary Yearly prices per category
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/year-price [get]
func (h *statsHandleApi) FindCategoryYearPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-year-price", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:year-price:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.Category.FindYearPrice(ctx, &categorypb.FindYearCategory{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryYearPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly category total prices by merchant
// @Summary Monthly category total prices by merchant
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/monthly-total-prices [get]
func (h *statsHandleApi) FindCategoryMonthlyTotalPricesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-monthly-total-prices", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-monthly-total-prices", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-merchant:monthly-total-prices:%d:%d:%d", merchantID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryByMerchant.FindMonthlyTotalPricesByMerchant(ctx, &categorypb.FindYearMonthTotalPriceByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly category total prices by merchant
// @Summary Yearly category total prices by merchant
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/yearly-total-prices [get]
func (h *statsHandleApi) FindCategoryYearlyTotalPricesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-yearly-total-prices", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-yearly-total-prices", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-merchant:yearly-total-prices:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryByMerchant.FindYearlyTotalPricesByMerchant(ctx, &categorypb.FindYearTotalPriceByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly prices per category by merchant
// @Summary Monthly prices per category by merchant
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/month-price [get]
func (h *statsHandleApi) FindCategoryMonthPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-month-price", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-month-price", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-merchant:month-price:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryByMerchant.FindMonthPriceByMerchant(ctx, &categorypb.FindYearCategoryByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryMonthPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly prices per category by merchant
// @Summary Yearly prices per category by merchant
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/year-price [get]
func (h *statsHandleApi) FindCategoryYearPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-year-price", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-merchant-year-price", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-merchant:year-price:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryByMerchant.FindYearPriceByMerchant(ctx, &categorypb.FindYearCategoryByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryYearPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly category total prices by category
// @Summary Monthly category total prices by category
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/monthly-total-prices [get]
func (h *statsHandleApi) FindCategoryMonthlyTotalPricesById(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-monthly-total-prices", nil)(c)
	}
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-monthly-total-prices", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-id:monthly-total-prices:%d:%d:%d", categoryID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryById.FindMonthlyTotalPricesById(ctx, &categorypb.FindYearMonthTotalPriceById{Year: int32(year), Month: int32(month), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly category total prices by category
// @Summary Yearly category total prices by category
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/yearly-total-prices [get]
func (h *statsHandleApi) FindCategoryYearlyTotalPricesById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-yearly-total-prices", nil)(c)
	}
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-yearly-total-prices", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-id:yearly-total-prices:%d:%d", categoryID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryById.FindYearlyTotalPricesById(ctx, &categorypb.FindYearTotalPriceById{Year: int32(year), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly prices per category by category
// @Summary Monthly prices per category by category
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/month-price [get]
func (h *statsHandleApi) FindCategoryMonthPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-month-price", nil)(c)
	}
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-month-price", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-id:month-price:%d:%d", categoryID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryById.FindMonthPriceById(ctx, &categorypb.FindYearCategoryById{Year: int32(year), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryMonthPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly prices per category by category
// @Summary Yearly prices per category by category
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/year-price [get]
func (h *statsHandleApi) FindCategoryYearPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-year-price", nil)(c)
	}
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return h.apiHandler.Handle("get-category-stats-by-id-year-price", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:category-stats:by-id:year-price:%d:%d", categoryID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.CategoryById.FindYearPriceById(ctx, &categorypb.FindYearCategoryById{Year: int32(year), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseCategoryYearPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// Transaction stats
// ─────────────────────────────────────────────────────────────────────────────

// Monthly successful transactions
// @Summary Monthly successful transactions
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-status-success [get]
func (h *statsHandleApi) FindTransactionMonthStatusSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-month-status-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-status-success:%d:%d", year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindMonthStatusSuccess(ctx, &transactionpb.FindMonthlyTransactionStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly successful transactions
// @Summary Yearly successful transactions
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-status-success [get]
func (h *statsHandleApi) FindTransactionYearStatusSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-year-status-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-status-success:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindYearStatusSuccess(ctx, &transactionpb.FindYearlyTransactionStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly failed transactions
// @Summary Monthly failed transactions
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-status-failed [get]
func (h *statsHandleApi) FindTransactionMonthStatusFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-month-status-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-status-failed:%d:%d", year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindMonthStatusFailed(ctx, &transactionpb.FindMonthlyTransactionStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly failed transactions
// @Summary Yearly failed transactions
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-status-failed [get]
func (h *statsHandleApi) FindTransactionYearStatusFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-year-status-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-status-failed:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindYearStatusFailed(ctx, &transactionpb.FindYearlyTransactionStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly successful transactions per payment method
// @Summary Monthly successful transactions per payment method
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-method-success [get]
func (h *statsHandleApi) FindTransactionMonthMethodSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-month-method-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-method-success:%d:%d", year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindMonthMethodSuccess(ctx, &transactionpb.MonthTransactionMethod{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly successful transactions per payment method
// @Summary Yearly successful transactions per payment method
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-method-success [get]
func (h *statsHandleApi) FindTransactionYearMethodSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-year-method-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-method-success:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindYearMethodSuccess(ctx, &transactionpb.YearTransactionMethod{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly failed transactions per payment method
// @Summary Monthly failed transactions per payment method
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-method-failed [get]
func (h *statsHandleApi) FindTransactionMonthMethodFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-month-method-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-method-failed:%d:%d", year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindMonthMethodFailed(ctx, &transactionpb.MonthTransactionMethod{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly failed transactions per payment method
// @Summary Yearly failed transactions per payment method
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-method-failed [get]
func (h *statsHandleApi) FindTransactionYearMethodFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-year-method-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-method-failed:%d", year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindYearMethodFailed(ctx, &transactionpb.YearTransactionMethod{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly successful transactions by merchant
// @Summary Monthly successful transactions by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-status-success [get]
func (h *statsHandleApi) FindTransactionMonthStatusSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-status-success", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-status-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:month-status-success:%d:%d:%d", merchantID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindMonthStatusSuccessByMerchant(ctx, &transactionpb.FindMonthlyTransactionStatusByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly successful transactions by merchant
// @Summary Yearly successful transactions by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-status-success [get]
func (h *statsHandleApi) FindTransactionYearStatusSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-status-success", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-status-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:year-status-success:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindYearStatusSuccessByMerchant(ctx, &transactionpb.FindYearlyTransactionStatusByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly failed transactions by merchant
// @Summary Monthly failed transactions by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-status-failed [get]
func (h *statsHandleApi) FindTransactionMonthStatusFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-status-failed", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-status-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:month-status-failed:%d:%d:%d", merchantID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindMonthStatusFailedByMerchant(ctx, &transactionpb.FindMonthlyTransactionStatusByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly failed transactions by merchant
// @Summary Yearly failed transactions by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-status-failed [get]
func (h *statsHandleApi) FindTransactionYearStatusFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-status-failed", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-status-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:year-status-failed:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionStatus.FindYearStatusFailedByMerchant(ctx, &transactionpb.FindYearlyTransactionStatusByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly successful payment methods by merchant
// @Summary Monthly successful payment methods by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-method-success [get]
func (h *statsHandleApi) FindTransactionMonthMethodSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-method-success", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-method-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:month-method-success:%d:%d:%d", merchantID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindMonthMethodByMerchantSuccess(ctx, &transactionpb.MonthTransactionMethodByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly successful payment methods by merchant
// @Summary Yearly successful payment methods by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-method-success [get]
func (h *statsHandleApi) FindTransactionYearMethodSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-method-success", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-method-success", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:year-method-success:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindYearMethodByMerchantSuccess(ctx, &transactionpb.YearTransactionMethodByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Monthly failed payment methods by merchant
// @Summary Monthly failed payment methods by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Param month query int true "Month (1-12, 0 for the whole year)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-method-failed [get]
func (h *statsHandleApi) FindTransactionMonthMethodFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-method-failed", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-month-method-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:month-method-failed:%d:%d:%d", merchantID, year, month)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindMonthMethodByMerchantFailed(ctx, &transactionpb.MonthTransactionMethodByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}

// Yearly failed payment methods by merchant
// @Summary Yearly failed payment methods by merchant
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-method-failed [get]
func (h *statsHandleApi) FindTransactionYearMethodFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-method-failed", nil)(c)
	}
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return h.apiHandler.Handle("get-transaction-stats-by-merchant-year-method-failed", nil)(c)
	}
	cacheKey := fmt.Sprintf("apigw:transaction-stats:by-merchant:year-method-failed:%d:%d", merchantID, year)
	if got, found := gateway_cache.Get[any](ctx, h.cache, cacheKey); found && got != nil {
		return c.JSON(http.StatusOK, got)
	}
	res, err := h.clients.TransactionMethod.FindYearMethodByMerchantFailed(ctx, &transactionpb.YearTransactionMethodByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}
	resp := h.mapping.ToApiResponseTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, resp, 5*time.Minute)
	return c.JSON(http.StatusOK, resp)
}
