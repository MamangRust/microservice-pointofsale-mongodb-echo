package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"

	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type categoryHandleApi struct {
	clientQuery   categorypb.CategoryServiceClient
	clientCommand categorypb.CategoryServiceClient
	logger        logger.LoggerInterface
	mapping       response_api.CategoryResponseMapper
	apiHandler    errors.ApiHandler
	cache         *gateway_cache.GatewayCache
}

func NewHandlerCategory(
	router *echo.Echo,
	clientQuery categorypb.CategoryServiceClient,
	clientCommand categorypb.CategoryServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.CategoryResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *categoryHandleApi {
	categoryHandler := &categoryHandleApi{
		clientQuery:   clientQuery,
		clientCommand: clientCommand,
		logger:        logger,
		mapping:       mapping,
		apiHandler:    apiHandler,
		cache:         cache,
	}

	routercategory := router.Group("/api/category")

	routercategory.GET("", apiHandler.Handle("get-category-findallcategory", categoryHandler.FindAllCategory))
	routercategory.GET("/:id", apiHandler.Handle("get-category-findbyid", categoryHandler.FindById))
	routercategory.GET("/active", apiHandler.Handle("get-category-findbyactive", categoryHandler.FindByActive))
	routercategory.GET("/trashed", apiHandler.Handle("get-category-findbytrashed", categoryHandler.FindByTrashed))

	routercategory.POST("/create", apiHandler.Handle("post-category-create", categoryHandler.Create))
	routercategory.POST("/update/:id", apiHandler.Handle("post-category-update", categoryHandler.Update))

	routercategory.POST("/trashed/:id", apiHandler.Handle("post-category-trashedcategory", categoryHandler.TrashedCategory))
	routercategory.POST("/restore/:id", apiHandler.Handle("post-category-restorecategory", categoryHandler.RestoreCategory))
	routercategory.DELETE("/permanent/:id", apiHandler.Handle("delete-category-deletecategorypermanent", categoryHandler.DeleteCategoryPermanent))

	routercategory.POST("/restore/all", apiHandler.Handle("post-category-restoreallcategory", categoryHandler.RestoreAllCategory))
	routercategory.POST("/permanent/all", apiHandler.Handle("post-category-deleteallcategorypermanent", categoryHandler.DeleteAllCategoryPermanent))

	return categoryHandler
}

// List all category (paginated)
// List all category (paginated)
// List all category (paginated)
// @Summary List all category (paginated)
// @Tags Category
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindAllCategory(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("category:findallcategory:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCategory](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &categorypb.FindAllCategoryRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCategory(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category by ID
// Get category by ID
// Get category by ID
// @Summary Get category by ID
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/:id [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	cacheKey := fmt.Sprintf("category:findbyid:id_%d", id)
	if cached, found := gateway_cache.Get[response.ApiResponseCategory](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &categorypb.FindByIdCategoryRequest{
		Id: int32(id),
	}

	res, err := h.clientQuery.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategory(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active category
// List active category
// List active category
// @Summary List active category
// @Tags Category
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/active [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("category:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCategoryDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &categorypb.FindAllCategoryRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCategoryDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed category
// List trashed category
// List trashed category
// @Summary List trashed category
// @Tags Category
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/trashed [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("category:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCategoryDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &categorypb.FindAllCategoryRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCategoryDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create category
// Create category
// Create category
// @Summary Create category
// @Tags Category
// @Accept json
// @Produce json
// @Param request body requests.CreateCategoryRequest true "Request body"
// @Success 200 {object} response.ApiResponseCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/create [post]
// @Security BearerAuth
func (h *categoryHandleApi) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var body requests.CreateCategoryRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &categorypb.CreateCategoryRequest{
		Name:        body.Name,
		Description: body.Description,
	}

	res, err := h.clientCommand.Create(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategory(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusCreated, so)
}

// Update category
// Update category
// Update category
// @Summary Update category
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateCategoryRequest true "Request body"
// @Success 200 {object} response.ApiResponseCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/update/:id [post]
// @Security BearerAuth
func (h *categoryHandleApi) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	var body requests.UpdateCategoryRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &categorypb.UpdateCategoryRequest{
		CategoryId:  int32(id),
		Name:        body.Name,
		Description: body.Description,
	}

	res, err := h.clientCommand.Update(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategory(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Trash category
// Trash category
// Trash category
// @Summary Trash category
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/trashed/:id [post]
// @Security BearerAuth
func (h *categoryHandleApi) TrashedCategory(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	req := &categorypb.FindByIdCategoryRequest{Id: int32(id)}
	res, err := h.clientCommand.TrashedCategory(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Restore category
// Restore category
// Restore category
// @Summary Restore category
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/restore/:id [post]
// @Security BearerAuth
func (h *categoryHandleApi) RestoreCategory(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	req := &categorypb.FindByIdCategoryRequest{Id: int32(id)}
	res, err := h.clientCommand.RestoreCategory(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Delete category permanently
// Delete category permanently
// Delete category permanently
// @Summary Delete category permanently
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategoryDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/permanent/:id [delete]
// @Security BearerAuth
func (h *categoryHandleApi) DeleteCategoryPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	req := &categorypb.FindByIdCategoryRequest{Id: int32(id)}
	res, err := h.clientCommand.DeleteCategoryPermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed category
// Restore all trashed category
// Restore all trashed category
// @Summary Restore all trashed category
// @Tags Category
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCategoryAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/restore/all [post]
// @Security BearerAuth
func (h *categoryHandleApi) RestoreAllCategory(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllCategory(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed category permanently
// Delete all trashed category permanently
// Delete all trashed category permanently
// @Summary Delete all trashed category permanently
// @Tags Category
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCategoryAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/permanent/all [post]
// @Security BearerAuth
func (h *categoryHandleApi) DeleteAllCategoryPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllCategoryPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}
