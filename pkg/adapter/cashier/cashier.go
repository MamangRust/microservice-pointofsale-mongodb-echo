// Package cashier adapts the Cashier service gRPC API into the shared domain
// model. Query and bulk are exposed as separate interfaces so consumers depend
// on the narrower one: ordinary consumers read a single cashier, while offline
// consumers (the stats backfill) enumerate the whole collection.
package cashier

import (
	"context"

	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
)

// QueryRepository reads a single cashier from the Cashier service over gRPC.
type QueryRepository interface {
	FindById(ctx context.Context, id int) (*models.Cashier, error)
}

// BulkRepository enumerates cashiers page by page. It is deliberately a
// separate narrow interface: offline consumers (the stats backfill) need the
// whole collection, while ordinary consumers only ever read one cashier at a
// time and should not depend on the extra method.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Cashier, int, error)
}

// Repository implements QueryRepository and BulkRepository over a single gRPC
// client, so consumers can share one instance across both roles.
type Repository struct {
	client pbcashier.CashierServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a cashier adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pbcashier.CashierServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) FindById(ctx context.Context, id int) (*models.Cashier, error) {
	var resp *pbcashier.ApiResponseCashier
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindById(ctx, &pbcashier.FindByIdCashierRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, cashier_errors.ErrFindCashierById
	}

	return mapCashier(resp.Data), nil
}

// FindAll returns one page of active cashiers plus the total number of active
// rows, so callers can walk the collection page by page.
func (r *Repository) FindAll(ctx context.Context, page, pageSize int) ([]*models.Cashier, int, error) {
	var resp *pbcashier.ApiResponsePaginationCashier
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindAll(ctx, &pbcashier.FindAllCashierRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil || resp == nil {
		return nil, 0, cashier_errors.ErrFindAllCashiers
	}

	cashiers := make([]*models.Cashier, 0, len(resp.Data))
	for _, c := range resp.Data {
		if c == nil {
			continue
		}
		cashiers = append(cashiers, mapCashier(c))
	}

	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return cashiers, total, nil
}

// mapCashier converts a proto cashier into the shared domain model.
func mapCashier(c *pbcashier.CashierResponse) *models.Cashier {
	if c == nil {
		return nil
	}
	return &models.Cashier{
		CashierID:  c.Id,
		MerchantID: c.MerchantId,
		Name:       c.Name,
		CreatedAt:  convert.TimePtr(c.CreatedAt),
		UpdatedAt:  convert.TimePtr(c.UpdatedAt),
	}
}
