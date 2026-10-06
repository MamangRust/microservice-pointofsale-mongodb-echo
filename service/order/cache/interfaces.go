package mencache

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderQueryCache interface {
	GetOrderAllCache(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResult, *int, bool)
	SetOrderAllCache(ctx context.Context, req *requests.FindAllOrders, data []*repository.OrderResult, total *int)

	GetCachedOrderCache(ctx context.Context, orderID int) (*models.Order, bool)
	SetCachedOrderCache(ctx context.Context, data *models.Order)

	GetCachedOrderMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*repository.OrderResult, *int, bool)
	SetCachedOrderMerchant(ctx context.Context, req *requests.FindAllOrderMerchant, res []*repository.OrderResult, total *int)

	GetOrderActiveCache(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResultDeleteAt, *int, bool)
	SetOrderActiveCache(ctx context.Context, req *requests.FindAllOrders, data []*repository.OrderResultDeleteAt, total *int)

	GetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrders) ([]*repository.OrderResultDeleteAt, *int, bool)
	SetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrders, data []*repository.OrderResultDeleteAt, total *int)
}

type OrderCommandCache interface {
	DeleteOrderCache(ctx context.Context, id int)
	DeleteOrderAllCache(ctx context.Context)
}
