package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	transactionpb "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type transactionHandleApi struct {
	clientQuery   transactionpb.TransactionServiceClient
	clientCommand transactionpb.TransactionServiceClient
	logger        logger.LoggerInterface
	mapping       response_api.TransactionResponseMapper
	apiHandler    errors.ApiHandler
	cache         *gateway_cache.GatewayCache
}

func NewHandlerTransaction(
	router *echo.Echo,
	clientQuery transactionpb.TransactionServiceClient,
	clientCommand transactionpb.TransactionServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.TransactionResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *transactionHandleApi {
	transactionHandle := &transactionHandleApi{
		clientQuery:   clientQuery,
		clientCommand: clientCommand,
		logger:        logger,
		mapping:       mapping,
		apiHandler:    apiHandler,
		cache:         cache,
	}

	routerTransaction := router.Group("/api/transaction")

	routerTransaction.GET("", apiHandler.Handle("get-transaction-findalltransaction", transactionHandle.FindAllTransaction))
	routerTransaction.GET("/:id", apiHandler.Handle("get-transaction-findbyid", transactionHandle.FindById))
	routerTransaction.GET("/merchant/:merchant_id", apiHandler.Handle("get-transaction-findbymerchant", transactionHandle.FindByMerchant))
	routerTransaction.GET("/active", apiHandler.Handle("get-transaction-findbyactive", transactionHandle.FindByActive))
	routerTransaction.GET("/trashed", apiHandler.Handle("get-transaction-findbytrashed", transactionHandle.FindByTrashed))

	routerTransaction.POST("/create", apiHandler.Handle("post-transaction-create", transactionHandle.Create))
	routerTransaction.POST("/update/:id", apiHandler.Handle("post-transaction-update", transactionHandle.Update))

	routerTransaction.POST("/trashed/:id", apiHandler.Handle("post-transaction-trashedtransaction", transactionHandle.TrashedTransaction))
	routerTransaction.POST("/restore/:id", apiHandler.Handle("post-transaction-restoretransaction", transactionHandle.RestoreTransaction))
	routerTransaction.DELETE("/permanent/:id", apiHandler.Handle("delete-transaction-deletetransactionpermanent", transactionHandle.DeleteTransactionPermanent))

	routerTransaction.POST("/restore/all", apiHandler.Handle("post-transaction-restorealltransaction", transactionHandle.RestoreAllTransaction))
	routerTransaction.POST("/permanent/all", apiHandler.Handle("delete-transaction-deletealltransactionpermanent", transactionHandle.DeleteAllTransactionPermanent))

	return transactionHandle
}

// List all transaction (paginated)
// List all transaction (paginated)
// List all transaction (paginated)
// @Summary List all transaction (paginated)
// @Tags Transaction
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationTransaction
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction [get]
// @Security BearerAuth
func (h *transactionHandleApi) FindAllTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("transaction:findalltransaction:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationTransaction](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &transactionpb.FindAllTransactionRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationTransaction(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List transaction by merchant
// List transaction by merchant
// List transaction by merchant
// @Summary List transaction by merchant
// @Tags Transaction
// @Accept json
// @Produce json
// @Param merchant_id path int true "merchant id"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationTransaction
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/merchant/:merchant_id [get]
// @Security BearerAuth
func (h *transactionHandleApi) FindByMerchant(c echo.Context) error {
	merchantID, err := strconv.Atoi(c.Param("merchant_id"))
	if err != nil || merchantID <= 0 {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("transaction:findbymerchant:merchant_%d:page_%d:size_%d:search_%s", merchantID, page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationTransaction](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &transactionpb.FindAllTransactionMerchantRequest{
		MerchantId: int32(merchantID),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		Search:     search,
	}

	res, err := h.clientQuery.FindByMerchant(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationTransaction(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get transaction by ID
// Get transaction by ID
// Get transaction by ID
// @Summary Get transaction by ID
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseTransaction
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/:id [get]
// @Security BearerAuth
func (h *transactionHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid transaction ID")
	}

	cacheKey := fmt.Sprintf("transaction:findbyid:id_%d", id)
	if cached, found := gateway_cache.Get[response.ApiResponseTransaction](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &transactionpb.FindByIdTransactionRequest{
		Id: int32(id),
	}

	res, err := h.clientQuery.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransaction(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active transaction
// List active transaction
// List active transaction
// @Summary List active transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationTransactionDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/active [get]
// @Security BearerAuth
func (h *transactionHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("transaction:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationTransactionDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &transactionpb.FindAllTransactionRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationTransactionDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed transaction
// List trashed transaction
// List trashed transaction
// @Summary List trashed transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationTransactionDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/trashed [get]
// @Security BearerAuth
func (h *transactionHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("transaction:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationTransactionDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &transactionpb.FindAllTransactionRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationTransactionDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create transaction
// Create transaction
// Create transaction
// @Summary Create transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param request body requests.CreateTransactionRequest true "Request body"
// @Success 200 {object} response.ApiResponseTransaction
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/create [post]
// @Security BearerAuth
func (h *transactionHandleApi) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var body requests.CreateTransactionRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	grpcReq := &transactionpb.CreateTransactionRequest{
		OrderId:       int32(body.OrderID),
		CashierId:     int32(body.CashierID),
		PaymentMethod: body.PaymentMethod,
		Amount:        int32(body.Amount),
	}

	res, err := h.clientCommand.Create(ctx, grpcReq)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransaction(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "transaction:*")
	return c.JSON(http.StatusOK, so)
}

// Update transaction
// Update transaction
// Update transaction
// @Summary Update transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateTransactionRequest true "Request body"
// @Success 200 {object} response.ApiResponseTransaction
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/update/:id [post]
// @Security BearerAuth
func (h *transactionHandleApi) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid transaction ID")
	}

	var body requests.UpdateTransactionRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	grpcReq := &transactionpb.UpdateTransactionRequest{
		TransactionId: int32(id),
		OrderId:       int32(body.OrderID),
		CashierId:     int32(body.CashierID),
		PaymentMethod: body.PaymentMethod,
		Amount:        int32(body.Amount),
	}

	res, err := h.clientCommand.Update(ctx, grpcReq)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransaction(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "transaction:*")
	return c.JSON(http.StatusOK, so)
}

// Trash transaction
// Trash transaction
// Trash transaction
// @Summary Trash transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseTransactionDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/trashed/:id [post]
// @Security BearerAuth
func (h *transactionHandleApi) TrashedTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid transaction ID")
	}

	req := &transactionpb.FindByIdTransactionRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.TrashedTransaction(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "transaction:*")
	return c.JSON(http.StatusOK, so)
}

// Restore transaction
// Restore transaction
// Restore transaction
// @Summary Restore transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseTransactionDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/restore/:id [post]
// @Security BearerAuth
func (h *transactionHandleApi) RestoreTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid transaction ID")
	}

	req := &transactionpb.FindByIdTransactionRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.RestoreTransaction(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "transaction:*")
	return c.JSON(http.StatusOK, so)
}

// Delete transaction permanently
// Delete transaction permanently
// Delete transaction permanently
// @Summary Delete transaction permanently
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseTransactionDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/permanent/:id [delete]
// @Security BearerAuth
func (h *transactionHandleApi) DeleteTransactionPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid transaction ID")
	}

	req := &transactionpb.FindByIdTransactionRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.DeleteTransactionPermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "transaction:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed transaction
// Restore all trashed transaction
// Restore all trashed transaction
// @Summary Restore all trashed transaction
// @Tags Transaction
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseTransactionAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/restore/all [post]
// @Security BearerAuth
func (h *transactionHandleApi) RestoreAllTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllTransaction(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "transaction:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed transaction permanently
// Delete all trashed transaction permanently
// Delete all trashed transaction permanently
// @Summary Delete all trashed transaction permanently
// @Tags Transaction
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseTransactionAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction/permanent/all [post]
// @Security BearerAuth
func (h *transactionHandleApi) DeleteAllTransactionPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllTransactionPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "transaction:*")
	return c.JSON(http.StatusOK, so)
}
