package order_test

import (
	"context"
	"testing"
	"time"

	order_cache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-order/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderServiceTestSuite struct {
	tests.BaseTestSuite
	svc       *service.Service
	orderID   int
	cashierID int
	productID int
}

func (s *OrderServiceTestSuite) SetupSuite() {
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()

	s.SetupOrderService()

	db := s.MongoDB()
	ctx := context.Background()
	now := time.Now().UTC()

	userID, _ := database.GetNextID(ctx, s.DBForCollection("users"), "users")
	s.Coll("users").InsertOne(ctx, bson.M{"_id": userID, "firstname": "Order", "lastname": "Svc", "email": "order.svc@example.com", "password": "password123", "verification_code": "test-verify", "is_verified": true, "created_at": now, "updated_at": now})

	mid, _ := database.GetNextID(ctx, s.DBForCollection("merchants"), "merchants")
	s.orderID = int(mid)
	s.Coll("merchants").InsertOne(ctx, bson.M{"_id": mid, "user_id": userID, "name": "Order Svc Merchant", "description": "Desc", "address": "Addr", "contact_email": "os@example.com", "contact_phone": "123", "status": "active", "created_at": now, "updated_at": now})

	cid, _ := database.GetNextID(ctx, s.DBForCollection("cashiers"), "cashiers")
	s.cashierID = int(cid)
	s.Coll("cashiers").InsertOne(ctx, bson.M{"_id": cid, "merchant_id": mid, "user_id": userID, "name": "Order Cashier", "created_at": now, "updated_at": now})

	catID, _ := database.GetNextID(ctx, s.DBForCollection("categories"), "categories")
	s.Coll("categories").InsertOne(ctx, bson.M{"_id": catID, "name": "Order Category", "description": "Desc", "created_at": now, "updated_at": now})

	pid, _ := database.GetNextID(ctx, s.DBForCollection("products"), "products")
	s.productID = int(pid)
	s.Coll("products").InsertOne(ctx, bson.M{"_id": pid, "merchant_id": mid, "category_id": catID, "name": "Order Product", "description": "Desc", "price": int32(10000), "count_in_stock": int32(100), "brand": "Brand", "weight": int32(100), "image_product": "img", "created_at": now, "updated_at": now})

	mencache := order_cache.NewMencache(s.GetCacheStore())
	repos := repository.NewRepositories(
		db,
		pbcashier.NewCashierServiceClient(s.Conns["cashier"]),
		pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]),
		pbproduct.NewProductServiceClient(s.Conns["product"]),
		pborderitem.NewOrderItemServiceClient(s.Conns["order-item"]),
	)

	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})
}

func (s *OrderServiceTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *OrderServiceTestSuite) TestOrderLifecycle() {
	ctx := context.Background()

	req := &requests.CreateOrderRequest{
		MerchantID: s.orderID,
		CashierID:  s.cashierID,
		Items: []requests.CreateOrderItemRequest{
			{ProductID: s.productID, Quantity: 1},
		},
	}
	created, err := s.svc.OrderCommand.CreateOrder(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	orderID := int(created.OrderID)

	var itemDoc bson.M
	err = s.MongoDB().Collection("order_items").FindOne(ctx, bson.M{"order_id": int32(orderID)}).Decode(&itemDoc)
	s.Require().NoError(err)
	orderItemID := int(bsonMToInt32FromDoc(itemDoc, "_id"))

	found, err := s.svc.OrderQuery.FindById(ctx, orderID)
	s.Require().NoError(err)
	s.NotNil(found)

	updateReq := &requests.UpdateOrderRequest{
		OrderID: &orderID,
		Items: []requests.UpdateOrderItemRequest{
			{OrderItemID: orderItemID, ProductID: s.productID, Quantity: 2},
		},
	}
	updated, err := s.svc.OrderCommand.UpdateOrder(ctx, updateReq)
	s.Require().NoError(err)
	s.NotNil(updated)

	_, total, err := s.svc.OrderQuery.FindAll(ctx, &requests.FindAllOrders{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	_, err = s.svc.OrderCommand.TrashedOrder(ctx, orderID)
	s.Require().NoError(err)

	_, totalTrashed, err := s.svc.OrderQuery.FindByTrashed(ctx, &requests.FindAllOrders{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	active, _, err := s.svc.OrderQuery.FindByActive(ctx, &requests.FindAllOrders{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, o := range active {
		s.NotEqual(orderID, int(o.OrderID))
	}

	_, err = s.svc.OrderCommand.RestoreOrder(ctx, orderID)
	s.Require().NoError(err)

	_, err = s.svc.OrderCommand.TrashedOrder(ctx, orderID)
	s.Require().NoError(err)
	success, err := s.svc.OrderCommand.DeleteOrderPermanent(ctx, orderID)
	s.Require().NoError(err)
	s.True(success)

	o1, _ := s.svc.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRequest{
		MerchantID: s.orderID, CashierID: s.cashierID,
		Items: []requests.CreateOrderItemRequest{{ProductID: s.productID, Quantity: 1}},
	})
	o2, _ := s.svc.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRequest{
		MerchantID: s.orderID, CashierID: s.cashierID,
		Items: []requests.CreateOrderItemRequest{{ProductID: s.productID, Quantity: 2}},
	})

	s.svc.OrderCommand.TrashedOrder(ctx, int(o1.OrderID))
	s.svc.OrderCommand.TrashedOrder(ctx, int(o2.OrderID))

	resRestoreAll, err := s.svc.OrderCommand.RestoreAllOrder(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.OrderCommand.TrashedOrder(ctx, int(o1.OrderID))
	s.svc.OrderCommand.TrashedOrder(ctx, int(o2.OrderID))

	resDeleteAll, err := s.svc.OrderCommand.DeleteAllOrderPermanent(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func bsonMToInt32FromDoc(d bson.M, key string) int32 {
	if val, ok := d[key]; ok {
		switch n := val.(type) {
		case int32:
			return n
		case int64:
			return int32(n)
		case float64:
			return int32(n)
		}
	}
	return 0
}

func TestOrderServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderServiceTestSuite))
}
