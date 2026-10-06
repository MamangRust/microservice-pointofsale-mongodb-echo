package role_test

import (
	"context"
	"testing"

	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	"github.com/MamangRust/microservice-point-of-sale-role/repository"
	"github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
}

func (s *RoleGapiTestSuite) SetupSuite() {
	s.Context = tests.ContextIdentity
	s.BaseTestSuite.SetupSuite()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	db := s.MongoDB()

	// Role dependencies
	mencache := role_cache.NewMencache(cacheStore)
	repos := repository.NewRepositories(db)
	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})

	// Handler
	handlers := role_handler.NewHandler(svc)

	// Server
	server := grpc.NewServer()
	pbrole.RegisterRoleServiceServer(server, handlers)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = conn
}

func (s *RoleGapiTestSuite) TestRoleGapiLifecycle() {
	ctx := context.Background()

	cmdClient := pbrole.NewRoleServiceClient(s.client)
	queryClient := pbrole.NewRoleServiceClient(s.client)

	// 1. Create
	createRes, err := cmdClient.CreateRole(ctx, &pbrole.CreateRoleRequest{
		Name: "Gapi Role",
	})
	s.Require().NotNil(createRes)
	roleID := createRes.Data.Id

	// 2. FindById
	getRes, err := queryClient.FindByIdRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)
	s.Equal("Gapi Role", getRes.Data.Name)

	// 3. FindAll
	allRes, err := queryClient.FindAllRole(ctx, &pbrole.FindAllRoleRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := queryClient.FindByActive(ctx, &pbrole.FindAllRoleRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, err := cmdClient.UpdateRole(ctx, &pbrole.UpdateRoleRequest{
		Id:   roleID,
		Name: "Gapi Role Updated",
	})
	s.Require().NoError(err)
	s.Equal("Gapi Role Updated", updateRes.Data.Name)

	// 6. Trash
	_, err = cmdClient.TrashedRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)

	// 7. FindByTrashed
	trashedRes, err := queryClient.FindByTrashed(ctx, &pbrole.FindAllRoleRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = cmdClient.RestoreRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, _ = cmdClient.TrashedRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: roleID})
	_, err = cmdClient.DeleteRolePermanent(ctx, &pbrole.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)

	// 10. RestoreAll
	_, err = cmdClient.RestoreAllRole(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 11. DeleteAll
	_, err = cmdClient.DeleteAllRolePermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestRoleGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleGapiTestSuite))
}
