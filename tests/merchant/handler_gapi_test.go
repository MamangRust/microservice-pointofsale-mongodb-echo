package merchant_test

import (
	"context"
	"testing"

	merchant_cache "github.com/MamangRust/microservice-point-of-sale-merchant/cache"
	"github.com/MamangRust/microservice-point-of-sale-merchant/handler"
	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MerchantGapiTestSuite struct {
	tests.BaseTestSuite
	client     *grpc.ClientConn
	userID     int
	merchantID int
}

func (s *MerchantGapiTestSuite) SetupSuite() {
	s.Context = tests.ContextMerchant
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	db := s.MongoDB()
	repos := repository.NewRepositories(db, pbuser.NewUserServiceClient(s.Conns["user"]))

	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	mencache := merchant_cache.NewMencache(cacheStore)

	svc := service.NewService(&service.Deps{
		Kafka:         nil,
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})

	merchantHandlers, merchantDocHandlers := handler.NewHandler(svc)
	server := grpc.NewServer()
	pbmerchant.RegisterMerchantServiceServer(server, merchantHandlers)
	_ = merchantDocHandlers

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)
	s.client = conn
	s.userID = s.SeedUser(context.Background())
}

func (s *MerchantGapiTestSuite) TestMerchantGapiLifecycle() {
	ctx := context.Background()

	cmdClient := pbmerchant.NewMerchantServiceClient(s.client)
	queryClient := pbmerchant.NewMerchantServiceClient(s.client)

	createReq := &pbmerchant.CreateMerchantRequest{
		UserId:       int32(s.userID),
		Name:         "Gapi Merchant",
		Description:  "Detailed description of the merchant.",
		Address:      "Merchant Street No. 1",
		ContactEmail: "gapi.merchant@example.com",
		ContactPhone: "08123456789",
		Status:       "active",
	}
	res, err := cmdClient.Create(ctx, createReq)
	s.NoError(err)
	s.Equal(createReq.Name, res.Data.Name)
	merchantID := res.Data.Id

	found, err := queryClient.FindById(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)
	s.Equal(merchantID, found.Data.Id)

	allRes, err := queryClient.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	activeRes, err := queryClient.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	updateReq := &pbmerchant.UpdateMerchantRequest{
		MerchantId:   merchantID,
		UserId:       int32(s.userID),
		Name:         "Gapi Merchant Updated",
		Description:  "Updated description.",
		Address:      "New Street 2",
		ContactEmail: "updated@example.com",
		ContactPhone: "08987654321",
		Status:       "waiting",
	}
	updateRes, err := cmdClient.Update(ctx, updateReq)
	s.NoError(err)
	s.Equal(updateReq.Name, updateRes.Data.Name)

	statusRes, err := cmdClient.UpdateMerchantStatus(ctx, &pbmerchant.UpdateMerchantStatusRequest{MerchantId: merchantID, Status: "active"})
	s.NoError(err)
	s.Equal("active", statusRes.Data.Status)

	_, err = cmdClient.TrashedMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	trashedRes, err := queryClient.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	_, err = cmdClient.RestoreMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	_, _ = cmdClient.TrashedMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	_, err = cmdClient.DeleteMerchantPermanent(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	_, err = cmdClient.RestoreAllMerchant(ctx, &emptypb.Empty{})
	s.NoError(err)

	_, err = cmdClient.DeleteAllMerchantPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestMerchantGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantGapiTestSuite))
}
