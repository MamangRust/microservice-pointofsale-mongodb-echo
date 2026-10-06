package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// OrderResult is the result type for paginated order queries.
type OrderResult struct {
	OrderID    int32
	MerchantID int32
	CashierID  int32
	TotalPrice int64
	CreatedAt  string
	UpdatedAt  string
	TotalCount int64
}

// OrderResultDeleteAt is the result type for paginated order queries with deleted_at.
type OrderResultDeleteAt struct {
	OrderID    int32
	MerchantID int32
	CashierID  int32
	TotalPrice int64
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
	TotalCount int64
}

type OrderQueryRepository interface {
	FindAllOrders(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResult, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResultDeleteAt, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*OrderResult, *int, error)
	FindById(ctx context.Context, orderID int) (*models.Order, error)
	FindByTrashedId(ctx context.Context, orderID int) (*models.Order, error)
}

type OrderCommandRepository interface {
	DeleteOrder(ctx context.Context, orderID int) error
	CreateOrder(ctx context.Context, request *requests.CreateOrderRecordRequest) (*models.Order, error)
	UpdateOrder(ctx context.Context, request *requests.UpdateOrderRecordRequest) (*models.Order, error)
	FindAllTrashed(ctx context.Context) ([]*models.Order, error)
	TrashedOrder(ctx context.Context, orderID int) (*models.Order, error)
	RestoreOrder(ctx context.Context, orderID int) (*models.Order, error)
	DeleteOrderPermanent(ctx context.Context, orderID int) (bool, error)
	DeleteAllOrderPermanent(ctx context.Context) (bool, error)
}
