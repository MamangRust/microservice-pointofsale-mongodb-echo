package auth_test

import (
	"context"
	"net"
	"testing"

	mencache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
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
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandlerGapiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	redisClient *redis.Client
	queryClient pbauth.AuthServiceClient
	cmdClient   pbauth.AuthServiceClient
	conn        *grpc.ClientConn
	grpcServer  *grpc.Server
	userBck     *grpcBackend
	roleBck     *grpcBackend
	roleClient  pbrole.RoleServiceClient
	email       string
	password    string
	accessToken string
}

func (s *AuthHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextIdentity)
	s.ts = ts

	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	db := s.ts.MongoDB

	log, _ := logger.NewLogger("test", nil)
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)
	obs, _ := observability.NewObservability("test", log)

	// Run the real user and role backends in-process; auth reaches them only
	// through the pkg/adapter gRPC path.
	s.roleBck, err = startRoleBackend(db, cacheStore, log, obs)
	s.Require().NoError(err)
	s.userBck, err = startUserBackend(db, pbrole.NewRoleServiceClient(s.roleBck.conn), pbuserrole.NewUserRoleServiceClient(s.roleBck.conn), cacheStore, log, obs)
	s.Require().NoError(err)

	s.roleClient = pbrole.NewRoleServiceClient(s.roleBck.conn)

	// 3. Setup Auth Service
	repos := repository.NewRepositories(db, pbuser.NewUserServiceClient(s.userBck.conn), pbrole.NewRoleServiceClient(s.roleBck.conn), pbuserrole.NewUserRoleServiceClient(s.roleBck.conn))

	tokenManager, _ := auth.NewManager("mysecret")
	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache.NewMencache(cacheStore),
		Token:         tokenManager,
		Hash:          hasher,
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

	s.email = "auth.handler.gapi.test@example.com"
	s.password = "password123"

	// Seed ROLE_ADMIN through the role service itself so the role id is int32.
	if _, err := s.roleClient.CreateRole(s.ts.Ctx, &pbrole.CreateRoleRequest{Name: "ROLE_ADMIN"}); err != nil {
		s.T().Logf("seed ROLE_ADMIN: %v", err)
	}
}

func (s *AuthHandlerGapiTestSuite) TearDownSuite() {
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

func (s *AuthHandlerGapiTestSuite) verifyUser(email string) {
	usersColl := s.ts.MongoDB.Collection("users")
	_, err := usersColl.UpdateOne(s.ts.Ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"is_verified": true}},
	)
	s.Require().NoError(err)
}

func (s *AuthHandlerGapiTestSuite) Test1_Register() {
	ctx := context.Background()
	req := &pbauth.RegisterRequest{
		Firstname:       "Auth",
		Lastname:        "Handler",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	}

	res, err := s.cmdClient.RegisterUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func (s *AuthHandlerGapiTestSuite) Test2_Login() {
	ctx := context.Background()

	s.verifyUser(s.email)

	req := &pbauth.LoginRequest{
		Email:    s.email,
		Password: s.password,
	}

	res, err := s.queryClient.LoginUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.NotEmpty(res.Data.AccessToken)
	s.accessToken = res.Data.AccessToken
}

func (s *AuthHandlerGapiTestSuite) Test4_LoginLockout() {
	ctx := context.Background()
	email := "locked.gapi@example.com"
	password := "wrongpassword"

	regReq := &pbauth.RegisterRequest{
		Firstname:       "Locked",
		Lastname:        "Gapi",
		Email:           email,
		Password:        "correctpassword",
		ConfirmPassword: "correctpassword",
	}
	_, err := s.cmdClient.RegisterUser(ctx, regReq)
	s.NoError(err)
	s.verifyUser(email)

	loginReq := &pbauth.LoginRequest{
		Email:    email,
		Password: password,
	}

	for i := 0; i < 5; i++ {
		_, err := s.queryClient.LoginUser(ctx, loginReq)
		s.Error(err)
	}

	_, err = s.queryClient.LoginUser(ctx, loginReq)
	s.Error(err)
	s.Contains(err.Error(), "Account is locked")
}

func (s *AuthHandlerGapiTestSuite) Test3_GetMe() {
	s.Require().NotEmpty(s.accessToken)
	ctx := context.Background()

	res, err := s.queryClient.GetMe(ctx, &pbauth.GetMeRequest{AccessToken: s.accessToken})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func TestAuthHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerGapiTestSuite))
}
