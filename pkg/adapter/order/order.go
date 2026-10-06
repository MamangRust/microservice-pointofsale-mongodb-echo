// Package order adapts the Order service gRPC API into the shared domain model.
// Query and bulk are exposed as separate interfaces so consumers depend on the
// narrower one: ordinary consumers read a single order, while offline consumers
// (the stats backfill) enumerate the whole collection.
package order

import (
	"context"

	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
)

// QueryRepository reads a single order from the Order service over gRPC.
type QueryRepository interface {
	FindById(ctx context.Context, id int) (*models.Order, error)
}

// BulkRepository enumerates orders page by page. It is deliberately a separate
// narrow interface: offline consumers (the stats backfill) need the whole
// collection, while ordinary consumers only ever read one order at a time and
// should not depend on the extra method.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Order, int, error)
}

// Repository implements QueryRepository and BulkRepository over a single gRPC
// client, so consumers can share one instance across both roles.
type Repository struct {
	client pborder.OrderServiceClient
	guard  *resilience.DependencyGuard
}

// New builds an order adapter. Passing zero options leaves the guard nil, which
// makes DependencyGuard.Call a plain passthrough.
func New(client pborder.OrderServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) FindById(ctx context.Context, id int) (*models.Order, error) {
	var resp *pborder.ApiResponseOrder
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindById(ctx, &pborder.FindByIdOrderRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, order_errors.ErrFindById
	}

	return mapOrder(resp.Data), nil
}

// FindAll returns one page of active orders plus the total number of active
// rows, so callers can walk the collection page by page.
func (r *Repository) FindAll(ctx context.Context, page, pageSize int) ([]*models.Order, int, error) {
	var resp *pborder.ApiResponsePaginationOrder
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindAll(ctx, &pborder.FindAllOrderRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil || resp == nil {
		return nil, 0, order_errors.ErrFindAllOrders
	}

	orders := make([]*models.Order, 0, len(resp.Data))
	for _, o := range resp.Data {
		if o == nil {
			continue
		}
		orders = append(orders, mapOrder(o))
	}

	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return orders, total, nil
}

// mapOrder converts a proto order into the shared domain model.
func mapOrder(o *pborder.OrderResponse) *models.Order {
	if o == nil {
		return nil
	}
	return &models.Order{
		OrderID:    o.Id,
		MerchantID: o.MerchantId,
		CashierID:  o.CashierId,
		TotalPrice: int64(o.TotalPrice),
		CreatedAt:  convert.TimePtr(o.CreatedAt),
		UpdatedAt:  convert.TimePtr(o.UpdatedAt),
	}
}
