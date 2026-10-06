package handler

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-order-item/service"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"go.uber.org/zap"
)

type orderItemCommandHandleGrpc struct {
	pborderitem.UnimplementedOrderItemServiceServer
	orderItemService service.OrderItemCommandService
	logger           logger.LoggerInterface
}

func NewOrderItemCommandHandleGrpc(
	orderItemService service.OrderItemCommandService,
	logger logger.LoggerInterface,
) pborderitem.OrderItemServiceServer {
	return &orderItemCommandHandleGrpc{
		orderItemService: orderItemService,
		logger:           logger,
	}
}

func (s *orderItemCommandHandleGrpc) CreateOrderItem(ctx context.Context, request *pborderitem.CreateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	s.logger.Info("CreateOrderItem called", zap.Int32("order_id", request.GetOrderId()), zap.Int32("product_id", request.GetProductId()))

	if request.GetOrderId() <= 0 || request.GetProductId() <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}

	req := &requests.CreateOrderItemRecordRequest{
		OrderID:   int(request.GetOrderId()),
		ProductID: int(request.GetProductId()),
		Quantity:  int(request.GetQuantity()),
		Price:     int(request.GetPrice()),
	}

	item, err := s.orderItemService.CreateOrderItem(ctx, req)
	if err != nil {
		s.logger.Error("CreateOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("CreateOrderItem success", zap.Int32("order_item.id", item.OrderItemID))

	return &pborderitem.ApiResponseOrderItem{
		Status:  "success",
		Message: "Successfully created order item",
		Data:    mapOrderItemModel(item),
	}, nil
}

func (s *orderItemCommandHandleGrpc) UpdateOrderItem(ctx context.Context, request *pborderitem.UpdateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	s.logger.Info("UpdateOrderItem called", zap.Int32("order_item_id", request.GetOrderItemId()))

	if request.GetOrderItemId() <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}

	req := &requests.UpdateOrderItemRecordRequest{
		OrderItemID: int(request.GetOrderItemId()),
		OrderID:     int(request.GetOrderId()),
		ProductID:   int(request.GetProductId()),
		Quantity:    int(request.GetQuantity()),
		Price:       int(request.GetPrice()),
	}

	item, err := s.orderItemService.UpdateOrderItem(ctx, req)
	if err != nil {
		s.logger.Error("UpdateOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("UpdateOrderItem success", zap.Int32("order_item.id", item.OrderItemID))

	return &pborderitem.ApiResponseOrderItem{
		Status:  "success",
		Message: "Successfully updated order item",
		Data:    mapOrderItemModel(item),
	}, nil
}

func (s *orderItemCommandHandleGrpc) DeleteOrderItem(ctx context.Context, request *pborderitem.DeleteOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItemDelete, error) {
	s.logger.Info("DeleteOrderItem called", zap.Int32("order_item_id", request.GetOrderItemId()))

	orderItemID := int(request.GetOrderItemId())
	if orderItemID <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}

	if _, err := s.orderItemService.DeleteOrderItem(ctx, orderItemID); err != nil {
		s.logger.Error("DeleteOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteOrderItem success", zap.Int("order_item.id", orderItemID))

	return &pborderitem.ApiResponseOrderItemDelete{
		Status:  "success",
		Message: "Successfully deleted order item",
	}, nil
}
