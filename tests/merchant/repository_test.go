package merchant_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"net"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MerchantRepositoryTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       *repository.Repositories
	merchantID int
	userID     int
}

func (s *MerchantRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextMerchant)
	s.ts = ts
	s.Require().NoError(err)

	userLis, _ := net.Listen("tcp", "localhost:0")
	userConn, _ := grpc.NewClient(userLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	db := s.ts.MongoDB
	s.repo = repository.NewRepositories(db, pbuser.NewUserServiceClient(userConn))

	ctx := context.Background()
	now := time.Now().UTC()
	uid, _ := database.GetNextID(ctx, s.ts.DBForCollection("users"), "users")
	s.userID = int(uid)
	s.ts.Coll("users").InsertOne(ctx, bson.M{"_id": uid, "firstname": "Merchant", "lastname": "RepoTest", "email": "merchant.repo@example.com", "password": "password123", "verification_code": "test-verify", "is_verified": true, "created_at": now, "updated_at": now})
}

func (s *MerchantRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *MerchantRepositoryTestSuite) Test1_CreateMerchant() {
	ctx := context.Background()
	req := &requests.CreateMerchantRequest{UserID: s.userID, Name: "Test Merchant", Description: "Detailed description", Address: "Merchant Street No. 1", ContactEmail: "merchant@example.com", ContactPhone: "08123456789", Status: "active"}
	res, err := s.repo.MerchantCommand.CreateMerchant(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Name, res.Name)
	s.Equal(int32(s.userID), res.UserID)
	s.merchantID = int(res.MerchantID)
}

func (s *MerchantRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()
	found, err := s.repo.MerchantQuery.FindById(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.merchantID, int(found.MerchantID))
}

func (s *MerchantRepositoryTestSuite) Test3_UpdateMerchant() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()
	req := &requests.UpdateMerchantRequest{MerchantID: &s.merchantID, UserID: s.userID, Name: "Updated Merchant", Description: "Updated description", Address: "Updated Street", ContactEmail: "updated@example.com", ContactPhone: "08987654321", Status: "active"}
	res, err := s.repo.MerchantCommand.UpdateMerchant(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Merchant", res.Name)
}

func (s *MerchantRepositoryTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()
	trashed, err := s.repo.MerchantCommand.TrashedMerchant(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(trashed)
	restored, err := s.repo.MerchantCommand.RestoreMerchant(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(restored)
	found, err := s.repo.MerchantQuery.FindById(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(found)
}

func (s *MerchantRepositoryTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()
	_, err := s.repo.MerchantCommand.TrashedMerchant(ctx, s.merchantID)
	s.NoError(err)
	success, err := s.repo.MerchantCommand.DeleteMerchantPermanent(ctx, s.merchantID)
	s.NoError(err)
	s.True(success)
}

func TestMerchantRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantRepositoryTestSuite))
}
