// Package orderitem adapts the OrderItem service gRPC API into the shared
// domain model. Query and command are exposed as separate interfaces so
// consumers can depend on the narrower one.
package orderitem

import (
	"context"

	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
)

// QueryRepository reads order items from the OrderItem service over gRPC.
type QueryRepository interface {
	FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error)
	CalculateTotalPrice(ctx context.Context, orderID int) (*int32, error)
}

// CommandRepository mutates order items via the OrderItem service over gRPC.
type CommandRepository interface {
	CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error)
	UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error)
	DeleteOrderItem(ctx context.Context, orderItemID int) error
}

// BulkRepository enumerates order items page by page. It is deliberately a
// separate narrow interface: offline consumers (the stats backfill) need the
// whole collection, while ordinary consumers only ever read one order at a
// time and should not depend on the extra method.
type BulkRepository interface {
	FindAllOrderItems(ctx context.Context, page, pageSize int) ([]*models.OrderItem, int, error)
}

// Repository implements both QueryRepository and CommandRepository over a
// single gRPC client, so consumers can share one instance across both roles.
type Repository struct {
	client pborderitem.OrderItemServiceClient
	guard  *resilience.DependencyGuard
}

// New builds an order-item adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pborderitem.OrderItemServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) CalculateTotalPrice(ctx context.Context, orderID int) (*int32, error) {
	items, err := r.FindOrderItemByOrder(ctx, orderID)
	if err != nil {
		return nil, orderitem_errors.ErrCalculateTotalPrice
	}

	var total int32 = 0
	for _, item := range items {
		if item != nil {
			total += item.Quantity * int32(item.Price)
		}
	}

	return &total, nil
}

func (r *Repository) FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error) {
	var resp *pborderitem.ApiResponsesOrderItem
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindOrderItemByOrder(ctx, &pborderitem.FindByIdOrderItemRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, orderitem_errors.ErrFindOrderItemByOrder
	}

	var res []*models.OrderItem
	for _, item := range resp.Data {
		if item == nil {
			continue
		}
		res = append(res, mapOrderItem(item))
	}

	return res, nil
}

// FindAllOrderItems returns one page of active order items plus the total number
// of active rows, so callers can walk the collection page by page.
func (r *Repository) FindAllOrderItems(ctx context.Context, page, pageSize int) ([]*models.OrderItem, int, error) {
	var resp *pborderitem.ApiResponsePaginationOrderItem
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindAll(ctx, &pborderitem.FindAllOrderItemRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil || resp == nil {
		return nil, 0, orderitem_errors.ErrFindAllOrderItems
	}

	items := make([]*models.OrderItem, 0, len(resp.Data))
	for _, item := range resp.Data {
		if item == nil {
			continue
		}
		items = append(items, mapOrderItem(item))
	}

	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return items, total, nil
}

func (r *Repository) CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error) {
	if req == nil {
		return nil, orderitem_errors.ErrCreateOrderItem
	}

	var resp *pborderitem.ApiResponseOrderItem
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.CreateOrderItem(ctx, &pborderitem.CreateOrderItemRecordRequest{
			OrderId:   int32(req.OrderID),
			ProductId: int32(req.ProductID),
			Quantity:  int32(req.Quantity),
			Price:     int32(req.Price),
		})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, orderitem_errors.ErrCreateOrderItem
	}

	return mapOrderItem(resp.Data), nil
}

func (r *Repository) UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error) {
	if req == nil {
		return nil, orderitem_errors.ErrUpdateOrderItem
	}

	var resp *pborderitem.ApiResponseOrderItem
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.UpdateOrderItem(ctx, &pborderitem.UpdateOrderItemRecordRequest{
			OrderItemId: int32(req.OrderItemID),
			OrderId:     int32(req.OrderID),
			ProductId:   int32(req.ProductID),
			Quantity:    int32(req.Quantity),
			Price:       int32(req.Price),
		})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, orderitem_errors.ErrUpdateOrderItem
	}

	return mapOrderItem(resp.Data), nil
}

func (r *Repository) DeleteOrderItem(ctx context.Context, orderItemID int) error {
	var resp *pborderitem.ApiResponseOrderItemDelete
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.DeleteOrderItem(ctx, &pborderitem.DeleteOrderItemRecordRequest{
			OrderItemId: int32(orderItemID),
		})
		return callErr
	})
	if err != nil || resp == nil {
		return orderitem_errors.ErrDeleteOrderItemPermanent
	}

	return nil
}

// mapOrderItem converts a proto order item into the shared domain model.
func mapOrderItem(item *pborderitem.OrderItemResponse) *models.OrderItem {
	if item == nil {
		return nil
	}
	return &models.OrderItem{
		OrderItemID: item.Id,
		OrderID:     item.OrderId,
		ProductID:   item.ProductId,
		Quantity:    item.Quantity,
		Price:       int64(item.Price),
		CreatedAt:   convert.TimePtr(item.CreatedAt),
		UpdatedAt:   convert.TimePtr(item.UpdatedAt),
	}
}
