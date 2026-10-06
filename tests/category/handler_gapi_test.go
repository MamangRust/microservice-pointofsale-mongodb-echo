package category_test

import (
	"context"
	"testing"

	cat_cache "github.com/MamangRust/microservice-point-of-sale-category/cache"
	cat_handler "github.com/MamangRust/microservice-point-of-sale-category/handler"
	cat_repo "github.com/MamangRust/microservice-point-of-sale-category/repository"
	cat_service "github.com/MamangRust/microservice-point-of-sale-category/service"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CategoryGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
}

func (s *CategoryGapiTestSuite) SetupSuite() {
	s.Context = tests.ContextCatalog
	s.BaseTestSuite.SetupSuite()

	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	db := s.MongoDB()

	mencache := cat_cache.NewMencache(cacheStore)
	repos := cat_repo.NewRepositories(db)
	svc := cat_service.NewService(&cat_service.Deps{
		Ctx:           context.Background(),
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	handlers := cat_handler.NewHandler(svc)

	server := grpc.NewServer()
	pbcategory.RegisterCategoryServiceServer(server, handlers)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)
	s.client = conn
}

func (s *CategoryGapiTestSuite) TestCategoryGapiLifecycle() {
	ctx := context.Background()

	cmdClient := pbcategory.NewCategoryServiceClient(s.client)
	queryClient := pbcategory.NewCategoryServiceClient(s.client)

	// 1. Create
	createRes, err := cmdClient.Create(ctx, &pbcategory.CreateCategoryRequest{
		Name:        "GAPI Category",
		Description: "Testing via GRPC",
	})
	s.NoError(err)
	s.NotNil(createRes)
	catID := createRes.Data.Id

	// 2. FindById
	getRes, err := queryClient.FindById(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)
	s.Equal("GAPI Category", getRes.Data.Name)

	// 3. FindAll
	allRes, err := queryClient.FindAll(ctx, &pbcategory.FindAllCategoryRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := queryClient.FindByActive(ctx, &pbcategory.FindAllCategoryRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, err := cmdClient.Update(ctx, &pbcategory.UpdateCategoryRequest{
		CategoryId:  catID,
		Name:        "GAPI Category Updated",
		Description: "Updated via GRPC",
	})
	s.NoError(err)
	s.Equal("GAPI Category Updated", updateRes.Data.Name)

	// 6. Trash
	_, err = cmdClient.TrashedCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)

	// 7. FindByTrashed
	trashedRes, err := queryClient.FindByTrashed(ctx, &pbcategory.FindAllCategoryRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = cmdClient.RestoreCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)

	// 9. DeletePermanent
	_, _ = cmdClient.TrashedCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	_, err = cmdClient.DeleteCategoryPermanent(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)

	// 10. RestoreAll
	_, err = cmdClient.RestoreAllCategory(ctx, &emptypb.Empty{})
	s.NoError(err)

	// 11. DeleteAll
	_, err = cmdClient.DeleteAllCategoryPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestCategoryGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CategoryGapiTestSuite))
}
