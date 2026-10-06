package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	orderitempb "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
)

type orderItemHandleApi struct {
	clientQuery orderitempb.OrderItemServiceClient
	logger      logger.LoggerInterface
	mapping     response_api.OrderItemResponseMapper
	apiHandler  errors.ApiHandler
	cache       *gateway_cache.GatewayCache
}

func NewHandlerOrderItem(
	router *echo.Echo,
	clientQuery orderitempb.OrderItemServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.OrderItemResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *orderItemHandleApi {
	orderItemHandler := &orderItemHandleApi{
		clientQuery: clientQuery,
		logger:      logger,
		mapping:     mapping,
		apiHandler:  apiHandler,
		cache:       cache,
	}

	routerOrderItem := router.Group("/api/order-item")

	routerOrderItem.GET("", apiHandler.Handle("get-orderitem-findallorderitems", orderItemHandler.FindAllOrderItems))
	routerOrderItem.GET("/:order_id", apiHandler.Handle("get-orderitem-findorderitembyorder", orderItemHandler.FindOrderItemByOrder))
	routerOrderItem.GET("/active", apiHandler.Handle("get-orderitem-findbyactive", orderItemHandler.FindByActive))
	routerOrderItem.GET("/trashed", apiHandler.Handle("get-orderitem-findbytrashed", orderItemHandler.FindByTrashed))

	return orderItemHandler
}

// List all order item (paginated)
// List all order item (paginated)
// List all order item (paginated)
// @Summary List all order item (paginated)
// @Tags Order Item
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationOrderItem
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-item [get]
// @Security BearerAuth
func (h *orderItemHandleApi) FindAllOrderItems(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("order_item:findallorderitems:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationOrderItem](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &orderitempb.FindAllOrderItemRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationOrderItem(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active order item
// List active order item
// List active order item
// @Summary List active order item
// @Tags Order Item
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationOrderItemDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-item/active [get]
// @Security BearerAuth
func (h *orderItemHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("order_item:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationOrderItemDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &orderitempb.FindAllOrderItemRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationOrderItemDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed order item
// List trashed order item
// List trashed order item
// @Summary List trashed order item
// @Tags Order Item
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationOrderItemDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-item/trashed [get]
// @Security BearerAuth
func (h *orderItemHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("order_item:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationOrderItemDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &orderitempb.FindAllOrderItemRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationOrderItemDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// FindOrderItemByOrder
// FindOrderItemByOrder
// FindOrderItemByOrder
// @Summary FindOrderItemByOrder
// @Tags Order Item
// @Accept json
// @Produce json
// @Param order_id path int true "order id"
// @Success 200 {object} response.ApiResponseOrderItem
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-item/:order_id [get]
// @Security BearerAuth
func (h *orderItemHandleApi) FindOrderItemByOrder(c echo.Context) error {
	ctx := c.Request().Context()
	orderID, err := strconv.Atoi(c.Param("order_id"))
	if err != nil {
		return errors.NewBadRequestError("invalid order ID")
	}

	cacheKey := fmt.Sprintf("order_item:findorderitembyorder:order_id_%d", orderID)
	if cached, found := gateway_cache.Get[response.ApiResponsesOrderItem](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &orderitempb.FindByIdOrderItemRequest{
		Id: int32(orderID),
	}

	res, err := h.clientQuery.FindOrderItemByOrder(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesOrderItem(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}
