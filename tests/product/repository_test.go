package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-product/repository"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
	ts        *tests.TestSuite
	repo      *repository.Repositories
	productID int
}

func (s *ProductRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite(tests.ContextCatalog)
	s.ts = ts

	s.Require().NoError(err)

	db := s.ts.MongoDB
	s.repo = repository.NewRepositories(db, nil, nil)

	// Seed a merchant and category for product tests
	ctx := context.Background()
	now := time.Now().UTC()

	// Create user
	userID, _ := database.GetNextID(ctx, s.ts.DBForCollection("users"), "users")
	s.ts.Coll("users").InsertOne(ctx, bson.M{
		"_id":               userID,
		"firstname":         "Prod",
		"lastname":          "Repo",
		"email":             "prod.repo@example.com",
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
		"name":        "Test Category",
		"description": "Category for product tests",
		"created_at":  now,
		"updated_at":  now,
	})

	// Create merchant
	merchID, _ := database.GetNextID(ctx, s.ts.DBForCollection("merchants"), "merchants")
	s.ts.Coll("merchants").InsertOne(ctx, bson.M{
		"_id":           merchID,
		"user_id":       userID,
		"name":          "Test Merchant",
		"description":   "Desc",
		"address":       "Addr",
		"contact_email": "pm@example.com",
		"contact_phone": "123",
		"status":        "active",
		"created_at":    now,
		"updated_at":    now,
	})
}

func (s *ProductRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *ProductRepositoryTestSuite) Test1_CreateProduct() {
	ctx := context.Background()

	req := &requests.CreateProductRequest{
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Test Product",
		Description:  "Product description",
		Price:        100,
		CountInStock: 50,
		Brand:        "Test Brand",
		Weight:       1,
		ImageProduct: "test.jpg",
	}

	res, err := s.repo.ProductCommand.CreateProduct(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Name, res.Name)
	s.productID = int(res.ProductID)
}

func (s *ProductRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	found, err := s.repo.ProductQuery.FindById(ctx, s.productID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.productID, int(found.ProductID))
}

func (s *ProductRepositoryTestSuite) Test3_UpdateProduct() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	req := &requests.UpdateProductRequest{
		ProductID:    &s.productID,
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Updated Product",
		Description:  "Updated description",
		Price:        200,
		CountInStock: 100,
		Brand:        "Updated Brand",
		Weight:       2,
		ImageProduct: "updated.jpg",
	}

	res, err := s.repo.ProductCommand.UpdateProduct(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Product", res.Name)
}

func (s *ProductRepositoryTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	// Trash
	trashed, err := s.repo.ProductCommand.TrashedProduct(ctx, s.productID)
	s.NoError(err)
	s.NotNil(trashed)

	// Restore
	restored, err := s.repo.ProductCommand.RestoreProduct(ctx, s.productID)
	s.NoError(err)
	s.NotNil(restored)

	// Verify restored
	found, err := s.repo.ProductQuery.FindById(ctx, s.productID)
	s.NoError(err)
	s.NotNil(found)
}

func (s *ProductRepositoryTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	// Must be trashed first for permanent delete
	_, err := s.repo.ProductCommand.TrashedProduct(ctx, s.productID)
	s.NoError(err)

	success, err := s.repo.ProductCommand.DeleteProductPermanent(ctx, s.productID)
	s.NoError(err)
	s.True(success)
}

func TestProductRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductRepositoryTestSuite))
}
