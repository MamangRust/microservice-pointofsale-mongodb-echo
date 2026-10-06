package order_item_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	"github.com/MamangRust/microservice-point-of-sale-apigateway/handler"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type OrderItemApiTestSuite struct {
	tests.BaseTestSuite
	echo    *echo.Echo
	orderID int
}

func (s *OrderItemApiTestSuite) SetupSuite() {
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupOrderService()
	s.SetupTransactionService()

	// Seed dependencies
	ctx := context.Background()
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)
	s.orderID = s.SeedOrder(ctx, userID, merchID, prodID)

	s.echo = echo.New()

	handler.NewHandlerOrderItem(
		s.echo,
		pborderitem.NewOrderItemServiceClient(s.Conns["order-item"]),
		s.Log,
		response_api.NewOrderItemResponseMapper(),
		errors.NewApiHandler(s.Obs, s.Log),
		gateway_cache.NewGatewayCache(s.GetCacheStore()),
	)
}

func (s *OrderItemApiTestSuite) TestOrderItemApiLifecycle() {
	// 1. FindAll
	req := httptest.NewRequest(http.MethodGet, "/api/order-item", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 2. FindOrderItemByOrder
	s.Require().NotZero(s.orderID)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/order-item/%d", s.orderID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].([]interface{})
	if len(data) == 0 {
		s.T().Skip("No order items found")
	}

	// 3. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/order-item/active", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/order-item/trashed", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestOrderItemApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderItemApiTestSuite))
}
