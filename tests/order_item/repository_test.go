package order_item_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderItemRepositoryTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	repo          *repository.Repositories
	orderItemID   int
	orderID       int
	productID     int
	createdItemID int
}

func (s *OrderItemRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextSales)
	s.ts = ts
	s.Require().NoError(err)

	db := s.ts.MongoDB
	s.repo = repository.NewRepositories(db)
	ctx := context.Background()
	now := time.Now().UTC()

	userID, _ := database.GetNextID(ctx, s.ts.DBForCollection("users"), "users")
	s.ts.Coll("users").InsertOne(ctx, bson.M{"_id": userID, "firstname": "OItem", "lastname": "Repo", "email": "oitem.repo@example.com", "password": "password123", "verification_code": "test-verify", "is_verified": true, "created_at": now, "updated_at": now})

	merchantID, _ := database.GetNextID(ctx, s.ts.DBForCollection("merchants"), "merchants")
	s.ts.Coll("merchants").InsertOne(ctx, bson.M{"_id": merchantID, "user_id": userID, "name": "OItem Merchant", "status": "active", "created_at": now, "updated_at": now})

	categoryID, _ := database.GetNextID(ctx, s.ts.DBForCollection("categories"), "categories")
	s.ts.Coll("categories").InsertOne(ctx, bson.M{"_id": categoryID, "name": "OItem Category", "created_at": now, "updated_at": now})

	productID, _ := database.GetNextID(ctx, s.ts.DBForCollection("products"), "products")
	s.productID = int(productID)
	s.ts.Coll("products").InsertOne(ctx, bson.M{"_id": productID, "merchant_id": merchantID, "category_id": categoryID, "name": "OItem Product", "price": int32(1000), "count_in_stock": int32(10), "created_at": now, "updated_at": now})

	cashierID, _ := database.GetNextID(ctx, s.ts.DBForCollection("cashiers"), "cashiers")
	s.ts.Coll("cashiers").InsertOne(ctx, bson.M{"_id": cashierID, "merchant_id": merchantID, "user_id": userID, "name": "Test Cashier", "created_at": now, "updated_at": now})

	orderID, _ := database.GetNextID(ctx, db, "orders")
	s.orderID = int(orderID)
	db.Collection("orders").InsertOne(ctx, bson.M{"_id": orderID, "merchant_id": merchantID, "cashier_id": cashierID, "total_price": int64(5000), "created_at": now, "updated_at": now})

	orderItemID, _ := database.GetNextID(ctx, db, "order_items")
	s.orderItemID = int(orderItemID)
	db.Collection("order_items").InsertOne(ctx, bson.M{"_id": orderItemID, "order_id": orderID, "product_id": productID, "quantity": int32(2), "price": int32(5000), "created_at": now, "updated_at": now})
}

func (s *OrderItemRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *OrderItemRepositoryTestSuite) Test1_FindAll() {
	ctx := context.Background()
	results, total, err := s.repo.OrderItemQuery.FindAllOrderItems(ctx, &requests.FindAllOrderItems{Search: "", Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(*total, 1)
	s.GreaterOrEqual(len(results), 1)
}

func (s *OrderItemRepositoryTestSuite) Test2_FindOrderItemByOrder() {
	s.Require().NotZero(s.orderID)
	ctx := context.Background()
	found, err := s.repo.OrderItemQuery.FindOrderItemByOrder(ctx, s.orderID)
	s.NoError(err)
	s.NotEmpty(found)
	s.Equal(int32(s.orderID), found[0].OrderID)
}

func (s *OrderItemRepositoryTestSuite) Test3_FindByActive() {
	ctx := context.Background()
	results, total, err := s.repo.OrderItemQuery.FindByActive(ctx, &requests.FindAllOrderItems{Search: "", Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(*total, 1)
	s.GreaterOrEqual(len(results), 1)
}

func (s *OrderItemRepositoryTestSuite) Test4_FindByTrashed() {
	ctx := context.Background()
	db := s.ts.MongoDB
	_, _ = db.Collection("order_items").UpdateOne(ctx, bson.M{"_id": int32(s.orderItemID)}, bson.M{"$set": bson.M{"deleted_at": time.Now().UTC()}})

	results, total, err := s.repo.OrderItemQuery.FindByTrashed(ctx, &requests.FindAllOrderItems{Search: "", Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(*total, 1)
	s.GreaterOrEqual(len(results), 1)

	_, _ = db.Collection("order_items").UpdateOne(ctx, bson.M{"_id": int32(s.orderItemID)}, bson.M{"$unset": bson.M{"deleted_at": ""}})
}

func (s *OrderItemRepositoryTestSuite) Test5_CreateOrderItem() {
	ctx := context.Background()
	created, err := s.repo.OrderItemCommand.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{
		OrderID:   s.orderID,
		ProductID: s.productID,
		Quantity:  3,
		Price:     1500,
	})
	s.Require().NoError(err)
	s.Require().NotNil(created)
	s.NotZero(created.OrderItemID)
	s.Equal(int32(s.orderID), created.OrderID)
	s.Equal(int32(s.productID), created.ProductID)
	s.Equal(int32(3), created.Quantity)
	s.Equal(int64(1500), created.Price)

	s.createdItemID = int(created.OrderItemID)
}

func (s *OrderItemRepositoryTestSuite) Test6_UpdateOrderItem() {
	s.Require().NotZero(s.createdItemID)
	ctx := context.Background()

	updated, err := s.repo.OrderItemCommand.UpdateOrderItem(ctx, &requests.UpdateOrderItemRecordRequest{
		OrderItemID: s.createdItemID,
		OrderID:     s.orderID,
		ProductID:   s.productID,
		Quantity:    7,
		Price:       2500,
	})
	s.Require().NoError(err)
	s.Require().NotNil(updated)
	s.Equal(int32(7), updated.Quantity)
	s.Equal(int64(2500), updated.Price)
}

// Test7_DeleteOrderItem_DeletesOnlyTargetItem is a regression test: the row
// must be matched by _id (the order item's own ID), not by order_id. Deleting
// one item must leave the other items of the same order untouched.
func (s *OrderItemRepositoryTestSuite) Test7_DeleteOrderItem_DeletesOnlyTargetItem() {
	ctx := context.Background()
	db := s.ts.MongoDB

	first, err := s.repo.OrderItemCommand.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{
		OrderID: s.orderID, ProductID: s.productID, Quantity: 1, Price: 100,
	})
	s.Require().NoError(err)

	second, err := s.repo.OrderItemCommand.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{
		OrderID: s.orderID, ProductID: s.productID, Quantity: 2, Price: 200,
	})
	s.Require().NoError(err)

	s.Require().NoError(s.repo.OrderItemCommand.DeleteOrderItem(ctx, int(first.OrderItemID)))

	goneCount, err := db.Collection("order_items").CountDocuments(ctx, bson.M{"_id": first.OrderItemID})
	s.Require().NoError(err)
	s.Equal(int64(0), goneCount, "the deleted order item must be gone")

	keptCount, err := db.Collection("order_items").CountDocuments(ctx, bson.M{"_id": second.OrderItemID})
	s.Require().NoError(err)
	s.Equal(int64(1), keptCount, "other order items of the same order must survive")
}

func TestOrderItemRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderItemRepositoryTestSuite))
}
