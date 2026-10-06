package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	prod_cache "github.com/MamangRust/microservice-point-of-sale-product/cache"
	"github.com/MamangRust/microservice-point-of-sale-product/repository"
	"github.com/MamangRust/microservice-point-of-sale-product/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductServiceTestSuite struct {
	suite.Suite
	ts        *tests.TestSuite
	svc       *service.Service
	productID int
}

func (s *ProductServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextCatalog)
	s.ts = ts

	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	db := s.ts.MongoDB

	log, _ := logger.NewLogger("test", nil)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)
	mencache := prod_cache.NewMencache(cacheStore)

	obs, _ := observability.NewObservability("test", log)

	repos := repository.NewRepositories(db, nil, nil)
	repos.MerchantQuery = stubMerchantQueryRepository{}
	repos.CategoryQuery = stubCategoryQueryRepository{}

	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache,
		Observability: obs,
	})

	// Seed a merchant and category for product tests
	ctx := context.Background()
	now := time.Now().UTC()

	// Create user
	userID, _ := database.GetNextID(ctx, s.ts.DBForCollection("users"), "users")
	s.ts.Coll("users").InsertOne(ctx, bson.M{
		"_id":               userID,
		"firstname":         "Prod",
		"lastname":          "Svc",
		"email":             "prod.svc@example.com",
		"password":          "password123",
		"verification_code": "test-verify",
		"is_verified":       true,
		"created_at":        now,
		"updated_at":        now,
	})

	// Create category
	catID, _ := database.GetNextID(ctx, db, "categories")
	db.Collection("categories").InsertOne(ctx, bson.M{
		"_id":         catID,
		"name":        "Svc Category",
		"description": "Category for service tests",
		"created_at":  now,
		"updated_at":  now,
	})

	// Create merchant
	merchID, _ := database.GetNextID(ctx, s.ts.DBForCollection("merchants"), "merchants")
	s.ts.Coll("merchants").InsertOne(ctx, bson.M{
		"_id":           merchID,
		"user_id":       userID,
		"name":          "Svc Merchant",
		"description":   "Desc",
		"address":       "Addr",
		"contact_email": "ps@example.com",
		"contact_phone": "123",
		"status":        "active",
		"created_at":    now,
		"updated_at":    now,
	})
}

func (s *ProductServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *ProductServiceTestSuite) TestProductLifecycle() {
	ctx := context.Background()

	// 1. Create
	req := &requests.CreateProductRequest{
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Test Product",
		Description:  "Product description",
		Price:        10000,
		CountInStock: 100,
		Brand:        "Test Brand",
		Weight:       1000,
		ImageProduct: "test.jpg",
	}
	created, err := s.svc.ProductCommand.CreateProduct(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	productID := int(created.ProductID)

	// 2. FindByID
	found, err := s.svc.ProductQuery.FindById(ctx, productID)
	s.Require().NoError(err)
	s.Equal(req.Name, found.Name)

	// 3. Update
	updateReq := &requests.UpdateProductRequest{
		ProductID:    &productID,
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Updated Product Name",
		Description:  "Updated description",
		Price:        20000,
		CountInStock: 90,
		Brand:        "Updated Brand",
		Weight:       1000,
		ImageProduct: "updated.jpg",
	}
	updated, err := s.svc.ProductCommand.UpdateProduct(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(updateReq.Name, updated.Name)

	// 4. FindAll
	_, total, err := s.svc.ProductQuery.FindAll(ctx, &requests.FindAllProducts{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 5. Trash
	_, err = s.svc.ProductCommand.TrashProduct(ctx, productID)
	s.Require().NoError(err)

	// 6. FindTrashed
	_, totalTrashed, err := s.svc.ProductQuery.FindByTrashed(ctx, &requests.FindAllProducts{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	// 7. FindActive
	active, _, err := s.svc.ProductQuery.FindByActive(ctx, &requests.FindAllProducts{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, p := range active {
		s.NotEqual(productID, int(p.ProductID))
	}

	// 8. Restore
	_, err = s.svc.ProductCommand.RestoreProduct(ctx, productID)
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, err = s.svc.ProductCommand.TrashProduct(ctx, productID)
	s.Require().NoError(err)
	success, err := s.svc.ProductCommand.DeleteProductPermanent(ctx, productID)
	s.Require().NoError(err)
	s.True(success)

	// 10. RestoreAll & DeleteAll
	p1, _ := s.svc.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: 1, CategoryID: 1, Name: "P1", Description: "D1", Price: 100,
		CountInStock: 10, Brand: "B", Weight: 1, ImageProduct: "i1.jpg",
	})
	p2, _ := s.svc.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: 1, CategoryID: 1, Name: "P2", Description: "D2", Price: 200,
		CountInStock: 20, Brand: "B", Weight: 2, ImageProduct: "i2.jpg",
	})

	s.svc.ProductCommand.TrashProduct(ctx, int(p1.ProductID))
	s.svc.ProductCommand.TrashProduct(ctx, int(p2.ProductID))

	resRestoreAll, err := s.svc.ProductCommand.RestoreAllProducts(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.ProductCommand.TrashProduct(ctx, int(p1.ProductID))
	s.svc.ProductCommand.TrashProduct(ctx, int(p2.ProductID))

	resDeleteAll, err := s.svc.ProductCommand.DeleteAllProductsPermanent(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

// TestProductCategoryResolution covers the two query paths that used to read
// the `categories` collection directly from product-service:
//
//   - FindByMerchant used a Mongo $lookup to fill category_name
//   - FindByCategory looked the category up by name in Mongo
//
// Both now resolve through the Category adapter (stubbed here), so the product
// repository no longer touches a collection it does not own.
func (s *ProductServiceTestSuite) TestProductCategoryResolution() {
	ctx := context.Background()

	for _, name := range []string{"Cat Resolution A", "Cat Resolution B"} {
		_, err := s.svc.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
			MerchantID: 1, CategoryID: 1, Name: name, Description: "D", Price: 500,
			CountInStock: 5, Brand: "B", Weight: 1, ImageProduct: "c.jpg",
		})
		s.Require().NoError(err)
	}

	byMerchant, merchantTotal, err := s.svc.ProductQuery.FindByMerchant(ctx, &requests.ProductByMerchantRequest{
		MerchantID: 1, Page: 1, PageSize: 10,
	})
	s.Require().NoError(err)
	s.GreaterOrEqual(*merchantTotal, 2)
	s.Require().NotEmpty(byMerchant)
	for _, p := range byMerchant {
		s.Equal("Test Category", p.CategoryName, "category name must come from the Category adapter")
	}

	byCategory, categoryTotal, err := s.svc.ProductQuery.FindByCategory(ctx, &requests.ProductByCategoryRequest{
		CategoryName: "Svc Category", Page: 1, PageSize: 10,
	})
	s.Require().NoError(err)
	s.GreaterOrEqual(*categoryTotal, 2)
	s.Require().NotEmpty(byCategory)
	for _, p := range byCategory {
		s.Equal(int32(1), p.CategoryID, "category name must be resolved to an id via the Category adapter")
	}
}

func TestProductServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductServiceTestSuite))
}

// stubMerchantQueryRepository stands in for the Merchant service gRPC call so
// the product service tests do not need a running merchant server.
type stubMerchantQueryRepository struct{}

func (stubMerchantQueryRepository) FindById(_ context.Context, merchantID int) (*models.Merchant, error) {
	return &models.Merchant{
		MerchantID: int32(merchantID),
		UserID:     1,
		Name:       "Test Merchant",
		Status:     "active",
	}, nil
}

// stubCategoryQueryRepository stands in for the Category service gRPC call so
// the product service tests do not need a running category server.
type stubCategoryQueryRepository struct{}

func (stubCategoryQueryRepository) FindById(_ context.Context, categoryID int) (*models.Category, error) {
	return &models.Category{
		CategoryID: int32(categoryID),
		Name:       "Test Category",
	}, nil
}

func (stubCategoryQueryRepository) FindByIds(_ context.Context, categoryIDs []int) ([]*models.Category, error) {
	categories := make([]*models.Category, 0, len(categoryIDs))
	for _, id := range categoryIDs {
		categories = append(categories, &models.Category{
			CategoryID: int32(id),
			Name:       "Test Category",
		})
	}
	return categories, nil
}

func (stubCategoryQueryRepository) FindByName(_ context.Context, name string) (*models.Category, error) {
	return &models.Category{
		CategoryID: 1,
		Name:       name,
	}, nil
}
