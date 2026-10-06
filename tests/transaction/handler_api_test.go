package transaction_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	apigw_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	apigateway "github.com/MamangRust/microservice-point-of-sale-apigateway/handler"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TransactionApiTestSuite struct {
	tests.BaseTestSuite
	echo          *echo.Echo
	transactionID int
	userID        int
	cashierID     int
	merchID       int
	orderID       int
}

func (s *TransactionApiTestSuite) SetupSuite() {
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
	merchID := s.SeedMerchant(ctx, userID)
	catID := s.SeedCategory(ctx)
	prodID := s.SeedProduct(ctx, merchID, catID)
	orderID := s.SeedOrder(ctx, userID, merchID, prodID)
	s.SeedOrderItem(ctx, orderID, prodID)

	// cashier_id as seeded by SeedOrder (cashiers collection)
	var cashierDoc bson.M
	err := s.Coll("cashiers").FindOne(ctx, bson.M{"user_id": int32(userID), "merchant_id": int32(merchID)}).Decode(&cashierDoc)
	s.Require().NoError(err, "failed to query cashier")
	s.cashierID = int(bsonMToInt32FromTx(cashierDoc, "_id"))

	s.userID = userID
	s.merchID = merchID
	s.orderID = orderID

	s.echo = echo.New()

	commandQueryClient := pbtransaction.NewTransactionServiceClient(s.Conns["transaction"])
	commandCommandClient := pbtransaction.NewTransactionServiceClient(s.Conns["transaction"])

	apigateway.NewHandlerTransaction(
		s.echo,
		commandQueryClient,
		commandCommandClient,
		s.Log,
		response_api.NewTransactionResponseMapper(),
		errors.NewApiHandler(s.Obs, s.Log),
		apigw_cache.NewGatewayCache(s.GetCacheStore()),
	)
}

func (s *TransactionApiTestSuite) TestTransactionApiLifecycle() {
	// 1. Create
	fields := map[string]interface{}{
		"order_id":       s.orderID,
		"cashier_id":     s.cashierID,
		"payment_method": "Transfer Bank",
		"amount":         100000,
	}
	body, _ := json.Marshal(fields)
	req := httptest.NewRequest(http.MethodPost, "/api/transaction/create", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.transactionID = int(data["id"].(float64))

	// 2. FindAll
	req = httptest.NewRequest(http.MethodGet, "/api/transaction", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 3. FindById
	s.Require().NotZero(s.transactionID)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/transaction/%d", s.transactionID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal(float64(s.transactionID), data["id"])

	// 4. FindByMerchant
	s.Require().NotZero(s.merchID)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/transaction/merchant/%d", s.merchID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/transaction/active", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 6. Update (JSON body — handler binds via json tags)
	updateBody, _ := json.Marshal(map[string]interface{}{
		"order_id":       s.orderID,
		"cashier_id":     s.cashierID,
		"payment_method": "GOPAY",
		"amount":         100000,
	})
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/transaction/update/%d", s.transactionID), bytes.NewBuffer(updateBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/transaction/trashed/%d", s.transactionID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 8. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/transaction/trashed", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/transaction/restore/%d", s.transactionID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. DeletePermanent
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/transaction/trashed/%d", s.transactionID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/transaction/permanent/%d", s.transactionID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 11. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/transaction/restore/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 12. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/transaction/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestTransactionApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionApiTestSuite))
}
