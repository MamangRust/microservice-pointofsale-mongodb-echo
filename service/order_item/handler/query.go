package handler

import (
	"context"
	"math"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	"github.com/MamangRust/microservice-point-of-sale-order-item/service"
	"github.com/MamangRust/microservice-point-of-sale-pb/common"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"

	"github.com/MamangRust/microservice-point-of-sale-shared/convert"
	"go.uber.org/zap"
)

type orderItemQueryHandleGrpc struct {
	pborderitem.UnimplementedOrderItemServiceServer
	orderItemService service.OrderItemQueryService
	logger           logger.LoggerInterface
}

func NewOrderItemQueryHandleGrpc(
	orderItemService service.OrderItemQueryService,
	logger logger.LoggerInterface,
) pborderitem.OrderItemServiceServer {
	return &orderItemQueryHandleGrpc{
		orderItemService: orderItemService,
		logger:           logger,
	}
}

func (s *orderItemQueryHandleGrpc) FindAll(ctx context.Context, request *pborderitem.FindAllOrderItemRequest) (*pborderitem.ApiResponsePaginationOrderItem, error) {
	s.logger.Info("FindAll order items called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllOrderItems{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orderItems, totalRecords, err := s.orderItemService.FindAllOrderItems(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll order items failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAll order items success")

	return &pborderitem.ApiResponsePaginationOrderItem{
		Status:     "success",
		Message:    "Successfully fetched order items",
		Data:       mapResponsesOrderItem(orderItems),
		Pagination: mapOrderItemPaginationMeta(paginationMeta),
	}, nil
}

func (s *orderItemQueryHandleGrpc) FindByActive(ctx context.Context, request *pborderitem.FindAllOrderItemRequest) (*pborderitem.ApiResponsePaginationOrderItemDeleteAt, error) {
	s.logger.Info("FindByActive order items called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllOrderItems{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orderItems, totalRecords, err := s.orderItemService.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive order items failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByActive order items success")

	return &pborderitem.ApiResponsePaginationOrderItemDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active order items",
		Data:       mapResponsesOrderItemActive(orderItems),
		Pagination: mapOrderItemPaginationMeta(paginationMeta),
	}, nil
}

func (s *orderItemQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pborderitem.FindAllOrderItemRequest) (*pborderitem.ApiResponsePaginationOrderItemDeleteAt, error) {
	s.logger.Info("FindByTrashed order items called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllOrderItems{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orderItems, totalRecords, err := s.orderItemService.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed order items failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByTrashed order items success")

	return &pborderitem.ApiResponsePaginationOrderItemDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed order items",
		Data:       mapResponsesOrderItemTrashed(orderItems),
		Pagination: mapOrderItemPaginationMeta(paginationMeta),
	}, nil
}

func (s *orderItemQueryHandleGrpc) FindOrderItemByOrder(ctx context.Context, request *pborderitem.FindByIdOrderItemRequest) (*pborderitem.ApiResponsesOrderItem, error) {
	s.logger.Info("FindOrderItemByOrder called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}

	orderItems, err := s.orderItemService.FindOrderItemByOrder(ctx, id)
	if err != nil {
		s.logger.Error("FindOrderItemByOrder failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindOrderItemByOrder success")

	return &pborderitem.ApiResponsesOrderItem{
		Status:  "success",
		Message: "Successfully fetched order items by order",
		Data:    mapResponsesOrderItemFromModel(orderItems),
	}, nil
}

// Map helpers

func mapOrderItemPaginationMeta(meta *common.PaginationMeta) *common.PaginationMeta {
	if meta == nil {
		return nil
	}
	return &common.PaginationMeta{
		CurrentPage:  meta.CurrentPage,
		PageSize:     meta.PageSize,
		TotalPages:   meta.TotalPages,
		TotalRecords: meta.TotalRecords,
	}
}

func fmtOIStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func fmtOITimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return convert.FormatTimePtr(t)
}

func mapOrderItemResponse(item *repository.OrderItemResult) *pborderitem.OrderItemResponse {
	if item == nil {
		return nil
	}
	return &pborderitem.OrderItemResponse{
		Id: item.OrderItemID, OrderId: item.OrderID, ProductId: item.ProductID,
		Quantity: item.Quantity, Price: int32(item.Price),
		CreatedAt: fmtOIStr(item.CreatedAt), UpdatedAt: fmtOIStr(item.UpdatedAt),
	}
}

func mapResponsesOrderItem(items []*repository.OrderItemResult) []*pborderitem.OrderItemResponse {
	var res []*pborderitem.OrderItemResponse
	for _, i := range items {
		res = append(res, mapOrderItemResponse(i))
	}
	return res
}

func mapOrderItemDeleteAt(item *repository.OrderItemResultDeleteAt) *pborderitem.OrderItemResponseDeleteAt {
	if item == nil {
		return nil
	}
	return &pborderitem.OrderItemResponseDeleteAt{
		Id: item.OrderItemID, OrderId: item.OrderID, ProductId: item.ProductID,
		Quantity: item.Quantity, Price: int32(item.Price),
		CreatedAt: fmtOIStr(item.CreatedAt), UpdatedAt: fmtOIStr(item.UpdatedAt),
		DeletedAt: convert.StrValToWrappers(item.DeletedAt),
	}
}

func mapResponsesOrderItemActive(items []*repository.OrderItemResultDeleteAt) []*pborderitem.OrderItemResponseDeleteAt {
	var res []*pborderitem.OrderItemResponseDeleteAt
	for _, i := range items {
		res = append(res, mapOrderItemDeleteAt(i))
	}
	return res
}

func mapResponsesOrderItemTrashed(items []*repository.OrderItemResultDeleteAt) []*pborderitem.OrderItemResponseDeleteAt {
	return mapResponsesOrderItemActive(items)
}

func mapOrderItemModel(item *models.OrderItem) *pborderitem.OrderItemResponse {
	if item == nil {
		return nil
	}
	return &pborderitem.OrderItemResponse{
		Id: item.OrderItemID, OrderId: item.OrderID, ProductId: item.ProductID,
		Quantity: item.Quantity, Price: int32(item.Price),
		CreatedAt: fmtOITimePtr(item.CreatedAt), UpdatedAt: fmtOITimePtr(item.UpdatedAt),
	}
}

func mapResponsesOrderItemFromModel(items []*models.OrderItem) []*pborderitem.OrderItemResponse {
	var res []*pborderitem.OrderItemResponse
	for _, i := range items {
		res = append(res, mapOrderItemModel(i))
	}
	return res
}
