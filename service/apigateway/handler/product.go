package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	productpb "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/upload_image"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type productHandleApi struct {
	clientQuery   productpb.ProductServiceClient
	clientCommand productpb.ProductServiceClient
	logger        logger.LoggerInterface
	mapping       response_api.ProductResponseMapper
	upload_image  upload_image.ImageUploads
	apiHandler    errors.ApiHandler
	cache         *gateway_cache.GatewayCache
}

func NewHandlerProduct(
	router *echo.Echo,
	clientQuery productpb.ProductServiceClient,
	clientCommand productpb.ProductServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.ProductResponseMapper,
	upload_image upload_image.ImageUploads,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *productHandleApi {
	productHandler := &productHandleApi{
		clientQuery:   clientQuery,
		clientCommand: clientCommand,
		logger:        logger,
		mapping:       mapping,
		upload_image:  upload_image,
		apiHandler:    apiHandler,
		cache:         cache,
	}

	routerProduct := router.Group("/api/product")

	routerProduct.GET("", apiHandler.Handle("get-product-findallproduct", productHandler.FindAllProduct))
	routerProduct.GET("/:id", apiHandler.Handle("get-product-findbyid", productHandler.FindById))
	routerProduct.GET("/merchant/:merchant_id", apiHandler.Handle("get-product-findbymerchant", productHandler.FindByMerchant))
	routerProduct.GET("/category/:category_name", apiHandler.Handle("get-product-findbycategory", productHandler.FindByCategory))

	routerProduct.GET("/active", apiHandler.Handle("get-product-findbyactive", productHandler.FindByActive))
	routerProduct.GET("/trashed", apiHandler.Handle("get-product-findbytrashed", productHandler.FindByTrashed))

	routerProduct.POST("/create", apiHandler.Handle("post-product-create", productHandler.Create))
	routerProduct.POST("/update/:id", apiHandler.Handle("post-product-update", productHandler.Update))

	routerProduct.POST("/trashed/:id", apiHandler.Handle("post-product-trashedproduct", productHandler.TrashedProduct))
	routerProduct.POST("/restore/:id", apiHandler.Handle("post-product-restoreproduct", productHandler.RestoreProduct))
	routerProduct.DELETE("/permanent/:id", apiHandler.Handle("delete-product-deleteproductpermanent", productHandler.DeleteProductPermanent))

	routerProduct.POST("/restore/all", apiHandler.Handle("post-product-restoreallproduct", productHandler.RestoreAllProduct))
	routerProduct.POST("/permanent/all", apiHandler.Handle("delete-product-deleteallproductpermanent", productHandler.DeleteAllProductPermanent))

	return productHandler
}

// List all product (paginated)
// List all product (paginated)
// List all product (paginated)
// @Summary List all product (paginated)
// @Tags Product
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationProduct
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product [get]
// @Security BearerAuth
func (h *productHandleApi) FindAllProduct(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("product:findallproduct:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationProduct](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &productpb.FindAllProductRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationProduct(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List product by merchant
// List product by merchant
// List product by merchant
// @Summary List product by merchant
// @Tags Product
// @Accept json
// @Produce json
// @Param merchant_id path int true "merchant id"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsesProduct
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/merchant/:merchant_id [get]
// @Security BearerAuth
func (h *productHandleApi) FindByMerchant(c echo.Context) error {
	merchantID, err := strconv.Atoi(c.Param("merchant_id"))
	if err != nil || merchantID <= 0 {
		return errors.NewBadRequestError("invalid merchant ID")
	}

	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := strings.TrimSpace(c.QueryParam("search"))
	categoryIDStr := c.QueryParam("category_id")
	minPriceStr := c.QueryParam("min_price")
	maxPriceStr := c.QueryParam("max_price")

	cacheKey := fmt.Sprintf("product:findbymerchant:merchant_%d:page_%d:size_%d:search_%s:cat_%s:min_%s:max_%s",
		merchantID, page, pageSize, search, categoryIDStr, minPriceStr, maxPriceStr)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationProduct](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &productpb.FindAllProductMerchantRequest{
		MerchantId: int32(merchantID),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		Search:     search,
	}

	if categoryIDStr != "" {
		if id, err := strconv.Atoi(categoryIDStr); err == nil && id > 0 {
			req.CategoryId = int32(id)
		}
	}

	if minPriceStr != "" {
		if price, err := strconv.Atoi(minPriceStr); err == nil && price >= 0 {
			req.MinPrice = int32(price)
		}
	}

	if maxPriceStr != "" {
		if price, err := strconv.Atoi(maxPriceStr); err == nil && price >= 0 {
			req.MaxPrice = int32(price)
		}
	}

	res, err := h.clientQuery.FindByMerchant(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationProduct(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List products by category
// List products by category
// List products by category
// @Summary List products by category
// @Tags Product
// @Accept json
// @Produce json
// @Param category_name path string true "category name"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsesProduct
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/category/:category_name [get]
// @Security BearerAuth
func (h *productHandleApi) FindByCategory(c echo.Context) error {
	categoryName := c.Param("category_name")
	if categoryName == "" {
		return errors.NewBadRequestError("category name is required")
	}

	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("product:findbycategory:cat_%s:page_%d:size_%d:search_%s", categoryName, page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationProduct](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &productpb.FindAllProductCategoryRequest{
		CategoryName: categoryName,
		Page:         int32(page),
		PageSize:     int32(pageSize),
		Search:       search,
	}

	res, err := h.clientQuery.FindByCategory(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationProduct(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get product by ID
// Get product by ID
// Get product by ID
// @Summary Get product by ID
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseProduct
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/:id [get]
// @Security BearerAuth
func (h *productHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid product ID")
	}

	cacheKey := fmt.Sprintf("product:findbyid:id_%d", id)
	if cached, found := gateway_cache.Get[response.ApiResponseProduct](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &productpb.FindByIdProductRequest{
		Id: int32(id),
	}

	res, err := h.clientQuery.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseProduct(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active product
// List active product
// List active product
// @Summary List active product
// @Tags Product
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationProductDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/active [get]
// @Security BearerAuth
func (h *productHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("product:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationProductDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &productpb.FindAllProductRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationProductDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed product
// List trashed product
// List trashed product
// @Summary List trashed product
// @Tags Product
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationProductDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/trashed [get]
// @Security BearerAuth
func (h *productHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("product:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationProductDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &productpb.FindAllProductRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationProductDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create product
// Create product
// Create product
// @Summary Create product
// @Tags Product
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseProduct
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/create [post]
// @Security BearerAuth
func (h *productHandleApi) Create(c echo.Context) error {
	ctx := c.Request().Context()
	formData, err := h.parseProductForm(c, true)
	if err != nil {
		return err
	}

	req := &productpb.CreateProductRequest{
		MerchantId:   int32(formData.MerchantID),
		CategoryId:   int32(formData.CategoryID),
		Name:         formData.Name,
		Description:  formData.Description,
		Price:        int32(formData.Price),
		CountInStock: int32(formData.CountInStock),
		Brand:        formData.Brand,
		Weight:       int32(formData.Weight),
		ImageProduct: formData.ImagePath,
	}

	res, err := h.clientCommand.Create(ctx, req)
	if err != nil {
		if formData.ImagePath != "" {
			h.upload_image.CleanupImageOnFailure(formData.ImagePath)
		}
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseProduct(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "product:*")
	return c.JSON(http.StatusCreated, so)
}

// Update product
// Update product
// Update product
// @Summary Update product
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseProduct
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/update/:id [post]
// @Security BearerAuth
func (h *productHandleApi) Update(c echo.Context) error {
	ctx := c.Request().Context()
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid product ID")
	}

	formData, err := h.parseProductForm(c, false)
	if err != nil {
		return err
	}

	req := &productpb.UpdateProductRequest{
		ProductId:    int32(productID),
		MerchantId:   int32(formData.MerchantID),
		CategoryId:   int32(formData.CategoryID),
		Name:         formData.Name,
		Description:  formData.Description,
		Price:        int32(formData.Price),
		CountInStock: int32(formData.CountInStock),
		Brand:        formData.Brand,
		Weight:       int32(formData.Weight),
		ImageProduct: formData.ImagePath,
	}

	res, err := h.clientCommand.Update(ctx, req)
	if err != nil {
		if formData.ImagePath != "" {
			h.upload_image.CleanupImageOnFailure(formData.ImagePath)
		}
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseProduct(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "product:*")
	return c.JSON(http.StatusOK, so)
}

// Trash product
// Trash product
// Trash product
// @Summary Trash product
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseProductDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/trashed/:id [post]
// @Security BearerAuth
func (h *productHandleApi) TrashedProduct(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid product ID")
	}

	req := &productpb.FindByIdProductRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.TrashedProduct(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesProductDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "product:*")
	return c.JSON(http.StatusOK, so)
}

// Restore product
// Restore product
// Restore product
// @Summary Restore product
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseProductDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/restore/:id [post]
// @Security BearerAuth
func (h *productHandleApi) RestoreProduct(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid product ID")
	}

	req := &productpb.FindByIdProductRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.RestoreProduct(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesProductDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "product:*")
	return c.JSON(http.StatusOK, so)
}

// Delete product permanently
// Delete product permanently
// Delete product permanently
// @Summary Delete product permanently
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseProductDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/permanent/:id [delete]
// @Security BearerAuth
func (h *productHandleApi) DeleteProductPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid product ID")
	}

	req := &productpb.FindByIdProductRequest{
		Id: int32(id),
	}

	res, err := h.clientCommand.DeleteProductPermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseProductDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "product:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed product
// Restore all trashed product
// Restore all trashed product
// @Summary Restore all trashed product
// @Tags Product
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseProductAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/restore/all [post]
// @Security BearerAuth
func (h *productHandleApi) RestoreAllProduct(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllProduct(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseProductAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "product:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed product permanently
// Delete all trashed product permanently
// Delete all trashed product permanently
// @Summary Delete all trashed product permanently
// @Tags Product
// @Accept json
// @Produce json
// @Param request body requests.ProductFormData true "Request body"
// @Success 200 {object} response.ApiResponseProductAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/product/permanent/all [post]
// @Security BearerAuth
func (h *productHandleApi) DeleteAllProductPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllProductPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseProductAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "product:*")
	return c.JSON(http.StatusOK, so)
}

func (h *productHandleApi) parseProductForm(c echo.Context, requireImage bool) (requests.ProductFormData, error) {
	var formData requests.ProductFormData
	var err error

	formData.MerchantID, err = strconv.Atoi(c.FormValue("merchant_id"))
	if err != nil || formData.MerchantID <= 0 {
		return formData, errors.NewBadRequestError("Please provide a valid merchant ID")
	}

	formData.CategoryID, err = strconv.Atoi(c.FormValue("category_id"))
	if err != nil || formData.CategoryID <= 0 {
		return formData, errors.NewBadRequestError("Please provide a valid category ID")
	}

	formData.Name = strings.TrimSpace(c.FormValue("name"))
	if formData.Name == "" {
		return formData, errors.NewBadRequestError("Product name is required")
	}

	formData.Description = strings.TrimSpace(c.FormValue("description"))
	formData.Brand = strings.TrimSpace(c.FormValue("brand"))

	formData.Price, err = strconv.Atoi(c.FormValue("price"))
	if err != nil || formData.Price <= 0 {
		return formData, errors.NewBadRequestError("Please provide a valid positive price")
	}

	formData.CountInStock, err = strconv.Atoi(c.FormValue("count_in_stock"))
	if err != nil || formData.CountInStock < 0 {
		return formData, errors.NewBadRequestError("Please provide a valid stock count (zero or positive)")
	}

	formData.Weight, err = strconv.Atoi(c.FormValue("weight"))
	if err != nil || formData.Weight <= 0 {
		return formData, errors.NewBadRequestError("Please provide a valid positive weight")
	}

	file, err := c.FormFile("image_product")
	if err != nil {
		if requireImage {
			return formData, errors.NewBadRequestError("A product image is required")
		}
		return formData, nil
	}

	imagePath, err := h.upload_image.ProcessImageUpload(c.Response().Writer, file)
	if err != nil {
		return formData, err
	}

	formData.ImagePath = imagePath
	return formData, nil
}
