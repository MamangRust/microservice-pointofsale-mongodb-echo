// Package transaction adapts the Transaction service gRPC API into the shared
// domain model. Only the bulk (enumeration) surface is exposed today: the sole
// consumer is the stats backfill, which walks the whole collection.
package transaction

import (
	"context"

	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
)

// BulkRepository enumerates transactions page by page. It is a narrow interface
// for offline consumers (the stats backfill) that need the whole collection.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Transaction, int, error)
}

// Repository implements BulkRepository over a gRPC client.
type Repository struct {
	client pbtransaction.TransactionServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a transaction adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pbtransaction.TransactionServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

// FindAll returns one page of active transactions plus the total number of
// active rows, so callers can walk the collection page by page.
func (r *Repository) FindAll(ctx context.Context, page, pageSize int) ([]*models.Transaction, int, error) {
	var resp *pbtransaction.ApiResponsePaginationTransaction
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindAll(ctx, &pbtransaction.FindAllTransactionRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil || resp == nil {
		return nil, 0, transaction_errors.ErrFindAllTransactions
	}

	transactions := make([]*models.Transaction, 0, len(resp.Data))
	for _, t := range resp.Data {
		if t == nil {
			continue
		}
		transactions = append(transactions, mapTransaction(t))
	}

	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return transactions, total, nil
}

// mapTransaction converts a proto transaction into the shared domain model.
func mapTransaction(t *pbtransaction.TransactionResponse) *models.Transaction {
	if t == nil {
		return nil
	}
	return &models.Transaction{
		TransactionID: t.Id,
		OrderID:       t.OrderId,
		MerchantID:    t.MerchantId,
		PaymentMethod: t.PaymentMethod,
		Amount:        t.Amount,
		ChangeAmount:  convert.NullableInt32(t.ChangeAmount),
		PaymentStatus: convert.NullableString(t.PaymentStatus),
		CreatedAt:     convert.TimePtr(t.CreatedAt),
		UpdatedAt:     convert.TimePtr(t.UpdatedAt),
	}
}
