package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// ProductResult is the result type for paginated product queries.
type ProductResult struct {
	ProductID    int32
	MerchantID   int32
	CategoryID   int32
	Name         string
	Description  *string
	Price        int32
	CountInStock int32
	Brand        *string
	Weight       *int32
	SlugProduct  *string
	ImageProduct *string
	Barcode      *string
	CreatedAt    string
	UpdatedAt    string
	TotalCount   int64
}

// ProductResultDeleteAt is the result type for paginated product queries with deleted_at.
type ProductResultDeleteAt struct {
	ProductID    int32
	MerchantID   int32
	CategoryID   int32
	Name         string
	Description  *string
	Price        int32
	CountInStock int32
	Brand        *string
	Weight       *int32
	SlugProduct  *string
	ImageProduct *string
	Barcode      *string
	CreatedAt    string
	UpdatedAt    string
	DeletedAt    string
	TotalCount   int64
}

// ProductByMerchantResult is the result type for product-by-merchant queries.
// CategoryName is filled in by the service layer via the Category adapter, not
// by a Mongo $lookup.
type ProductByMerchantResult struct {
	TotalCount   int64
	ProductID    int32
	CategoryID   int32
	Name         string
	Description  *string
	Price        int32
	CountInStock int32
	Brand        *string
	ImageProduct *string
	Barcode      *string
	CreatedAt    string
	CategoryName string
}

// ProductByCategoryResult is the result type for product-by-category queries.
type ProductByCategoryResult struct {
	TotalCount   int64
	ProductID    int32
	MerchantID   int32
	CategoryID   int32
	SlugProduct  *string
	Weight       *int32
	Name         string
	Description  *string
	Price        int32
	CountInStock int32
	Brand        *string
	ImageProduct *string
	Barcode      *string
	CreatedAt    string
	UpdatedAt    string
}

type ProductQueryRepository interface {
	FindAllProducts(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResult, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResultDeleteAt, *int, error)
	FindByMerchant(ctx context.Context, req *requests.ProductByMerchantRequest) ([]*ProductByMerchantResult, *int, error)
	FindByCategoryID(ctx context.Context, categoryID int32, req *requests.ProductByCategoryRequest) ([]*ProductByCategoryResult, *int, error)
	FindById(ctx context.Context, product_id int) (*models.Product, error)
	FindByIdTrashed(ctx context.Context, id int) (*models.Product, error)
}

type ProductCommandRepository interface {
	CreateProduct(ctx context.Context, request *requests.CreateProductRequest) (*models.Product, error)
	UpdateProduct(ctx context.Context, request *requests.UpdateProductRequest) (*models.Product, error)
	UpdateProductCountStock(ctx context.Context, product_id int, stock int) (*models.Product, error)
	DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error)
	IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error)
	TrashedProduct(ctx context.Context, product_id int) (*models.Product, error)
	RestoreProduct(ctx context.Context, product_id int) (*models.Product, error)
	DeleteProductPermanent(ctx context.Context, product_id int) (bool, error)
	RestoreAllProducts(ctx context.Context) (bool, error)
	DeleteAllProductPermanent(ctx context.Context) (bool, error)
}
