package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-transacton/repository"
)

// EmailEventPublisher publishes email notification events to Kafka topics.
type EmailEventPublisher interface {
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
}

type TransactionQueryService interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllTransaction) ([]*repository.TransactionResult, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*repository.TransactionByMerchantResult, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransaction) ([]*repository.TransactionResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*repository.TransactionResultDeleteAt, *int, error)
	FindById(ctx context.Context, transactionID int) (*models.Transaction, error)
	FindByOrderId(ctx context.Context, orderID int) (*models.Transaction, error)
}

type TransactionCommandService interface {
	CreateTransaction(ctx context.Context, req *requests.CreateTransactionRequest) (*models.Transaction, error)
	UpdateTransaction(ctx context.Context, req *requests.UpdateTransactionRequest) (*models.Transaction, error)
	TrashedTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error)
	RestoreTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error)
	DeleteTransactionPermanently(ctx context.Context, transactionID int) (bool, error)
	RestoreAllTransactions(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
}
