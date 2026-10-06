package product_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	apigateway "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	gatewayhandler "github.com/MamangRust/microservice-point-of-sale-apigateway/handler"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	app_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ProductApiTestSuite struct {
	tests.BaseTestSuite
	echo         *echo.Echo
	productID    int
	merchantID   int
	categoryID   int
	categoryName string
}

func (s *ProductApiTestSuite) SetupSuite() {
	s.Context = tests.ContextCatalog
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()

	// Seed dependencies
	ctx := context.Background()
	userID := s.SeedUser(ctx)
	s.categoryID = s.SeedCategory(ctx)
	s.merchantID = s.SeedMerchant(ctx, userID)
	s.categoryName = "Seed Category" // Default from SeedCategory if not specified

	s.echo = echo.New()
	apiHandler := app_errors.NewApiHandler(s.Obs, s.Log)
	productMapper := response_api.NewProductResponseMapper()
	gatewayCache := apigateway.NewGatewayCache(s.GetCacheStore())
	commandQueryClient := pbproduct.NewProductServiceClient(s.Conns["product"])
	commandCommandClient := pbproduct.NewProductServiceClient(s.Conns["product"])

	gatewayhandler.NewHandlerProduct(
		s.echo,
		commandQueryClient,
		commandCommandClient,
		s.Log,
		productMapper,
		&tests.MockImageUpload{},
		apiHandler,
		gatewayCache,
	)
}

func (s *ProductApiTestSuite) TestProductApiLifecycle() {
	// 1. Create
	fields := map[string]string{
		"merchant_id":    fmt.Sprintf("%d", s.merchantID),
		"category_id":    fmt.Sprintf("%d", s.categoryID),
		"name":           "Test Product",
		"description":    "Test Description",
		"price":          "1000",
		"count_in_stock": "10",
		"brand":          "Test Brand",
		"weight":         "1",
		"rating":         "5",
		"slug_product":   "test-product",
	}
	body, contentType := s.BuildMultipartRequestBody(fields, "image_product", "product.jpg")
	req := httptest.NewRequest(http.MethodPost, "/api/product/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.Equal("Test Product", data["name"])
	s.productID = int(data["id"].(float64))

	// 2. FindById
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/product/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal(float64(s.productID), data["id"])

	// 3. FindAll
	req = httptest.NewRequest(http.MethodGet, "/api/product", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/product/active", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. FindByMerchant
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/product/merchant/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 6. FindByCategory
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/product/category/%s", url.PathEscape(s.categoryName)), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. Update
	updateFields := map[string]string{
		"merchant_id":    fmt.Sprintf("%d", s.merchantID),
		"category_id":    fmt.Sprintf("%d", s.categoryID),
		"name":           "Updated Product",
		"description":    "Updated Description",
		"price":          "2000",
		"count_in_stock": "20",
		"brand":          "Updated Brand",
		"weight":         "2",
		"rating":         "4",
		"slug_product":   "updated-product",
	}
	body, contentType = s.BuildMultipartRequestBody(updateFields, "image_product", "updated.jpg")
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product/update/%d", s.productID), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, contentType)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal("Updated Product", data["name"])

	// 8. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product/trashed/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/product/trashed", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product/restore/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 11. DeletePermanent
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product/trashed/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/product/permanent/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 12. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/product/restore/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 13. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/product/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestProductApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductApiTestSuite))
}
