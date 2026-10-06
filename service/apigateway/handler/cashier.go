package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type cashierHandleApi struct {
	clientQuery   cashierpb.CashierServiceClient
	clientCommand cashierpb.CashierServiceClient
	logger        logger.LoggerInterface
	mapping       response_api.CashierResponseMapper
	apiHandler    errors.ApiHandler
	cache         *gateway_cache.GatewayCache
}

func NewHandlerCashier(
	router *echo.Echo,
	clientQuery cashierpb.CashierServiceClient,
	clientCommand cashierpb.CashierServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.CashierResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *cashierHandleApi {
	cashierHandler := &cashierHandleApi{
		clientQuery:   clientQuery,
		clientCommand: clientCommand,
		logger:        logger,
		mapping:       mapping,
		apiHandler:    apiHandler,
		cache:         cache,
	}

	routerCashier := router.Group("/api/cashier")

	routerCashier.GET("", apiHandler.Handle("get-cashier-findallcashier", cashierHandler.FindAllCashier))
	routerCashier.GET("/:id", apiHandler.Handle("get-cashier-findbyid", cashierHandler.FindById))
	routerCashier.GET("/active", apiHandler.Handle("get-cashier-findbyactive", cashierHandler.FindByActive))
	routerCashier.GET("/trashed", apiHandler.Handle("get-cashier-findbytrashed", cashierHandler.FindByTrashed))

	routerCashier.POST("/create", apiHandler.Handle("post-cashier-createcashier", cashierHandler.CreateCashier))
	routerCashier.POST("/update/:id", apiHandler.Handle("post-cashier-updatecashier", cashierHandler.UpdateCashier))

	routerCashier.POST("/trashed/:id", apiHandler.Handle("post-cashier-trashedcashier", cashierHandler.TrashedCashier))
	routerCashier.POST("/restore/:id", apiHandler.Handle("post-cashier-restorecashier", cashierHandler.RestoreCashier))
	routerCashier.DELETE("/permanent/:id", apiHandler.Handle("delete-cashier-deletecashierpermanent", cashierHandler.DeleteCashierPermanent))

	routerCashier.POST("/restore/all", apiHandler.Handle("post-cashier-restoreallcashier", cashierHandler.RestoreAllCashier))
	routerCashier.POST("/permanent/all", apiHandler.Handle("post-cashier-deleteallcashierpermanent", cashierHandler.DeleteAllCashierPermanent))

	return cashierHandler
}

// List all cashier (paginated)
// @Summary List all cashier (paginated)
// @Tags Cashier
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCashier
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier [get]
// @Security BearerAuth
func (h *cashierHandleApi) FindAllCashier(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("cashier:findallcashier:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCashier](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &cashierpb.FindAllCashierRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCashier(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get cashier by ID
// @Summary Get cashier by ID
// @Tags Cashier
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCashier
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/:id [get]
// @Security BearerAuth
func (h *cashierHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid cashier ID")
	}

	cacheKey := fmt.Sprintf("cashier:findbyid:id_%d", id)
	if cached, found := gateway_cache.Get[response.ApiResponseCashier](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &cashierpb.FindByIdCashierRequest{Id: int32(id)}
	res, err := h.clientQuery.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashier(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active cashier
// @Summary List active cashier
// @Tags Cashier
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCashierDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/active [get]
// @Security BearerAuth
func (h *cashierHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("cashier:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCashierDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &cashierpb.FindAllCashierRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCashierDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed cashier
// @Summary List trashed cashier
// @Tags Cashier
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCashierDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/trashed [get]
// @Security BearerAuth
func (h *cashierHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("cashier:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCashierDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &cashierpb.FindAllCashierRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCashierDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create cashier
// @Summary Create cashier
// @Tags Cashier
// @Accept json
// @Produce json
// @Param request body requests.CreateCashierRequest true "Request body"
// @Success 200 {object} response.ApiResponseCashier
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/create [post]
// @Security BearerAuth
func (h *cashierHandleApi) CreateCashier(c echo.Context) error {
	ctx := c.Request().Context()
	var body requests.CreateCashierRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &cashierpb.CreateCashierRequest{
		MerchantId: int32(body.MerchantID),
		UserId:     int32(body.UserID),
		Name:       body.Name,
	}

	res, err := h.clientCommand.CreateCashier(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashier(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "cashier:*")
	return c.JSON(http.StatusCreated, so)
}

// Update cashier
// @Summary Update cashier
// @Tags Cashier
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateCashierRequest true "Request body"
// @Success 200 {object} response.ApiResponseCashier
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/update/:id [post]
// @Security BearerAuth
func (h *cashierHandleApi) UpdateCashier(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	idStr, err := strconv.Atoi(id)
	if err != nil {
		return errors.NewBadRequestError("invalid cashier ID")
	}

	var body requests.UpdateCashierRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &cashierpb.UpdateCashierRequest{
		CashierId: int32(idStr),
		Name:      body.Name,
	}

	res, err := h.clientCommand.UpdateCashier(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashier(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "cashier:*")
	return c.JSON(http.StatusOK, so)
}

// Trash cashier
// @Summary Trash cashier
// @Tags Cashier
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCashierDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/trashed/:id [post]
// @Security BearerAuth
func (h *cashierHandleApi) TrashedCashier(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid cashier ID")
	}

	req := &cashierpb.FindByIdCashierRequest{Id: int32(id)}
	res, err := h.clientCommand.TrashedCashier(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "cashier:*")
	return c.JSON(http.StatusOK, so)
}

// Restore cashier
// @Summary Restore cashier
// @Tags Cashier
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCashierDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/restore/:id [post]
// @Security BearerAuth
func (h *cashierHandleApi) RestoreCashier(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid cashier ID")
	}

	req := &cashierpb.FindByIdCashierRequest{Id: int32(id)}
	res, err := h.clientCommand.RestoreCashier(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "cashier:*")
	return c.JSON(http.StatusOK, so)
}

// Delete cashier permanently
// @Summary Delete cashier permanently
// @Tags Cashier
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCashierDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/permanent/:id [delete]
// @Security BearerAuth
func (h *cashierHandleApi) DeleteCashierPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid cashier ID")
	}

	req := &cashierpb.FindByIdCashierRequest{Id: int32(id)}
	res, err := h.clientCommand.DeleteCashierPermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "cashier:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed cashier
// @Summary Restore all trashed cashier
// @Tags Cashier
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCashierAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/restore/all [post]
// @Security BearerAuth
func (h *cashierHandleApi) RestoreAllCashier(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllCashier(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "cashier:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed cashier permanently
// @Summary Delete all trashed cashier permanently
// @Tags Cashier
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCashierAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier/permanent/all [post]
// @Security BearerAuth
func (h *cashierHandleApi) DeleteAllCashierPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllCashierPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "cashier:*")
	return c.JSON(http.StatusOK, so)
}
