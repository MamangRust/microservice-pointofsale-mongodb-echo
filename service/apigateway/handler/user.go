package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	userpb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userHandleApi struct {
	clientQuery   userpb.UserServiceClient
	clientCommand userpb.UserServiceClient
	logger        logger.LoggerInterface
	mapping       response_api.UserResponseMapper
	apiHandler    errors.ApiHandler
	cache         *gateway_cache.GatewayCache
}

func NewHandlerUser(
	router *echo.Echo,
	clientQuery userpb.UserServiceClient,
	clientCommand userpb.UserServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.UserResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *userHandleApi {
	userHandler := &userHandleApi{
		clientQuery:   clientQuery,
		clientCommand: clientCommand,
		logger:        logger,
		mapping:       mapping,
		apiHandler:    apiHandler,
		cache:         cache,
	}

	routerUser := router.Group("/api/user")

	routerUser.GET("", apiHandler.Handle("get-user-findalluser", userHandler.FindAllUser))
	routerUser.GET("/:id", apiHandler.Handle("get-user-findbyid", userHandler.FindById))
	routerUser.GET("/active", apiHandler.Handle("get-user-findbyactive", userHandler.FindByActive))
	routerUser.GET("/trashed", apiHandler.Handle("get-user-findbytrashed", userHandler.FindByTrashed))

	routerUser.POST("/create", apiHandler.Handle("post-user-create", userHandler.Create))
	routerUser.POST("/update/:id", apiHandler.Handle("post-user-update", userHandler.Update))

	routerUser.POST("/trashed/:id", apiHandler.Handle("post-user-trasheduser", userHandler.TrashedUser))
	routerUser.POST("/restore/:id", apiHandler.Handle("post-user-restoreuser", userHandler.RestoreUser))
	routerUser.DELETE("/permanent/:id", apiHandler.Handle("delete-user-deleteuserpermanent", userHandler.DeleteUserPermanent))

	routerUser.POST("/restore/all", apiHandler.Handle("post-user-restorealluser", userHandler.RestoreAllUser))
	routerUser.POST("/permanent/all", apiHandler.Handle("delete-user-deletealluserpermanent", userHandler.DeleteAllUserPermanent))

	return userHandler
}

// List all user (paginated)
// List all user (paginated)
// List all user (paginated)
// @Summary List all user (paginated)
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationUser
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user [get]
// @Security BearerAuth
func (h *userHandleApi) FindAllUser(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("user:findalluser:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationUser](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &userpb.FindAllUserRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationUser(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get user by ID
// Get user by ID
// Get user by ID
// @Summary Get user by ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseUser
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/:id [get]
// @Security BearerAuth
func (h *userHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid user ID")
	}

	cacheKey := fmt.Sprintf("user:findbyid:id_%d", id)
	if cached, found := gateway_cache.Get[response.ApiResponseUser](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &userpb.FindByIdUserRequest{
		Id: int32(id),
	}

	res, err := h.clientQuery.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUser(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active user
// List active user
// List active user
// @Summary List active user
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationUserDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/active [get]
// @Security BearerAuth
func (h *userHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("user:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationUserDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &userpb.FindAllUserRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationUserDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed user
// List trashed user
// List trashed user
// @Summary List trashed user
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationUserDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/trashed [get]
// @Security BearerAuth
func (h *userHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("user:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationUserDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &userpb.FindAllUserRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationUserDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create user
// Create user
// Create user
// @Summary Create user
// @Tags User
// @Accept json
// @Produce json
// @Param request body requests.CreateUserRequest true "Request body"
// @Success 200 {object} response.ApiResponseUser
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/create [post]
// @Security BearerAuth
func (h *userHandleApi) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var body requests.CreateUserRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &userpb.CreateUserRequest{
		Firstname:       body.FirstName,
		Lastname:        body.LastName,
		Email:           body.Email,
		Password:        body.Password,
		ConfirmPassword: body.ConfirmPassword,
	}

	res, err := h.clientCommand.Create(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUser(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "user:*")
	return c.JSON(http.StatusOK, so)
}

// Update user
// Update user
// Update user
// @Summary Update user
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateUserRequest true "Request body"
// @Success 200 {object} response.ApiResponseUser
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/update/:id [post]
// @Security BearerAuth
func (h *userHandleApi) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid user ID")
	}

	var body requests.UpdateUserRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &userpb.UpdateUserRequest{
		Id:              int32(id),
		Firstname:       body.FirstName,
		Lastname:        body.LastName,
		Email:           body.Email,
		Password:        body.Password,
		ConfirmPassword: body.ConfirmPassword,
	}

	res, err := h.clientCommand.Update(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUser(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "user:*")
	return c.JSON(http.StatusOK, so)
}

// Trash user
// Trash user
// Trash user
// @Summary Trash user
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseUserDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/trashed/:id [post]
// @Security BearerAuth
func (h *userHandleApi) TrashedUser(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid user ID")
	}

	req := &userpb.FindByIdUserRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.TrashedUser(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUserDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "user:*")
	return c.JSON(http.StatusOK, so)
}

// Restore user
// Restore user
// Restore user
// @Summary Restore user
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseUserDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/restore/:id [post]
// @Security BearerAuth
func (h *userHandleApi) RestoreUser(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid user ID")
	}

	req := &userpb.FindByIdUserRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.RestoreUser(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUserDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "user:*")
	return c.JSON(http.StatusOK, so)
}

// Delete user permanently
// Delete user permanently
// Delete user permanently
// @Summary Delete user permanently
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseUserDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/permanent/:id [delete]
// @Security BearerAuth
func (h *userHandleApi) DeleteUserPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid user ID")
	}

	req := &userpb.FindByIdUserRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.DeleteUserPermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUserDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "user:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed user
// Restore all trashed user
// Restore all trashed user
// @Summary Restore all trashed user
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseUserAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/restore/all [post]
// @Security BearerAuth
func (h *userHandleApi) RestoreAllUser(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllUser(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUserAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "user:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed user permanently
// Delete all trashed user permanently
// Delete all trashed user permanently
// @Summary Delete all trashed user permanently
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseUserAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/user/permanent/all [post]
// @Security BearerAuth
func (h *userHandleApi) DeleteAllUserPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllUserPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseUserAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "user:*")
	return c.JSON(http.StatusOK, so)
}
