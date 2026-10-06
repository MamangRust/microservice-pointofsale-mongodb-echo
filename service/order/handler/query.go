package handler

import (
	"context"
	"math"

	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-order/service"
	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	order_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
)

type orderQueryHandleGrpc struct {
	pb.UnimplementedOrderServiceServer

	orderQuery service.OrderQueryService

	orderCommandService service.OrderCommandService
}

func NewOrderHandleGrpc(query service.OrderQueryService, command service.OrderCommandService) *orderQueryHandleGrpc {
	return &orderQueryHandleGrpc{
		orderQuery:          query,
		orderCommandService: command,
	}
}

func (s *orderQueryHandleGrpc) FindAll(ctx context.Context, request *pb.FindAllOrderRequest) (*pb.ApiResponsePaginationOrder, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllOrders{Page: page, PageSize: pageSize, Search: search}
	orders, totalRecords, err := s.orderQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationOrder{
		Status: "success", Message: "Successfully fetched order",
		Data: mapResponsesOrder(orders), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *orderQueryHandleGrpc) FindById(ctx context.Context, request *pb.FindByIdOrderRequest) (*pb.ApiResponseOrder, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, order_errors.ErrGrpcFailedInvalidId
	}
	order, err := s.orderQuery.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseOrder{Status: "success", Message: "Successfully fetched order", Data: mapResponseOrderFromModel(order)}, nil
}

func (s *orderQueryHandleGrpc) FindByActive(ctx context.Context, request *pb.FindAllOrderRequest) (*pb.ApiResponsePaginationOrderDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllOrders{Page: page, PageSize: pageSize, Search: search}
	orders, totalRecords, err := s.orderQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationOrderDeleteAt{
		Status: "success", Message: "Successfully fetched active order",
		Data: mapResponsesOrderActive(orders), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *orderQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pb.FindAllOrderRequest) (*pb.ApiResponsePaginationOrderDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllOrders{Page: page, PageSize: pageSize, Search: search}
	orders, totalRecords, err := s.orderQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationOrderDeleteAt{
		Status: "success", Message: "Successfully fetched trashed order",
		Data: mapResponsesOrderTrashed(orders), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *orderQueryHandleGrpc) FindByMerchant(ctx context.Context, request *pb.FindAllOrderMerchantRequest) (*pb.ApiResponsePaginationOrder, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	merchantID := int(request.GetMerchantId())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllOrderMerchant{Page: page, PageSize: pageSize, Search: search, MerchantID: merchantID}
	orders, totalRecords, err := s.orderQuery.FindByMerchant(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationOrder{
		Status: "success", Message: "Successfully fetched order",
		Data: mapResponsesOrder(orders), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func mapResponseOrderFromModel(order *models.Order) *pb.OrderResponse {
	if order == nil {
		return nil
	}
	return &pb.OrderResponse{
		Id: order.OrderID, MerchantId: order.MerchantID, CashierId: order.CashierID,
		TotalPrice: int32(order.TotalPrice),
		CreatedAt:  convert.FormatTimePtr(order.CreatedAt), UpdatedAt: convert.FormatTimePtr(order.UpdatedAt),
	}
}

func mapResponseOrderDeleteAtFromModel(order *models.Order) *pb.OrderResponseDeleteAt {
	if order == nil {
		return nil
	}
	return &pb.OrderResponseDeleteAt{
		Id: order.OrderID, MerchantId: order.MerchantID, CashierId: order.CashierID,
		TotalPrice: int32(order.TotalPrice),
		CreatedAt:  convert.FormatTimePtr(order.CreatedAt), UpdatedAt: convert.FormatTimePtr(order.UpdatedAt),
		DeletedAt: convert.TimeToWrappers(order.DeletedAt),
	}
}

func mapResponsesOrder(orders []*repository.OrderResult) []*pb.OrderResponse {
	var mappedOrders []*pb.OrderResponse
	for _, order := range orders {
		if order == nil {
			continue
		}
		mappedOrders = append(mappedOrders, &pb.OrderResponse{
			Id: order.OrderID, MerchantId: order.MerchantID, CashierId: order.CashierID,
			TotalPrice: int32(order.TotalPrice), CreatedAt: order.CreatedAt, UpdatedAt: order.UpdatedAt,
		})
	}
	return mappedOrders
}

func mapResponsesOrderActive(orders []*repository.OrderResultDeleteAt) []*pb.OrderResponseDeleteAt {
	var mappedOrders []*pb.OrderResponseDeleteAt
	for _, order := range orders {
		if order == nil {
			continue
		}
		mappedOrders = append(mappedOrders, &pb.OrderResponseDeleteAt{
			Id: order.OrderID, MerchantId: order.MerchantID, CashierId: order.CashierID,
			TotalPrice: int32(order.TotalPrice), CreatedAt: order.CreatedAt, UpdatedAt: order.UpdatedAt,
			DeletedAt: convert.StrValToWrappers(&order.DeletedAt),
		})
	}
	return mappedOrders
}

func mapResponsesOrderTrashed(orders []*repository.OrderResultDeleteAt) []*pb.OrderResponseDeleteAt {
	return mapResponsesOrderActive(orders)
}
