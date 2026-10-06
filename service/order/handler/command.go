package handler

import (
	"context"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	order_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *orderQueryHandleGrpc) Create(ctx context.Context, request *pb.CreateOrderRequest) (*pb.ApiResponseOrder, error) {
	req := &requests.CreateOrderRequest{
		MerchantID: int(request.GetMerchantId()),
		CashierID:  int(request.GetCashierId()),
	}
	for _, item := range request.GetItems() {
		req.Items = append(req.Items, requests.CreateOrderItemRequest{
			ProductID: int(item.GetProductId()),
			Quantity:  int(item.GetQuantity()),
		})
	}
	if err := req.Validate(); err != nil {
		return nil, order_errors.ErrGrpcValidateCreateOrder
	}
	order, err := s.orderCommandService.CreateOrder(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrder{Status: "success", Message: "Successfully created order", Data: mapResponseOrderFromModel(order)}, nil
}

func (s *orderQueryHandleGrpc) Update(ctx context.Context, request *pb.UpdateOrderRequest) (*pb.ApiResponseOrder, error) {
	id := int(request.GetOrderId())
	if id <= 0 {
		return nil, order_errors.ErrGrpcFailedInvalidId
	}
	req := &requests.UpdateOrderRequest{OrderID: &id}
	for _, item := range request.GetItems() {
		req.Items = append(req.Items, requests.UpdateOrderItemRequest{
			OrderItemID: int(item.GetOrderItemId()),
			ProductID:   int(item.GetProductId()),
			Quantity:    int(item.GetQuantity()),
		})
	}
	if err := req.Validate(); err != nil {
		return nil, order_errors.ErrGrpcValidateUpdateOrder
	}
	order, err := s.orderCommandService.UpdateOrder(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrder{Status: "success", Message: "Successfully updated order", Data: mapResponseOrderFromModel(order)}, nil
}

func (s *orderQueryHandleGrpc) TrashedOrder(ctx context.Context, request *pb.FindByIdOrderRequest) (*pb.ApiResponseOrderDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, order_errors.ErrGrpcFailedInvalidId
	}
	order, err := s.orderCommandService.TrashedOrder(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrderDeleteAt{Status: "success", Message: "Successfully trashed order", Data: mapResponseOrderDeleteAtFromModel(order)}, nil
}

func (s *orderQueryHandleGrpc) RestoreOrder(ctx context.Context, request *pb.FindByIdOrderRequest) (*pb.ApiResponseOrderDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, order_errors.ErrGrpcFailedInvalidId
	}
	order, err := s.orderCommandService.RestoreOrder(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrderDeleteAt{Status: "success", Message: "Successfully restored order", Data: mapResponseOrderDeleteAtFromModel(order)}, nil
}

func (s *orderQueryHandleGrpc) DeleteOrderPermanent(ctx context.Context, request *pb.FindByIdOrderRequest) (*pb.ApiResponseOrderDelete, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, order_errors.ErrGrpcFailedInvalidId
	}
	_, err := s.orderCommandService.DeleteOrderPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrderDelete{Status: "success", Message: "Successfully deleted order permanently"}, nil
}

func (s *orderQueryHandleGrpc) RestoreAllOrder(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseOrderAll, error) {
	_, err := s.orderCommandService.RestoreAllOrder(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrderAll{Status: "success", Message: "Successfully restore all order"}, nil
}

func (s *orderQueryHandleGrpc) DeleteAllOrderPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseOrderAll, error) {
	_, err := s.orderCommandService.DeleteAllOrderPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrderAll{Status: "success", Message: "Successfully delete all order permanen"}, nil
}
