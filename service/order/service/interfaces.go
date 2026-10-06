package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResult, *int, error)
	FindById(ctx context.Context, orderID int) (*models.Order, error)
	FindByActive(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResultDeleteAt, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*repository.OrderResult, *int, error)
}

type OrderCommandService interface {
	CreateOrder(ctx context.Context, req *requests.CreateOrderRequest) (*models.Order, error)
	UpdateOrder(ctx context.Context, req *requests.UpdateOrderRequest) (*models.Order, error)
	TrashedOrder(ctx context.Context, orderID int) (*models.Order, error)
	RestoreOrder(ctx context.Context, orderID int) (*models.Order, error)
	DeleteOrderPermanent(ctx context.Context, orderID int) (bool, error)
	RestoreAllOrder(ctx context.Context) (bool, error)
	DeleteAllOrderPermanent(ctx context.Context) (bool, error)
}
