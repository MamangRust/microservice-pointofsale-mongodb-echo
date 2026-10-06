package mencache

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderItemQueryCache interface {
	GetCachedOrderItemsAll(ctx context.Context, req *requests.FindAllOrderItems) ([]*repository.OrderItemResult, *int, bool)
	SetCachedOrderItemsAll(ctx context.Context, req *requests.FindAllOrderItems, data []*repository.OrderItemResult, total *int)

	GetCachedOrderItemActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*repository.OrderItemResultDeleteAt, *int, bool)
	SetCachedOrderItemActive(ctx context.Context, req *requests.FindAllOrderItems, data []*repository.OrderItemResultDeleteAt, total *int)

	GetCachedOrderItemTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*repository.OrderItemResultDeleteAt, *int, bool)
	SetCachedOrderItemTrashed(ctx context.Context, req *requests.FindAllOrderItems, data []*repository.OrderItemResultDeleteAt, total *int)

	GetCachedOrderItems(ctx context.Context, orderID int) ([]*models.OrderItem, bool)
	SetCachedOrderItems(ctx context.Context, data []*models.OrderItem)
}

type OrderItemCommandCache interface {
	DeleteOrderItemCache(ctx context.Context, orderID int)
	DeleteOrderItemAllCache(ctx context.Context)
}
