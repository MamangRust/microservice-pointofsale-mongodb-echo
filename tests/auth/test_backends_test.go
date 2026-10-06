package auth_test

import (
	"net"

	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	role_repo "github.com/MamangRust/microservice-point-of-sale-role/repository"
	role_service "github.com/MamangRust/microservice-point-of-sale-role/service"
	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	user_handler "github.com/MamangRust/microservice-point-of-sale-user/handler"
	user_repo "github.com/MamangRust/microservice-point-of-sale-user/repository"
	user_service "github.com/MamangRust/microservice-point-of-sale-user/service"
)

// grpcBackend is an in-process gRPC server plus a client conn to it.
type grpcBackend struct {
	server *grpc.Server
	conn   *grpc.ClientConn
}

func (b *grpcBackend) Stop() {
	if b.conn != nil {
		b.conn.Close()
	}
	if b.server != nil {
		b.server.Stop()
	}
}

// startUserBackend runs the real user-service gRPC handler over an in-process
// server so the auth suites exercise the same adapter path as production. The
// user service resolves the default role through the role gRPC adapter, so a
// role client must be provided (start the role backend first).
func startUserBackend(db *mongo.Database, roleClient pbrole.RoleServiceClient, userRoleClient pbuserrole.UserRoleServiceClient, cacheStore *cache.CacheStore, log logger.LoggerInterface, obs observability.TraceLoggerObservability) (*grpcBackend, error) {
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories: user_repo.NewRepositories(&user_repo.Deps{
			Db:              db,
			RoleQueryClient: roleClient,
			UserRoleClient:  userRoleClient,
		}),
		Logger:        log,
		Hash:          hash.NewHashingPassword(),
		Mencache:      user_cache.NewMencache(cacheStore),
		Observability: obs,
	})
	srv := grpc.NewServer()
	pbuser.RegisterUserServiceServer(srv, user_handler.NewHandler(userSvc))

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		srv.Stop()
		return nil, err
	}
	return &grpcBackend{server: srv, conn: conn}, nil
}

// startRoleBackend runs the real role-service gRPC handler over an in-process
// server. The ROLE_ADMIN role expected by registration is seeded through the
// service itself (not direct Mongo inserts) so ids stay int32.
func startRoleBackend(db *mongo.Database, cacheStore *cache.CacheStore, log logger.LoggerInterface, obs observability.TraceLoggerObservability) (*grpcBackend, error) {
	roleSvc := role_service.NewService(&role_service.Deps{
		Repositories:  role_repo.NewRepositories(db),
		Logger:        log,
		Mencache:      role_cache.NewMencache(cacheStore),
		Observability: obs,
	})
	srv := grpc.NewServer()
	roleGapi := role_handler.NewHandler(roleSvc)
	pbrole.RegisterRoleServiceServer(srv, roleGapi)
	pbuserrole.RegisterUserRoleServiceServer(srv, roleGapi)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		srv.Stop()
		return nil, err
	}
	return &grpcBackend{server: srv, conn: conn}, nil
}
