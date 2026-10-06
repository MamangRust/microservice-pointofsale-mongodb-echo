package response_api

import (
	orderitempb "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type orderItemResponseMapper struct {
}

func NewOrderItemResponseMapper() *orderItemResponseMapper {
	return &orderItemResponseMapper{}
}

func (o *orderItemResponseMapper) ToResponseOrderItem(orderItem *orderitempb.OrderItemResponse) *response.OrderItemResponse {
	return &response.OrderItemResponse{
		ID:        int(orderItem.Id),
		OrderID:   int(orderItem.OrderId),
		ProductID: int(orderItem.ProductId),
		Quantity:  int(orderItem.Quantity),
		Price:     int(orderItem.Price),
		CreatedAt: orderItem.CreatedAt,
		UpdatedAt: orderItem.UpdatedAt,
	}
}

func (o *orderItemResponseMapper) ToResponsesOrderItem(orderItems []*orderitempb.OrderItemResponse) []*response.OrderItemResponse {
	var mappedOrderItems []*response.OrderItemResponse

	for _, orderItem := range orderItems {
		mappedOrderItems = append(mappedOrderItems, o.ToResponseOrderItem(orderItem))
	}

	return mappedOrderItems
}

func (o *orderItemResponseMapper) ToResponseOrderItemDeleteAt(orderItem *orderitempb.OrderItemResponseDeleteAt) *response.OrderItemResponseDeleteAt {
	var deletedAt string
	if orderItem.DeletedAt != nil {
		deletedAt = orderItem.DeletedAt.Value
	}
	return &response.OrderItemResponseDeleteAt{
		ID:        int(orderItem.Id),
		OrderID:   int(orderItem.OrderId),
		ProductID: int(orderItem.ProductId),
		Quantity:  int(orderItem.Quantity),
		Price:     int(orderItem.Price),
		CreatedAt: orderItem.CreatedAt,
		UpdatedAt: orderItem.UpdatedAt,
		DeleteAt:  &deletedAt,
	}
}

func (o *orderItemResponseMapper) ToResponsesOrderItemDeleteAt(orderItems []*orderitempb.OrderItemResponseDeleteAt) []*response.OrderItemResponseDeleteAt {
	var mappedOrderItems []*response.OrderItemResponseDeleteAt

	for _, orderItem := range orderItems {
		mappedOrderItems = append(mappedOrderItems, o.ToResponseOrderItemDeleteAt(orderItem))
	}

	return mappedOrderItems
}

func (o *orderItemResponseMapper) ToApiResponseOrderItem(pbResponse *orderitempb.ApiResponseOrderItem) *response.ApiResponseOrderItem {
	return &response.ApiResponseOrderItem{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    o.ToResponseOrderItem(pbResponse.Data),
	}
}

func (o *orderItemResponseMapper) ToApiResponsesOrderItem(pbResponse *orderitempb.ApiResponsesOrderItem) *response.ApiResponsesOrderItem {
	return &response.ApiResponsesOrderItem{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    o.ToResponsesOrderItem(pbResponse.Data),
	}
}

func (o *orderItemResponseMapper) ToApiResponseOrderItemDelete(pbResponse *orderitempb.ApiResponseOrderItemDelete) *response.ApiResponseOrderItemDelete {
	return &response.ApiResponseOrderItemDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (o *orderItemResponseMapper) ToApiResponseOrderItemAll(pbResponse *orderitempb.ApiResponseOrderItemAll) *response.ApiResponseOrderItemAll {
	return &response.ApiResponseOrderItemAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (o *orderItemResponseMapper) ToApiResponsePaginationOrderItemDeleteAt(pbResponse *orderitempb.ApiResponsePaginationOrderItemDeleteAt) *response.ApiResponsePaginationOrderItemDeleteAt {
	return &response.ApiResponsePaginationOrderItemDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       o.ToResponsesOrderItemDeleteAt(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}

func (o *orderItemResponseMapper) ToApiResponsePaginationOrderItem(pbResponse *orderitempb.ApiResponsePaginationOrderItem) *response.ApiResponsePaginationOrderItem {
	return &response.ApiResponsePaginationOrderItem{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       o.ToResponsesOrderItem(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}
