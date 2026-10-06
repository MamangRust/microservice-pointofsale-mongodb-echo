package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	merchantpb "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantHandleApi struct {
	clientQuery   merchantpb.MerchantServiceClient
	clientCommand merchantpb.MerchantServiceClient
	logger        logger.LoggerInterface
	mapping       response_api.MerchantResponseMapper
	apiHandler    errors.ApiHandler
	cache         *gateway_cache.GatewayCache
}

func NewHandlerMerchant(
	router *echo.Echo,
	clientQuery merchantpb.MerchantServiceClient,
	clientCommand merchantpb.MerchantServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.MerchantResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *merchantHandleApi {
	merchantHandler := &merchantHandleApi{
		clientQuery:   clientQuery,
		clientCommand: clientCommand,
		logger:        logger,
		mapping:       mapping,
		apiHandler:    apiHandler,
		cache:         cache,
	}

	routerMerchant := router.Group("/api/merchant")

	routerMerchant.GET("", apiHandler.Handle("get-merchant-findallmerchant", merchantHandler.FindAllMerchant))
	routerMerchant.GET("/:id", apiHandler.Handle("get-merchant-findbyid", merchantHandler.FindById))
	routerMerchant.GET("/active", apiHandler.Handle("get-merchant-findbyactive", merchantHandler.FindByActive))
	routerMerchant.GET("/trashed", apiHandler.Handle("get-merchant-findbytrashed", merchantHandler.FindByTrashed))

	routerMerchant.POST("/create", apiHandler.Handle("post-merchant-create", merchantHandler.Create))
	routerMerchant.POST("/update/:id", apiHandler.Handle("post-merchant-update", merchantHandler.Update))
	routerMerchant.POST("/update-status/:id", apiHandler.Handle("post-merchant-updatestatus", merchantHandler.UpdateStatus))

	routerMerchant.POST("/trashed/:id", apiHandler.Handle("post-merchant-trashedmerchant", merchantHandler.TrashedMerchant))
	routerMerchant.POST("/restore/:id", apiHandler.Handle("post-merchant-restoremerchant", merchantHandler.RestoreMerchant))
	routerMerchant.DELETE("/permanent/:id", apiHandler.Handle("delete-merchant-deletemerchantpermanent", merchantHandler.DeleteMerchantPermanent))

	routerMerchant.POST("/restore/all", apiHandler.Handle("post-merchant-restoreallmerchant", merchantHandler.RestoreAllMerchant))
	routerMerchant.POST("/permanent/all", apiHandler.Handle("post-merchant-deleteallmerchantpermanent", merchantHandler.DeleteAllMerchantPermanent))

	return merchantHandler
}

// List all merchant (paginated)
// List all merchant (paginated)
// List all merchant (paginated)
// @Summary List all merchant (paginated)
// @Tags Merchant
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationMerchant
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant [get]
// @Security BearerAuth
func (h *merchantHandleApi) FindAllMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("merchant:findallmerchant:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationMerchant](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &merchantpb.FindAllMerchantRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationMerchant(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get merchant by ID
// Get merchant by ID
// Get merchant by ID
// @Summary Get merchant by ID
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseMerchant
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/:id [get]
// @Security BearerAuth
func (h *merchantHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	cacheKey := fmt.Sprintf("merchant:findbyid:id_%d", id)
	if cached, found := gateway_cache.Get[response.ApiResponseMerchant](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &merchantpb.FindByIdMerchantRequest{
		Id: int32(id),
	}

	res, err := h.clientQuery.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchant(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active merchant
// List active merchant
// List active merchant
// @Summary List active merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationMerchantDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/active [get]
// @Security BearerAuth
func (h *merchantHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("merchant:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationMerchantDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &merchantpb.FindAllMerchantRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationMerchantDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed merchant
// List trashed merchant
// List trashed merchant
// @Summary List trashed merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationMerchantDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/trashed [get]
// @Security BearerAuth
func (h *merchantHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("merchant:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationMerchantDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &merchantpb.FindAllMerchantRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationMerchantDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create merchant
// Create merchant
// Create merchant
// @Summary Create merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Param request body requests.CreateMerchantRequest true "Request body"
// @Success 200 {object} response.ApiResponseMerchant
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/create [post]
// @Security BearerAuth
func (h *merchantHandleApi) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var body requests.CreateMerchantRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &merchantpb.CreateMerchantRequest{
		UserId:       int32(body.UserID),
		Name:         body.Name,
		Description:  body.Description,
		Address:      body.Address,
		ContactEmail: body.ContactEmail,
		ContactPhone: body.ContactPhone,
		Status:       body.Status,
	}

	res, err := h.clientCommand.Create(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchant(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}

// Update merchant
// Update merchant
// Update merchant
// @Summary Update merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateMerchantRequest true "Request body"
// @Success 200 {object} response.ApiResponseMerchant
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/update/:id [post]
// @Security BearerAuth
func (h *merchantHandleApi) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	var body requests.UpdateMerchantRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &merchantpb.UpdateMerchantRequest{
		MerchantId:   int32(id),
		UserId:       int32(body.UserID),
		Name:         body.Name,
		Description:  body.Description,
		Address:      body.Address,
		ContactEmail: body.ContactEmail,
		ContactPhone: body.ContactPhone,
		Status:       body.Status,
	}

	res, err := h.clientCommand.Update(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchant(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}

// Update merchant status
// Update merchant status
// Update merchant status
// @Summary Update merchant status
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateMerchantStatusRequest true "Request body"
// @Success 200 {object} response.ApiResponseMerchant
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/update-status/:id [post]
// @Security BearerAuth
func (h *merchantHandleApi) UpdateStatus(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	var body requests.UpdateMerchantStatusRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &merchantpb.UpdateMerchantStatusRequest{
		MerchantId: int32(id),
		Status:     body.Status,
	}

	res, err := h.clientCommand.UpdateMerchantStatus(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchant(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}

// Trash merchant
// Trash merchant
// Trash merchant
// @Summary Trash merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseMerchantDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/trashed/:id [post]
// @Security BearerAuth
func (h *merchantHandleApi) TrashedMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	req := &merchantpb.FindByIdMerchantRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.TrashedMerchant(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchantDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}

// Restore merchant
// Restore merchant
// Restore merchant
// @Summary Restore merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseMerchantDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/restore/:id [post]
// @Security BearerAuth
func (h *merchantHandleApi) RestoreMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	req := &merchantpb.FindByIdMerchantRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.RestoreMerchant(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchant(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}

// Delete merchant permanently
// Delete merchant permanently
// Delete merchant permanently
// @Summary Delete merchant permanently
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseMerchantDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/permanent/:id [delete]
// @Security BearerAuth
func (h *merchantHandleApi) DeleteMerchantPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	req := &merchantpb.FindByIdMerchantRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.DeleteMerchantPermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchantDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed merchant
// Restore all trashed merchant
// Restore all trashed merchant
// @Summary Restore all trashed merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/restore/all [post]
// @Security BearerAuth
func (h *merchantHandleApi) RestoreAllMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllMerchant(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchantAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed merchant permanently
// Delete all trashed merchant permanently
// Delete all trashed merchant permanently
// @Summary Delete all trashed merchant permanently
// @Tags Merchant
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/merchant/permanent/all [post]
// @Security BearerAuth
func (h *merchantHandleApi) DeleteAllMerchantPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllMerchantPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseMerchantAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "merchant:*")
	return c.JSON(http.StatusOK, so)
}
