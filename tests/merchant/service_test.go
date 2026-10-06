package merchant_test

import (
	"context"
	"testing"

	merchant_cache "github.com/MamangRust/microservice-point-of-sale-merchant/cache"
	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MerchantServiceTestSuite struct {
	tests.BaseTestSuite
	merchantService *service.Service
	merchantID      int
	userID          int
}

func (s *MerchantServiceTestSuite) SetupSuite() {
	s.Context = tests.ContextMerchant
	s.BaseTestSuite.SetupSuite()

	db := s.MongoDB()
	s.SetupUserService()

	ctx := context.Background()

	uid, _ := database.GetNextID(ctx, s.DBForCollection("users"), "users")
	s.userID = int(uid)
	s.Coll("users").InsertOne(ctx, bson.M{"_id": uid, "firstname": "Merchant", "lastname": "ServiceTest", "email": "merchant.svc@example.com", "password": "password123", "verification_code": "test-verify", "is_verified": true, "created_at": bson.DateTime(0), "updated_at": bson.DateTime(0)})

	mencache := merchant_cache.NewMencache(s.GetCacheStore())
	userClient := pbuser.NewUserServiceClient(s.Conns["user"])
	repos := repository.NewRepositories(db, userClient)

	s.merchantService = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Kafka:         nil,
		Observability: s.Obs,
	})
}

func (s *MerchantServiceTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *MerchantServiceTestSuite) TestMerchantLifecycle() {
	ctx := context.Background()

	req := &requests.CreateMerchantRequest{
		UserID:       s.userID,
		Name:         "Service Merchant",
		Description:  "Merchant created via service layer",
		Address:      "Service Street No. 1",
		ContactEmail: "service.merchant@example.com",
		ContactPhone: "08123456781",
		Status:       "active",
	}
	created, err := s.merchantService.MerchantCommand.CreateMerchant(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	s.Equal(req.Name, created.Name)
	merchantID := int(created.MerchantID)

	found, err := s.merchantService.MerchantQuery.FindById(ctx, merchantID)
	s.Require().NoError(err)
	s.Equal(merchantID, int(found.MerchantID))

	updateReq := &requests.UpdateMerchantRequest{
		MerchantID:   &merchantID,
		UserID:       s.userID,
		Name:         "Updated Service Merchant",
		Description:  "Updated description",
		Address:      "Updated Street",
		ContactEmail: "updated.merchant@example.com",
		ContactPhone: "08987654321",
		Status:       "active",
	}
	updated, err := s.merchantService.MerchantCommand.UpdateMerchant(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(updateReq.Name, updated.Name)

	_, total, err := s.merchantService.MerchantQuery.FindAll(ctx, &requests.FindAllMerchants{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	_, err = s.merchantService.MerchantCommand.TrashedMerchant(ctx, merchantID)
	s.Require().NoError(err)

	_, totalTrashed, err := s.merchantService.MerchantQuery.FindByTrashed(ctx, &requests.FindAllMerchants{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	active, _, err := s.merchantService.MerchantQuery.FindByActive(ctx, &requests.FindAllMerchants{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, m := range active {
		s.NotEqual(merchantID, int(m.MerchantID))
	}

	_, err = s.merchantService.MerchantCommand.RestoreMerchant(ctx, merchantID)
	s.Require().NoError(err)

	_, err = s.merchantService.MerchantCommand.TrashedMerchant(ctx, merchantID)
	s.Require().NoError(err)
	success, err := s.merchantService.MerchantCommand.DeleteMerchantPermanent(ctx, merchantID)
	s.Require().NoError(err)
	s.True(success)

	m1, _ := s.merchantService.MerchantCommand.CreateMerchant(ctx, &requests.CreateMerchantRequest{UserID: s.userID, Name: "M1", Description: "D1", Address: "A1", ContactEmail: "m1@example.com", ContactPhone: "111", Status: "active"})
	m2, _ := s.merchantService.MerchantCommand.CreateMerchant(ctx, &requests.CreateMerchantRequest{UserID: s.userID, Name: "M2", Description: "D2", Address: "A2", ContactEmail: "m2@example.com", ContactPhone: "222", Status: "active"})

	s.merchantService.MerchantCommand.TrashedMerchant(ctx, int(m1.MerchantID))
	s.merchantService.MerchantCommand.TrashedMerchant(ctx, int(m2.MerchantID))

	resRestoreAll, err := s.merchantService.MerchantCommand.RestoreAllMerchant(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.merchantService.MerchantCommand.TrashedMerchant(ctx, int(m1.MerchantID))
	s.merchantService.MerchantCommand.TrashedMerchant(ctx, int(m2.MerchantID))

	resDeleteAll, err := s.merchantService.MerchantCommand.DeleteAllMerchantPermanent(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestMerchantServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantServiceTestSuite))
}
