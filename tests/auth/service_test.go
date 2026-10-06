package auth_test

import (
	"context"
	"testing"

	mencache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	authService *service.Service
	userBck     *grpcBackend
	roleBck     *grpcBackend
	redisCli    *redis.Client
	roleClient  pbrole.RoleServiceClient
	email       string
	password    string
}

func (s *AuthServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextIdentity)
	s.ts = ts

	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisCli = redis.NewClient(opts)

	db := s.ts.MongoDB

	log, _ := logger.NewLogger("test", nil)
	hasher := hash.NewHashingPassword()
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisCli, log, cacheMetrics)
	mencacheService := mencache.NewMencache(cacheStore)

	s.roleBck, err = startRoleBackend(db, cacheStore, log, obs)
	s.Require().NoError(err)
	s.userBck, err = startUserBackend(db, pbrole.NewRoleServiceClient(s.roleBck.conn), pbuserrole.NewUserRoleServiceClient(s.roleBck.conn), cacheStore, log, obs)
	s.Require().NoError(err)

	s.roleClient = pbrole.NewRoleServiceClient(s.roleBck.conn)

	repos := repository.NewRepositories(db, pbuser.NewUserServiceClient(s.userBck.conn), pbrole.NewRoleServiceClient(s.roleBck.conn), pbuserrole.NewUserRoleServiceClient(s.roleBck.conn))

	tokenManager, _ := auth.NewManager("mysecretkey")

	s.authService = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencacheService,
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: obs,
	})

	s.email = "auth.service.test@example.com"
	s.password = "password123"

	// Seed ROLE_ADMIN through the role service itself so the role id is int32.
	if _, err := s.roleClient.CreateRole(context.Background(), &pbrole.CreateRoleRequest{Name: "ROLE_ADMIN"}); err != nil {
		// A duplicate name from a previous run is fine.
		s.T().Logf("seed ROLE_ADMIN: %v", err)
	}
}

func (s *AuthServiceTestSuite) TearDownSuite() {
	if s.userBck != nil {
		s.userBck.Stop()
	}
	if s.roleBck != nil {
		s.roleBck.Stop()
	}
	if s.redisCli != nil {
		s.redisCli.Close()
	}
	s.ts.Teardown()
}

func (s *AuthServiceTestSuite) TestAuthLifecycle() {
	ctx := context.Background()

	// 1. Register
	regReq := &requests.RegisterRequest{
		FirstName:       "Auth",
		LastName:        "Service",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	}

	created, err := s.authService.Register.Register(ctx, regReq)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	s.Equal(s.email, created.Email)

	// 1b. Verify email (login hanya menerima user is_verified = true)
	var verifyCode string
	err = s.ts.MongoDB.Collection("users").FindOne(ctx, bson.M{"email": s.email}).Decode(
		&struct {
			VerificationCode string `bson:"verification_code"`
		}{},
	)
	// Fetch verification code
	var userDoc bson.M
	err = s.ts.MongoDB.Collection("Users").FindOne(ctx, bson.M{"email": s.email}).Decode(&userDoc)
	if err != nil {
		// Try lowercase
		err = s.ts.MongoDB.Collection("users").FindOne(ctx, bson.M{"email": s.email}).Decode(&userDoc)
	}
	s.Require().NoError(err)
	if vc, ok := userDoc["verification_code"].(string); ok {
		verifyCode = vc
	}

	verified, err := s.authService.PasswordReset.VerifyCode(ctx, verifyCode)
	s.Require().NoError(err)
	s.True(verified)

	// 2. Login
	loginReq := &requests.AuthRequest{
		Email:    s.email,
		Password: s.password,
	}

	tokenRes, err := s.authService.Login.Login(ctx, loginReq)
	s.Require().NoError(err)
	s.Require().NotNil(tokenRes)
	s.NotEmpty(tokenRes.AccessToken)
	s.NotEmpty(tokenRes.RefreshToken)

	// 3. ForgotPassword
	success, err := s.authService.PasswordReset.ForgotPassword(ctx, s.email)
	s.Require().NoError(err)
	s.True(success)
}

func TestAuthServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthServiceTestSuite))
}
