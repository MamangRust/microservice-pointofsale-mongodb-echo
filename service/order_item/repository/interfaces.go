package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderItemResult struct {
	OrderItemID int32
	OrderID     int32
	ProductID   int32
	Quantity    int32
	Price       int32
	CreatedAt   *string
	UpdatedAt   *string
	TotalCount  int64
}

type OrderItemResultDeleteAt struct {
	OrderItemID int32
	OrderID     int32
	ProductID   int32
	Quantity    int32
	Price       int32
	CreatedAt   *string
	UpdatedAt   *string
	DeletedAt   *string
	TotalCount  int64
}

type OrderItemQueryRepository interface {
	FindAllOrderItems(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResult, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResultDeleteAt, *int, error)
	FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error)
}

type OrderItemCommandRepository interface {
	CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error)
	UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error)
	DeleteOrderItem(ctx context.Context, orderItemID int) error
}
