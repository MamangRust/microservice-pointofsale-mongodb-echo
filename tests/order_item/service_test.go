package order_item_test

import (
	"context"
	"testing"
	"time"

	item_cache "github.com/MamangRust/microservice-point-of-sale-order-item/cache"
	"github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	"github.com/MamangRust/microservice-point-of-sale-order-item/service"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderItemServiceTestSuite struct {
	suite.Suite
	ts  *tests.TestSuite
	svc *service.Service
}

func (s *OrderItemServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextSales)
	s.ts = ts
	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	db := s.ts.MongoDB

	log, _ := logger.NewLogger("test", nil)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)
	mencache := item_cache.NewMencache(cacheStore)

	obs, _ := observability.NewObservability("test", log)
	repos := repository.NewRepositories(db)

	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache,
		Observability: obs,
	})

	// Seed data
	ctx := context.Background()
	now := time.Now().UTC()

	userID, _ := database.GetNextID(ctx, s.ts.DBForCollection("users"), "users")
	s.ts.Coll("users").InsertOne(ctx, bson.M{"_id": userID, "firstname": "OItem", "lastname": "Svc", "email": "oitem.svc@example.com", "password": "password123", "verification_code": "test-verify", "is_verified": true, "created_at": now, "updated_at": now})

	merchantID, _ := database.GetNextID(ctx, s.ts.DBForCollection("merchants"), "merchants")
	s.ts.Coll("merchants").InsertOne(ctx, bson.M{"_id": merchantID, "user_id": userID, "name": "OI Merchant", "description": "Desc", "address": "Addr", "contact_email": "oi@example.com", "contact_phone": "123", "status": "active", "created_at": now, "updated_at": now})

	s.ts.Coll("cashiers").InsertOne(ctx, bson.M{"_id": int32(1), "merchant_id": merchantID, "user_id": userID, "name": "OI Cashier", "created_at": now, "updated_at": now})

	categoryID, _ := database.GetNextID(ctx, s.ts.DBForCollection("categories"), "categories")
	s.ts.Coll("categories").InsertOne(ctx, bson.M{"_id": categoryID, "name": "Category", "description": "Desc", "created_at": now, "updated_at": now})

	productID, _ := database.GetNextID(ctx, s.ts.DBForCollection("products"), "products")
	s.ts.Coll("products").InsertOne(ctx, bson.M{"_id": productID, "merchant_id": merchantID, "category_id": categoryID, "name": "Product", "description": "Desc", "price": int32(1000), "count_in_stock": int32(10), "brand": "Brand", "weight": int32(1), "image_product": "img.jpg", "created_at": now, "updated_at": now})

	orderID, _ := database.GetNextID(ctx, db, "orders")
	db.Collection("orders").InsertOne(ctx, bson.M{"_id": orderID, "merchant_id": merchantID, "cashier_id": int32(1), "total_price": int64(5000), "created_at": now, "updated_at": now})

	oItemID, _ := database.GetNextID(ctx, db, "order_items")
	db.Collection("order_items").InsertOne(ctx, bson.M{"_id": oItemID, "order_id": int32(orderID), "product_id": int32(productID), "quantity": int32(2), "price": int32(5000), "created_at": now, "updated_at": now})
}

func (s *OrderItemServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *OrderItemServiceTestSuite) TestOrderItemCommandLifecycle() {
	ctx := context.Background()
	db := s.ts.MongoDB

	var orderDoc bson.M
	err := db.Collection("orders").FindOne(ctx, bson.M{"total_price": int64(5000)}).Decode(&orderDoc)
	s.Require().NoError(err)
	orderID := int(bsonMToInt32FromOI(orderDoc, "_id"))

	var productDoc bson.M
	err = s.ts.Coll("products").FindOne(ctx, bson.M{"name": "Product"}).Decode(&productDoc)
	s.Require().NoError(err)
	productID := int(bsonMToInt32FromOI(productDoc, "_id"))

	created, err := s.svc.OrderItemCommand.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{
		OrderID: orderID, ProductID: productID, Quantity: 3, Price: 1200,
	})
	s.Require().NoError(err)
	s.Require().NotNil(created)
	s.NotZero(created.OrderItemID)

	updated, err := s.svc.OrderItemCommand.UpdateOrderItem(ctx, &requests.UpdateOrderItemRecordRequest{
		OrderItemID: int(created.OrderItemID), OrderID: orderID, ProductID: productID, Quantity: 9, Price: 999,
	})
	s.Require().NoError(err)
	s.Require().NotNil(updated)
	s.Equal(int32(9), updated.Quantity)
	s.Equal(int64(999), updated.Price)

	ok, err := s.svc.OrderItemCommand.DeleteOrderItem(ctx, int(created.OrderItemID))
	s.Require().NoError(err)
	s.True(ok)

	count, err := db.Collection("order_items").CountDocuments(ctx, bson.M{"_id": created.OrderItemID})
	s.Require().NoError(err)
	s.Equal(int64(0), count)
}

func (s *OrderItemServiceTestSuite) TestOrderItemLifecycle() {
	ctx := context.Background()
	db := s.ts.MongoDB

	// Find the order ID we seeded
	var orderDoc bson.M
	err := db.Collection("orders").FindOne(ctx, bson.M{"total_price": int64(5000)}).Decode(&orderDoc)
	s.Require().NoError(err)
	orderID := int(bsonMToInt32FromOI(orderDoc, "_id"))

	// 2. FindByOrder
	items, err := s.svc.OrderItemQuery.FindOrderItemByOrder(ctx, orderID)
	s.Require().NoError(err)
	s.NotEmpty(items)

	// 3. FindAll
	_, total, err := s.svc.OrderItemQuery.FindAllOrderItems(ctx, &requests.FindAllOrderItems{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 4. FindByActive
	active, _, err := s.svc.OrderItemQuery.FindByActive(ctx, &requests.FindAllOrderItems{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(active)

	// 5. Trash the order item
	var oiDoc bson.M
	err = db.Collection("order_items").FindOne(ctx, bson.M{"order_id": int32(orderID)}).Decode(&oiDoc)
	s.Require().NoError(err)
	oiID := bsonMToInt32FromOI(oiDoc, "_id")

	_, _ = db.Collection("order_items").UpdateOne(ctx, bson.M{"_id": oiID}, bson.M{"$set": bson.M{"deleted_at": time.Now().UTC()}})

	// 6. FindByTrashed
	_, totalTrashed, err := s.svc.OrderItemQuery.FindByTrashed(ctx, &requests.FindAllOrderItems{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)
}

func bsonMToInt32FromOI(d bson.M, key string) int32 {
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

func TestOrderItemServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderItemServiceTestSuite))
}
