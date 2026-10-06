package response_api

import (
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type cashierResponseMapper struct{}

func NewCashierResponseMapper() *cashierResponseMapper {
	return &cashierResponseMapper{}
}

func (c *cashierResponseMapper) ToResponseCashier(cashier *cashierpb.CashierResponse) *response.CashierResponse {
	return &response.CashierResponse{
		ID:         int(cashier.Id),
		MerchantID: int(cashier.MerchantId),
		Name:       cashier.Name,
		CreatedAt:  cashier.CreatedAt,
		UpdatedAt:  cashier.UpdatedAt,
	}
}

func (c *cashierResponseMapper) ToResponsesCashier(cashiers []*cashierpb.CashierResponse) []*response.CashierResponse {
	var mappedCashiers []*response.CashierResponse

	for _, cashier := range cashiers {
		mappedCashiers = append(mappedCashiers, c.ToResponseCashier(cashier))
	}

	return mappedCashiers
}

func (c *cashierResponseMapper) ToResponseCashierDeleteAt(cashier *cashierpb.CashierResponseDeleteAt) *response.CashierResponseDeleteAt {
	var deletedAt string
	if cashier.DeletedAt != nil {
		deletedAt = cashier.DeletedAt.Value
	}

	return &response.CashierResponseDeleteAt{
		ID:         int(cashier.Id),
		MerchantID: int(cashier.MerchantId),
		Name:       cashier.Name,
		CreatedAt:  cashier.CreatedAt,
		UpdatedAt:  cashier.UpdatedAt,
		DeletedAt:  &deletedAt,
	}
}

func (c *cashierResponseMapper) ToResponsesCashierDeleteAt(cashiers []*cashierpb.CashierResponseDeleteAt) []*response.CashierResponseDeleteAt {
	var mappedCashiers []*response.CashierResponseDeleteAt

	for _, cashier := range cashiers {
		mappedCashiers = append(mappedCashiers, c.ToResponseCashierDeleteAt(cashier))
	}

	return mappedCashiers
}

func (c *cashierResponseMapper) ToApiResponseCashier(pbResponse *cashierpb.ApiResponseCashier) *response.ApiResponseCashier {
	return &response.ApiResponseCashier{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponseCashier(pbResponse.Data),
	}
}

func (c *cashierResponseMapper) ToApiResponsesCashier(pbResponse *cashierpb.ApiResponsesCashier) *response.ApiResponsesCashier {
	return &response.ApiResponsesCashier{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponsesCashier(pbResponse.Data),
	}
}

func (c *cashierResponseMapper) ToApiResponseCashierDeleteAt(pbResponse *cashierpb.ApiResponseCashierDeleteAt) *response.ApiResponseCashierDeleteAt {
	return &response.ApiResponseCashierDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponseCashierDeleteAt(pbResponse.Data),
	}
}

func (c *cashierResponseMapper) ToApiResponseCashierDelete(pbResponse *cashierpb.ApiResponseCashierDelete) *response.ApiResponseCashierDelete {
	return &response.ApiResponseCashierDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (c *cashierResponseMapper) ToApiResponseCashierAll(pbResponse *cashierpb.ApiResponseCashierAll) *response.ApiResponseCashierAll {
	return &response.ApiResponseCashierAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (c *cashierResponseMapper) ToApiResponsePaginationCashierDeleteAt(pbResponse *cashierpb.ApiResponsePaginationCashierDeleteAt) *response.ApiResponsePaginationCashierDeleteAt {
	return &response.ApiResponsePaginationCashierDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       c.ToResponsesCashierDeleteAt(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}

func (c *cashierResponseMapper) ToApiResponsePaginationCashier(pbResponse *cashierpb.ApiResponsePaginationCashier) *response.ApiResponsePaginationCashier {
	return &response.ApiResponsePaginationCashier{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       c.ToResponsesCashier(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}
