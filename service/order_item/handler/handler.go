package handler

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-order-item/service"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

// Handler groups all order_item gRPC handlers.
type Handler interface {
	pborderitem.OrderItemServiceServer
}

// handler implements the full OrderItemServiceServer by embedding the query
// sub-handler (which promotes FindAll/FindByActive/FindByTrashed/
// FindOrderItemByOrder) and delegating the command RPCs to the command
// sub-handler.
//
// The command sub-handler is a named field rather than an embedded one: both
// sub-handlers implement the whole OrderItemServiceServer, so embedding both
// would make every method ambiguous.
type handler struct {
	OrderItemQueryHandleGrpc
	command OrderItemCommandHandleGrpc
}

// NewHandler initializes order_item gRPC handlers.
func NewHandler(svc *service.Service, logger logger.LoggerInterface) Handler {
	return &handler{
		OrderItemQueryHandleGrpc: NewOrderItemQueryHandleGrpc(svc.OrderItemQuery, logger),
		command:                  NewOrderItemCommandHandleGrpc(svc.OrderItemCommand, logger),
	}
}

func (h *handler) CreateOrderItem(ctx context.Context, request *pborderitem.CreateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	return h.command.CreateOrderItem(ctx, request)
}

func (h *handler) UpdateOrderItem(ctx context.Context, request *pborderitem.UpdateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	return h.command.UpdateOrderItem(ctx, request)
}

func (h *handler) DeleteOrderItem(ctx context.Context, request *pborderitem.DeleteOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItemDelete, error) {
	return h.command.DeleteOrderItem(ctx, request)
}
