package auth_test

import (
	"context"
	"net"
	"testing"

	auth_cache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	"github.com/MamangRust/microservice-point-of-sale-auth/handler"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type AuthPasswordResetTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	redisClient *redis.Client
	queryClient pbauth.AuthServiceClient
	cmdClient   pbauth.AuthServiceClient
	conn        *grpc.ClientConn
	grpcServer  *grpc.Server
	userBck     *grpcBackend
	roleBck     *grpcBackend
	hasher      hash.HashPassword
}

func (s *AuthPasswordResetTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextIdentity)
	s.ts = ts

	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	db := s.ts.MongoDB

	log, _ := logger.NewLogger("test", nil)
	s.hasher = hash.NewHashingPassword()
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	// Registration and credential lookups go through the user and role gRPC
	// adapters, so run the real backends in-process.
	s.roleBck, err = startRoleBackend(db, cacheStore, log, obs)
	s.Require().NoError(err)
	s.userBck, err = startUserBackend(db, pbrole.NewRoleServiceClient(s.roleBck.conn), pbuserrole.NewUserRoleServiceClient(s.roleBck.conn), cacheStore, log, obs)
	s.Require().NoError(err)

	roleClient := pbrole.NewRoleServiceClient(s.roleBck.conn)
	if _, err := roleClient.CreateRole(context.Background(), &pbrole.CreateRoleRequest{Name: "ROLE_ADMIN"}); err != nil {
		// A duplicate name from a previous run is fine.
		s.T().Logf("seed ROLE_ADMIN: %v", err)
	}

	repos := repository.NewRepositories(db, pbuser.NewUserServiceClient(s.userBck.conn), pbrole.NewRoleServiceClient(s.roleBck.conn), pbuserrole.NewUserRoleServiceClient(s.roleBck.conn))
	tokenManager, _ := auth.NewManager("mysecret")

	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      auth_cache.NewMencache(cacheStore),
		Token:         tokenManager,
		Hash:          s.hasher,
		Kafka:         nil,
		Observability: obs,
	})

	h := handler.NewAuthHandleGrpc(svc, log)
	s.grpcServer = grpc.NewServer()
	pbauth.RegisterAuthServiceServer(s.grpcServer, h)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() {
		_ = s.grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.queryClient = pbauth.NewAuthServiceClient(conn)
	s.cmdClient = pbauth.NewAuthServiceClient(conn)
}

func (s *AuthPasswordResetTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.userBck != nil {
		s.userBck.Stop()
	}
	if s.roleBck != nil {
		s.roleBck.Stop()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	s.ts.Teardown()
}

func (s *AuthPasswordResetTestSuite) registerUser(email, password string) int32 {
	ctx := context.Background()
	res, err := s.cmdClient.RegisterUser(ctx, &pbauth.RegisterRequest{
		Firstname:       "Reset",
		Lastname:        "Tester",
		Email:           email,
		Password:        password,
		ConfirmPassword: password,
	})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().NotZero(res.Data.Id)
	return res.Data.Id
}

func (s *AuthPasswordResetTestSuite) markUserVerified(userID int32) {
	_, err := s.ts.MongoDB.Collection("users").UpdateOne(s.ts.Ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"is_verified": true}},
	)
	s.Require().NoError(err)
}

func (s *AuthPasswordResetTestSuite) fetchVerificationCode(userID int32) string {
	var userDoc bson.M
	err := s.ts.MongoDB.Collection("users").FindOne(s.ts.Ctx, bson.M{"_id": userID}).Decode(&userDoc)
	s.Require().NoError(err)
	if code, ok := userDoc["verification_code"].(string); ok {
		return code
	}
	return ""
}

func (s *AuthPasswordResetTestSuite) fetchResetToken(userID int32) string {
	var tokenDoc bson.M
	err := s.ts.MongoDB.Collection("reset_tokens").FindOne(s.ts.Ctx, bson.M{"user_id": userID}).Decode(&tokenDoc)
	s.Require().NoError(err)
	if token, ok := tokenDoc["token"].(string); ok {
		return token
	}
	return ""
}

func (s *AuthPasswordResetTestSuite) storedPassword(userID int32) string {
	var userDoc bson.M
	err := s.ts.MongoDB.Collection("users").FindOne(s.ts.Ctx, bson.M{"_id": userID}).Decode(&userDoc)
	s.Require().NoError(err)
	if pw, ok := userDoc["password"].(string); ok {
		return pw
	}
	return ""
}

func (s *AuthPasswordResetTestSuite) Test1_VerifyCodeResolver_Success() {
	userID := s.registerUser("verifycode.success@example.com", "password123")
	code := s.fetchVerificationCode(userID)
	s.Require().NotEmpty(code)

	res, err := s.queryClient.VerifyCode(context.Background(), &pbauth.VerifyCodeRequest{Code: code})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)

	// Check is_verified is true
	var userDoc bson.M
	err = s.ts.MongoDB.Collection("users").FindOne(s.ts.Ctx, bson.M{"_id": userID}).Decode(&userDoc)
	s.Require().NoError(err)
	s.Equal(true, userDoc["is_verified"])
}

func (s *AuthPasswordResetTestSuite) Test2_VerifyCodeResolver_InvalidCode() {
	_, err := s.queryClient.VerifyCode(context.Background(), &pbauth.VerifyCodeRequest{Code: "does-not-exist"})
	s.Require().Error(err)
	s.Equal(codes.NotFound, status.Code(err))
}

func (s *AuthPasswordResetTestSuite) Test3_ForgotPasswordResolver_Success() {
	userID := s.registerUser("forgot.success@example.com", "password123")

	res, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "forgot.success@example.com"})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)

	s.Require().NotEmpty(s.fetchResetToken(userID))
}

func (s *AuthPasswordResetTestSuite) Test4_ForgotPasswordResolver_UnknownEmail_DoesNotRevealAccount() {
	res, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "ghost@example.com"})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)
}

func (s *AuthPasswordResetTestSuite) Test5_ResetPasswordResolver_Success_AndHashingFix() {
	const newPassword = "freshpassword456"

	userID := s.registerUser("reset.success@example.com", "password123")

	_, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "reset.success@example.com"})
	s.Require().NoError(err)

	token := s.fetchResetToken(userID)
	s.Require().NotEmpty(token)

	res, err := s.cmdClient.ResetPassword(context.Background(), &pbauth.ResetPasswordRequest{
		ResetToken:      token,
		Password:        newPassword,
		ConfirmPassword: newPassword,
	})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)

	// Hashing fix: the DB must hold a bcrypt hash, never the plaintext
	stored := s.storedPassword(userID)
	s.NotEqual(newPassword, stored, "plaintext password must never be persisted")
	s.NoError(s.hasher.ComparePassword(stored, newPassword),
		"stored password should verify against the new password via bcrypt")

	// The reset token must be consumed after a successful reset
	count, err := s.ts.MongoDB.Collection("reset_tokens").CountDocuments(s.ts.Ctx, bson.M{"user_id": userID})
	s.Require().NoError(err)
	s.Equal(int64(0), count)

	// Login with the new password must succeed after the reset
	s.markUserVerified(userID)
	loginRes, err := s.queryClient.LoginUser(context.Background(), &pbauth.LoginRequest{
		Email:    "reset.success@example.com",
		Password: newPassword,
	})
	s.Require().NoError(err)
	s.Require().NotNil(loginRes)
	s.NotEmpty(loginRes.Data.AccessToken)
}

func (s *AuthPasswordResetTestSuite) Test6_ResetPasswordResolver_InvalidToken() {
	_, err := s.cmdClient.ResetPassword(context.Background(), &pbauth.ResetPasswordRequest{
		ResetToken:      "bogus-token",
		Password:        "newpassword123",
		ConfirmPassword: "newpassword123",
	})
	s.Require().Error(err)
	s.Equal(codes.NotFound, status.Code(err))
}

func (s *AuthPasswordResetTestSuite) Test7_ResetPasswordResolver_PasswordMismatch() {
	userID := s.registerUser("reset.mismatch@example.com", "password123")

	_, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "reset.mismatch@example.com"})
	s.Require().NoError(err)

	token := s.fetchResetToken(userID)
	s.Require().NotEmpty(token)

	_, err = s.cmdClient.ResetPassword(context.Background(), &pbauth.ResetPasswordRequest{
		ResetToken:      token,
		Password:        "newpassword123",
		ConfirmPassword: "differentpassword",
	})
	s.Require().Error(err)
}

func TestAuthPasswordResetSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthPasswordResetTestSuite))
}
