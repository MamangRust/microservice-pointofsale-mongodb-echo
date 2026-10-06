package order_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderRepositoryTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       *repository.Repositories
	orderID    int
	merchantID int
	cashierID  int
}

func (s *OrderRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextSales)
	s.ts = ts
	s.Require().NoError(err)

	db := s.ts.MongoDB
	ctx := context.Background()
	now := time.Now().UTC()

	userID, _ := database.GetNextID(ctx, s.ts.DBForCollection("users"), "users")
	s.ts.Coll("users").InsertOne(ctx, bson.M{"_id": userID, "firstname": "Order", "lastname": "Repo", "email": "order.repo@example.com", "password": "password123", "verification_code": "test-verify", "is_verified": true, "created_at": now, "updated_at": now})

	mid, _ := database.GetNextID(ctx, s.ts.DBForCollection("merchants"), "merchants")
	s.merchantID = int(mid)
	s.ts.Coll("merchants").InsertOne(ctx, bson.M{"_id": mid, "user_id": userID, "name": "Order Test Merchant", "description": "Desc", "address": "Addr", "contact_email": "o@example.com", "contact_phone": "123", "status": "active", "created_at": now, "updated_at": now})

	cid, _ := database.GetNextID(ctx, s.ts.DBForCollection("cashiers"), "cashiers")
	s.cashierID = int(cid)
	s.ts.Coll("cashiers").InsertOne(ctx, bson.M{"_id": cid, "merchant_id": mid, "user_id": userID, "name": "Order Test Cashier", "created_at": now, "updated_at": now})

	oid, _ := database.GetNextID(ctx, db, "orders")
	s.orderID = int(oid)
	db.Collection("orders").InsertOne(ctx, bson.M{"_id": oid, "merchant_id": mid, "cashier_id": cid, "total_price": int64(5000), "created_at": now, "updated_at": now})

	s.repo = repository.NewRepositories(db, nil, nil, nil, nil)
}

func (s *OrderRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *OrderRepositoryTestSuite) Test1_FindById() {
	s.Require().NotZero(s.orderID)
	ctx := context.Background()
	found, err := s.repo.OrderQuery.FindById(ctx, s.orderID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int32(s.orderID), found.OrderID)
}

func (s *OrderRepositoryTestSuite) Test2_CreateOrder() {
	ctx := context.Background()
	req := &requests.CreateOrderRecordRequest{MerchantID: s.merchantID, CashierID: s.cashierID, TotalPrice: 10000}
	res, err := s.repo.OrderCommand.CreateOrder(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *OrderRepositoryTestSuite) Test3_UpdateOrder() {
	s.Require().NotZero(s.orderID)
	ctx := context.Background()
	req := &requests.UpdateOrderRecordRequest{OrderID: s.orderID, TotalPrice: 15000}
	res, err := s.repo.OrderCommand.UpdateOrder(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(int64(15000), res.TotalPrice)
}

func (s *OrderRepositoryTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.orderID)
	ctx := context.Background()
	trashed, err := s.repo.OrderCommand.TrashedOrder(ctx, s.orderID)
	s.NoError(err)
	s.NotNil(trashed)
	restored, err := s.repo.OrderCommand.RestoreOrder(ctx, s.orderID)
	s.NoError(err)
	s.NotNil(restored)
	found, err := s.repo.OrderQuery.FindById(ctx, s.orderID)
	s.NoError(err)
	s.NotNil(found)
}

func (s *OrderRepositoryTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.orderID)
	ctx := context.Background()
	_, err := s.repo.OrderCommand.TrashedOrder(ctx, s.orderID)
	s.NoError(err)
	success, err := s.repo.OrderCommand.DeleteOrderPermanent(ctx, s.orderID)
	s.NoError(err)
	s.True(success)
}

func TestOrderRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderRepositoryTestSuite))
}
