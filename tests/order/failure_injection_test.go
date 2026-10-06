package order_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	"github.com/MamangRust/microservice-point-of-sale-order/handler"
	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-order/service"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type faultMerchantServer struct {
	pbmerchant.MerchantServiceServer
	calls int32
}

func (f *faultMerchantServer) FindById(ctx context.Context, req *pbmerchant.FindByIdMerchantRequest) (*pbmerchant.ApiResponseMerchant, error) {
	atomic.AddInt32(&f.calls, 1)
	return nil, status.Error(codes.Unavailable, "merchant service unavailable (injected)")
}

func (f *faultMerchantServer) callCount() int {
	return int(atomic.LoadInt32(&f.calls))
}

type OrderFailureInjectionTestSuite struct {
	tests.BaseTestSuite
	merchantServer *faultMerchantServer
	svc            *service.Service
	client         *grpc.ClientConn
	orderID        int
	cashierID      int
}

func (s *OrderFailureInjectionTestSuite) SetupSuite() {
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()

	db := s.MongoDB()
	ctx := context.Background()
	now := time.Now().UTC()

	userID, _ := database.GetNextID(ctx, s.DBForCollection("users"), "users")
	s.Coll("users").InsertOne(ctx, bson.M{"_id": userID, "firstname": "Fail", "lastname": "Inject", "email": "fail.inject@example.com", "password": "password123", "verification_code": "test-verify", "is_verified": true, "created_at": now, "updated_at": now})

	mid, _ := database.GetNextID(ctx, s.DBForCollection("merchants"), "merchants")
	s.orderID = int(mid)
	s.Coll("merchants").InsertOne(ctx, bson.M{"_id": mid, "user_id": userID, "name": "Failure Inject Merchant", "description": "Desc", "address": "Addr", "contact_email": "fi@example.com", "contact_phone": "123", "status": "active", "created_at": now, "updated_at": now})

	cid, _ := database.GetNextID(ctx, s.DBForCollection("cashiers"), "cashiers")
	s.cashierID = int(cid)
	s.Coll("cashiers").InsertOne(ctx, bson.M{"_id": cid, "merchant_id": mid, "user_id": userID, "name": "Failure Cashier", "created_at": now, "updated_at": now})

	s.merchantServer = &faultMerchantServer{}
	merchantGRPC := grpc.NewServer()
	pbmerchant.RegisterMerchantServiceServer(merchantGRPC, s.merchantServer)
	addr, err := tests.RunGRPCServer(merchantGRPC)
	s.Require().NoError(err)
	merchantConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Servers = append(s.Servers, merchantGRPC)
	s.Conns["merchant"] = merchantConn

	merchantGuard := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)

	orderItemRepo := &stubOrderItemRepo{}
	repos := &repository.Repositories{
		CashierQuery:     &stubCashierRepo{cashierID: s.cashierID},
		MerchantQuery:    merchantadapter.New(pbmerchant.NewMerchantServiceClient(merchantConn), adapter.WithDependencyGuard(merchantGuard)),
		ProductQuery:     &stubProductRepo{},
		ProductCommand:   &stubProductCommandRepo{},
		OrderQuery:       repository.NewOrderQueryRepository(db),
		OrderCommand:     repository.NewOrderCommandRepository(db),
		OrderItemQuery:   orderItemRepo,
		OrderItemCommand: orderItemRepo,
	}

	mencacheObj := mencache.NewMencache(s.GetCacheStore())
	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencacheObj,
		Observability: s.Obs,
	})

	orderGapi := handler.NewHandler(s.svc)
	orderGRPC := grpc.NewServer()
	pborder.RegisterOrderServiceServer(orderGRPC, orderGapi)
	orderAddr, err := tests.RunGRPCServer(orderGRPC)
	s.Require().NoError(err)
	s.Servers = append(s.Servers, orderGRPC)
	orderConn, err := grpc.NewClient(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["order"] = orderConn
	s.client = orderConn
}

func (s *OrderFailureInjectionTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *OrderFailureInjectionTestSuite) TestF1_InvalidRequest_RejectedWith400_BeforeRemoteCalls() {
	ctx := context.Background()
	orderCmdClient := pborder.NewOrderServiceClient(s.client)
	_, err := orderCmdClient.Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(s.orderID),
		CashierId:  int32(s.cashierID),
	})
	s.Require().Error(err, "empty items must be rejected")
	s.Equal(codes.InvalidArgument, status.Code(err))
	s.Zero(s.merchantServer.callCount(), "no remote call may happen before validation")
}

func (s *OrderFailureInjectionTestSuite) TestF2_DependencyDown_OpensCircuit_ThenFailsFast() {
	ctx := context.Background()
	req := &requests.CreateOrderRequest{
		MerchantID: s.orderID,
		CashierID:  s.cashierID,
		Items: []requests.CreateOrderItemRequest{
			{ProductID: 1, Quantity: 1},
		},
	}

	baseline := s.merchantServer.callCount()

	for i := 0; i < 5; i++ {
		_, err := s.svc.OrderCommand.CreateOrder(ctx, req)
		s.Require().Error(err, "dependency down must fail the create (%d)", i+1)
	}
	s.Equal(baseline+5, s.merchantServer.callCount(), "all 5 failures must reach the dependency")

	afterOpen := s.merchantServer.callCount()
	_, err := s.svc.OrderCommand.CreateOrder(ctx, req)
	s.Require().Error(err, "open circuit must fail fast with an error")
	s.Equal(afterOpen, s.merchantServer.callCount(), "fail-fast must not reach the dependency")
}

func TestOrderFailureInjectionSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderFailureInjectionTestSuite))
}
