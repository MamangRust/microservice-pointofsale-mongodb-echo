package order_test

import (
	"context"
	"testing"
	"time"

	order_cache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	order_handler "github.com/MamangRust/microservice-point-of-sale-order/handler"
	order_repo "github.com/MamangRust/microservice-point-of-sale-order/repository"
	order_service "github.com/MamangRust/microservice-point-of-sale-order/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
}

func (s *OrderGapiTestSuite) SetupSuite() {
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()

	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupCashierService()
	s.SetupTransactionService()

	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	db := s.MongoDB()

	mencache := order_cache.NewMencache(cacheStore)
	repos := order_repo.NewRepositories(
		db,
		pbcashier.NewCashierServiceClient(s.Conns["cashier"]),
		pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]),
		pbproduct.NewProductServiceClient(s.Conns["product"]),
		pborderitem.NewOrderItemServiceClient(s.Conns["order-item"]),
	)
	svc := order_service.NewService(&order_service.Deps{
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	handlers := order_handler.NewHandler(svc)
	server := grpc.NewServer()
	pborder.RegisterOrderServiceServer(server, handlers)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)
	s.client = conn
}

func (s *OrderGapiTestSuite) TestOrderGapiLifecycle() {
	ctx := context.Background()

	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)

	// Seed a cashier (cashiers belong to the merchant context)
	now := time.Now().UTC()
	cashierID, _ := database.GetNextID(ctx, s.DBForCollection("cashiers"), "cashiers")
	s.Coll("cashiers").InsertOne(ctx, bson.M{"_id": cashierID, "merchant_id": merchID, "user_id": userID, "name": "Order Gapi Cashier", "created_at": now, "updated_at": now})

	cmdClient := pborder.NewOrderServiceClient(s.client)
	queryClient := pborder.NewOrderServiceClient(s.client)

	createRes, err := cmdClient.Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(merchID),
		CashierId:  int32(cashierID),
		Items: []*pborder.CreateOrderItemRequest{
			{ProductId: int32(prodID), Quantity: 1},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	orderID := createRes.Data.Id

	getRes, err := queryClient.FindById(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	s.Equal(int32(userID), getRes.Data.CashierId)

	allRes, err := queryClient.FindAll(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	activeRes, err := queryClient.FindByActive(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	itemClient := pborderitem.NewOrderItemServiceClient(s.Conns["order-item"])
	itemsRes, err := itemClient.FindOrderItemByOrder(ctx, &pborderitem.FindByIdOrderItemRequest{Id: orderID})
	s.Require().NoError(err)
	s.NotEmpty(itemsRes.Data)
	orderItemID := itemsRes.Data[0].Id

	_, err = cmdClient.Update(ctx, &pborder.UpdateOrderRequest{
		OrderId: orderID,
		Items: []*pborder.UpdateOrderItemRequest{
			{OrderItemId: orderItemID, ProductId: int32(prodID), Quantity: 1},
		},
	})
	s.Require().NoError(err)

	_, err = cmdClient.TrashedOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	trashedRes, err := queryClient.FindByTrashed(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	_, err = cmdClient.RestoreOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	_, _ = cmdClient.TrashedOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	_, err = cmdClient.DeleteOrderPermanent(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	_, err = cmdClient.RestoreAllOrder(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	_, err = cmdClient.DeleteAllOrderPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestOrderGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderGapiTestSuite))
}
