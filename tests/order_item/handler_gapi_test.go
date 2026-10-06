package order_item_test

import (
	"context"
	"testing"
	"time"

	item_cache "github.com/MamangRust/microservice-point-of-sale-order-item/cache"
	item_handler "github.com/MamangRust/microservice-point-of-sale-order-item/handler"
	item_repo "github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	item_service "github.com/MamangRust/microservice-point-of-sale-order-item/service"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
)

type OrderItemGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
}

func (s *OrderItemGapiTestSuite) SetupSuite() {
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()

	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupOrderService()

	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	db := s.MongoDB()

	mencache := item_cache.NewMencache(cacheStore)
	repos := item_repo.NewRepositories(db)
	svc := item_service.NewService(&item_service.Deps{
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	handlers := item_handler.NewHandler(svc, s.Log)
	server := grpc.NewServer()
	pborderitem.RegisterOrderItemServiceServer(server, handlers)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)
	s.client = conn
}

func (s *OrderItemGapiTestSuite) TestOrderItemGapiLifecycle() {
	ctx := context.Background()

	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	productID := s.SeedProduct(ctx, merchantID, catID)
	orderID := s.SeedOrder(ctx, userID, merchantID, productID)

	// Create an order item directly in DB
	db := s.MongoDB()
	now := time.Now().UTC()
	orderItemID, _ := database.GetNextID(ctx, db, "order_items")
	db.Collection("order_items").InsertOne(ctx, bson.M{"_id": orderItemID, "order_id": int32(orderID), "product_id": int32(productID), "quantity": int32(10), "price": int32(700), "created_at": now, "updated_at": now})

	queryClient := pborderitem.NewOrderItemServiceClient(s.client)

	findByOrderRes, err := queryClient.FindOrderItemByOrder(ctx, &pborderitem.FindByIdOrderItemRequest{Id: int32(orderID)})
	s.Require().NoError(err)
	s.NotEmpty(findByOrderRes.Data)

	allRes, err := queryClient.FindAll(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	activeRes, err := queryClient.FindByActive(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// Trash the order item
	_, _ = db.Collection("order_items").UpdateOne(ctx, bson.M{"_id": int32(orderItemID)}, bson.M{"$set": bson.M{"deleted_at": now}})

	trashedRes, err := queryClient.FindByTrashed(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	_, err = queryClient.FindAll(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
}

func TestOrderItemGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderItemGapiTestSuite))
}
