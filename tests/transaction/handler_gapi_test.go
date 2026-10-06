package transaction_test

import (
	"context"
	"testing"

	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	trans_cache "github.com/MamangRust/microservice-point-of-sale-transacton/cache"
	trans_handler "github.com/MamangRust/microservice-point-of-sale-transacton/handler"
	trans_repo "github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	trans_service "github.com/MamangRust/microservice-point-of-sale-transacton/service"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TransactionGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
}

func (s *TransactionGapiTestSuite) SetupSuite() {
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

	cashierClient := pbcashier.NewCashierServiceClient(s.Conns["cashier"])
	merchantClient := pbmerchant.NewMerchantServiceClient(s.Conns["merchant"])
	orderClient := pborder.NewOrderServiceClient(s.Conns["order"])
	orderItemClient := pborderitem.NewOrderItemServiceClient(s.Conns["order-item"])

	mencache := trans_cache.NewMencache(cacheStore)
	repos := trans_repo.NewRepositories(db, cashierClient, merchantClient, orderClient, orderItemClient)
	svc := trans_service.NewService(&trans_service.Deps{
		Kafka:         nil,
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	handlers := trans_handler.NewHandler(svc, s.Log)
	server := grpc.NewServer()
	pbtransaction.RegisterTransactionServiceServer(server, handlers)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)
	s.client = conn
}

func (s *TransactionGapiTestSuite) TestTransactionGapiLifecycle() {
	ctx := context.Background()

	userID := s.SeedUser(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	catID := s.SeedCategory(ctx)
	prodID := s.SeedProduct(ctx, merchID, catID)
	orderID := s.SeedOrder(ctx, userID, merchID, prodID)

	var cashierDoc bson.M
	err := s.Coll("cashiers").FindOne(ctx, bson.M{"user_id": int32(userID), "merchant_id": int32(merchID)}).Decode(&cashierDoc)
	s.Require().NoError(err)
	cashierID := int(bsonMToInt32FromTx(cashierDoc, "_id"))

	cmdClient := pbtransaction.NewTransactionServiceClient(s.client)
	queryClient := pbtransaction.NewTransactionServiceClient(s.client)

	createRes, err := cmdClient.Create(ctx, &pbtransaction.CreateTransactionRequest{
		OrderId:       int32(orderID),
		CashierId:     int32(cashierID),
		PaymentMethod: "E-Wallet",
		PaymentStatus: "pending",
		Amount:        100000,
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	transID := createRes.Data.Id

	getRes, err := queryClient.FindById(ctx, &pbtransaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)
	s.Equal("E-Wallet", getRes.Data.PaymentMethod)

	allRes, err := queryClient.FindAll(ctx, &pbtransaction.FindAllTransactionRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	activeRes, err := queryClient.FindByActive(ctx, &pbtransaction.FindAllTransactionRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	updateRes, err := cmdClient.Update(ctx, &pbtransaction.UpdateTransactionRequest{
		TransactionId: transID,
		OrderId:       int32(orderID),
		CashierId:     int32(cashierID),
		PaymentMethod: "Credit Card",
		Amount:        150000,
	})
	s.Require().NoError(err)
	s.Equal("Credit Card", updateRes.Data.PaymentMethod)

	_, err = cmdClient.TrashedTransaction(ctx, &pbtransaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)

	trashedRes, err := queryClient.FindByTrashed(ctx, &pbtransaction.FindAllTransactionRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	_, err = cmdClient.RestoreTransaction(ctx, &pbtransaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)

	_, _ = cmdClient.TrashedTransaction(ctx, &pbtransaction.FindByIdTransactionRequest{Id: transID})
	_, err = cmdClient.DeleteTransactionPermanent(ctx, &pbtransaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)

	_, err = cmdClient.RestoreAllTransaction(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	_, err = cmdClient.DeleteAllTransactionPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestTransactionGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionGapiTestSuite))
}
