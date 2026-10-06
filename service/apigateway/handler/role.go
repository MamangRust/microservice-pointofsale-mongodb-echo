package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	rolepb "github.com/MamangRust/microservice-point-of-sale-pb/role"
	userrolepb "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type roleHandleApi struct {
	clientQuery    rolepb.RoleServiceClient
	clientCommand  rolepb.RoleServiceClient
	clientUserRole userrolepb.UserRoleServiceClient
	logger         logger.LoggerInterface
	mapping        response_api.RoleResponseMapper
	apiHandler     errors.ApiHandler
	cache          *gateway_cache.GatewayCache
}

func NewHandlerRole(
	router *echo.Echo,
	clientQuery rolepb.RoleServiceClient,
	clientCommand rolepb.RoleServiceClient,
	clientUserRole userrolepb.UserRoleServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.RoleResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *roleHandleApi {
	roleHandler := &roleHandleApi{
		clientQuery:    clientQuery,
		clientCommand:  clientCommand,
		clientUserRole: clientUserRole,
		logger:         logger,
		mapping:        mapping,
		apiHandler:     apiHandler,
		cache:          cache,
	}

	routerRole := router.Group("/api/role")

	routerRole.GET("", apiHandler.Handle("get-role-findall", roleHandler.FindAll))
	routerRole.GET("/:id", apiHandler.Handle("get-role-findbyid", roleHandler.FindById))
	routerRole.GET("/active", apiHandler.Handle("get-role-findbyactive", roleHandler.FindByActive))
	routerRole.GET("/trashed", apiHandler.Handle("get-role-findbytrashed", roleHandler.FindByTrashed))
	routerRole.GET("/user/:user_id", apiHandler.Handle("get-role-findbyuserid", roleHandler.FindByUserId))
	routerRole.POST("", apiHandler.Handle("post-role-create", roleHandler.Create))
	routerRole.POST("/update/:id", apiHandler.Handle("post-role-update", roleHandler.Update))
	routerRole.POST("/trashed/:id", apiHandler.Handle("post-role-trashed", roleHandler.Trashed))
	routerRole.POST("/restore/:id", apiHandler.Handle("post-role-restore", roleHandler.Restore))
	routerRole.DELETE("/permanent/:id", apiHandler.Handle("delete-role-deletepermanent", roleHandler.DeletePermanent))
	routerRole.POST("/restore/all", apiHandler.Handle("post-role-restoreall", roleHandler.RestoreAll))
	routerRole.DELETE("/permanent-all", apiHandler.Handle("delete-role-deleteallpermanent", roleHandler.DeleteAllPermanent))

	return roleHandler
}

// List all role (paginated)
// List all role (paginated)
// List all role (paginated)
// @Summary List all role (paginated)
// @Tags Role
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationRole
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role [get]
// @Security BearerAuth
func (h *roleHandleApi) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("role:findall:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationRole](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &rolepb.FindAllRoleRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAllRole(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationRole(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get role by ID
// Get role by ID
// Get role by ID
// @Summary Get role by ID
// @Tags Role
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseRole
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/:id [get]
// @Security BearerAuth
func (h *roleHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("invalid role ID")
	}

	cacheKey := fmt.Sprintf("role:findbyid:id_%d", roleID)
	if cached, found := gateway_cache.Get[response.ApiResponseRole](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &rolepb.FindByIdRoleRequest{
		RoleId: int32(roleID),
	}

	res, err := h.clientQuery.FindByIdRole(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRole(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active role
// List active role
// List active role
// @Summary List active role
// @Tags Role
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationRoleDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/active [get]
// @Security BearerAuth
func (h *roleHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("role:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationRoleDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &rolepb.FindAllRoleRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationRoleDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed role
// List trashed role
// List trashed role
// @Summary List trashed role
// @Tags Role
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationRoleDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/trashed [get]
// @Security BearerAuth
func (h *roleHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("role:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationRoleDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &rolepb.FindAllRoleRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationRoleDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get role by user ID
// Get role by user ID
// Get role by user ID
// @Summary Get role by user ID
// @Tags Role
// @Accept json
// @Produce json
// @Param user_id path int true "user id"
// @Success 200 {object} response.ApiResponseRole
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/user/:user_id [get]
// @Security BearerAuth
func (h *roleHandleApi) FindByUserId(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || userID <= 0 {
		return errors.NewBadRequestError("invalid user ID")
	}

	cacheKey := fmt.Sprintf("role:findbyuserid:user_id_%d", userID)
	if cached, found := gateway_cache.Get[response.ApiResponsesRole](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &userrolepb.FindByIdUserRoleRequest{
		UserId: int32(userID),
	}

	res, err := h.clientUserRole.FindByUserId(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesRole(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create role
// Create role
// Create role
// @Summary Create role
// @Tags Role
// @Accept json
// @Produce json
// @Param request body requests.CreateRoleRequest true "Request body"
// @Success 200 {object} response.ApiResponseRole
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role [post]
// @Security BearerAuth
func (h *roleHandleApi) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := req.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	reqPb := &rolepb.CreateRoleRequest{
		Name: req.Name,
	}

	res, err := h.clientCommand.CreateRole(ctx, reqPb)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRole(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "role:*")
	return c.JSON(http.StatusOK, so)
}

// Update role
// Update role
// Update role
// @Summary Update role
// @Tags Role
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateRoleRequest true "Request body"
// @Success 200 {object} response.ApiResponseRole
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/update/:id [post]
// @Security BearerAuth
func (h *roleHandleApi) Update(c echo.Context) error {
	ctx := c.Request().Context()
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("invalid role ID")
	}

	var req requests.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := req.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	reqPb := &rolepb.UpdateRoleRequest{
		Id:   int32(roleID),
		Name: req.Name,
	}

	res, err := h.clientCommand.UpdateRole(ctx, reqPb)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRole(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "role:*")
	return c.JSON(http.StatusOK, so)
}

// Trash role
// Trash role
// Trash role
// @Summary Trash role
// @Tags Role
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseRole
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/trashed/:id [post]
// @Security BearerAuth
func (h *roleHandleApi) Trashed(c echo.Context) error {
	ctx := c.Request().Context()
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("invalid role ID")
	}

	req := &rolepb.FindByIdRoleRequest{
		RoleId: int32(roleID),
	}

	res, err := h.clientCommand.TrashedRole(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRole(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "role:*")
	return c.JSON(http.StatusOK, so)
}

// Restore role
// Restore role
// Restore role
// @Summary Restore role
// @Tags Role
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseRole
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/restore/:id [post]
// @Security BearerAuth
func (h *roleHandleApi) Restore(c echo.Context) error {
	ctx := c.Request().Context()
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("invalid role ID")
	}

	req := &rolepb.FindByIdRoleRequest{
		RoleId: int32(roleID),
	}

	res, err := h.clientCommand.RestoreRole(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRole(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "role:*")
	return c.JSON(http.StatusOK, so)
}

// Delete role permanently
// Delete role permanently
// Delete role permanently
// @Summary Delete role permanently
// @Tags Role
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseRoleDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/permanent/:id [delete]
// @Security BearerAuth
func (h *roleHandleApi) DeletePermanent(c echo.Context) error {
	ctx := c.Request().Context()
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("invalid role ID")
	}

	req := &rolepb.FindByIdRoleRequest{
		RoleId: int32(roleID),
	}

	res, err := h.clientCommand.DeleteRolePermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRoleDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "role:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed role
// Restore all trashed role
// Restore all trashed role
// @Summary Restore all trashed role
// @Tags Role
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseRoleAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/restore/all [post]
// @Security BearerAuth
func (h *roleHandleApi) RestoreAll(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllRole(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRoleAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "role:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed role permanently
// Delete all trashed role permanently
// Delete all trashed role permanently
// @Summary Delete all trashed role permanently
// @Tags Role
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseRoleAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/role/permanent-all [delete]
// @Security BearerAuth
func (h *roleHandleApi) DeleteAllPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllRolePermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseRoleAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "role:*")
	return c.JSON(http.StatusOK, so)
}
