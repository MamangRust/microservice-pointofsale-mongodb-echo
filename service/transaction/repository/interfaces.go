package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// TransactionResult is the result type for paginated transaction queries.
type TransactionResult struct {
	TransactionID int32
	OrderID       int32
	MerchantID    int32
	PaymentMethod string
	Amount        int32
	ChangeAmount  *int32
	PaymentStatus string
	CreatedAt     string
	TotalCount    int64
}

// TransactionResultDeleteAt is the result type with deleted_at.
type TransactionResultDeleteAt struct {
	TransactionID int32
	OrderID       int32
	MerchantID    int32
	PaymentMethod string
	Amount        int32
	ChangeAmount  *int32
	PaymentStatus string
	CreatedAt     string
	DeletedAt     string
	TotalCount    int64
}

// TransactionByMerchantResult is the result type for merchant-filtered queries.
type TransactionByMerchantResult struct {
	TransactionID int32
	OrderID       int32
	MerchantID    int32
	PaymentMethod string
	Amount        int32
	ChangeAmount  *int32
	PaymentStatus string
	CreatedAt     string
	TotalCount    int64
}

type TransactionQueryRepository interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResult, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResultDeleteAt, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*TransactionByMerchantResult, *int, error)
	FindById(ctx context.Context, transaction_id int) (*models.Transaction, error)
	FindByOrderId(ctx context.Context, order_id int) (*models.Transaction, error)
}

type TransactionCommandRepository interface {
	CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*models.Transaction, error)
	UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*models.Transaction, error)
	TrashTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error)
	RestoreTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error)
	DeleteTransactionPermanently(ctx context.Context, transaction_id int) (bool, error)
	RestoreAllTransactions(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
}
