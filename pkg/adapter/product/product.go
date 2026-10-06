// Package product adapts the Product service gRPC API into the shared domain
// model. Query and command are exposed as separate interfaces so consumers can
// depend on the narrower one.
package product

import (
	"context"

	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryRepository reads products from the Product service over gRPC.
type QueryRepository interface {
	FindById(ctx context.Context, id int) (*models.Product, error)
}

// CommandRepository mutates product stock via the Product service over gRPC.
type CommandRepository interface {
	DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error)
	IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error)
}

// BulkRepository enumerates products page by page. It is deliberately a
// separate narrow interface: offline consumers (the stats backfill) need the
// whole catalog, while ordinary consumers only ever read one product at a time
// and should not depend on the extra method.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Product, int, error)
}

// Repository implements both QueryRepository and CommandRepository over a
// single gRPC client, so consumers can share one instance across both roles.
type Repository struct {
	client pbproduct.ProductServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a product adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pbproduct.ProductServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) FindById(ctx context.Context, id int) (*models.Product, error) {
	var resp *pbproduct.ApiResponseProduct
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindById(ctx, &pbproduct.FindByIdProductRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, product_errors.ErrFindById
	}

	return mapResponse(resp.Data), nil
}

// FindAll returns one page of active products plus the total number of active
// rows, so callers can walk the catalog page by page.
func (r *Repository) FindAll(ctx context.Context, page, pageSize int) ([]*models.Product, int, error) {
	var resp *pbproduct.ApiResponsePaginationProduct
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindAll(ctx, &pbproduct.FindAllProductRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil || resp == nil {
		return nil, 0, product_errors.ErrFindAllProducts
	}

	products := make([]*models.Product, 0, len(resp.Data))
	for _, p := range resp.Data {
		if p == nil {
			continue
		}
		products = append(products, mapResponse(p))
	}

	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return products, total, nil
}

func (r *Repository) DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return r.adjustProductCountStock(ctx, productID, quantity, true)
}

func (r *Repository) IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return r.adjustProductCountStock(ctx, productID, quantity, false)
}

// adjustProductCountStock delegates the stock mutation to the Product service,
// which owns the products collection. Consumers must never write to products
// directly.
func (r *Repository) adjustProductCountStock(ctx context.Context, productID int, quantity int, decrement bool) (*models.Product, error) {
	req := &pbproduct.AdjustProductStockRequest{
		ProductId: int32(productID),
		Quantity:  int32(quantity),
	}

	var resp *pbproduct.ApiResponseProduct
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		if decrement {
			resp, callErr = r.client.DecrementStock(ctx, req)
		} else {
			resp, callErr = r.client.IncrementStock(ctx, req)
		}
		return callErr
	})
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return nil, order_errors.ErrInsufficientProductStock
		}
		return nil, product_errors.ErrUpdateProductCountStock
	}

	if resp == nil || resp.Data == nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}

	return mapResponse(resp.Data), nil
}

// mapResponse converts a proto product into the shared domain model.
func mapResponse(p *pbproduct.ProductResponse) *models.Product {
	if p == nil {
		return nil
	}
	return &models.Product{
		ProductID:    p.Id,
		MerchantID:   p.MerchantId,
		CategoryID:   p.CategoryId,
		Name:         p.Name,
		Description:  convert.NullableString(p.Description),
		Price:        p.Price,
		CountInStock: p.CountInStock,
		Brand:        convert.NullableString(p.Brand),
		Weight:       convert.NullableInt32(p.Weight),
		SlugProduct:  convert.NullableString(p.SlugProduct),
		ImageProduct: convert.NullableString(p.ImageProduct),
		Barcode:      convert.NullableString(p.Barcode),
		CreatedAt:    convert.TimePtr(p.CreatedAt),
		UpdatedAt:    convert.TimePtr(p.UpdatedAt),
	}
}
