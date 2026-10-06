package protomapper

import (
	commonpb "github.com/MamangRust/microservice-point-of-sale-pb/common"
	merchantpb "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type merchantProtoMapper struct{}

func NewMerchantProtoMaper() *merchantProtoMapper {
	return &merchantProtoMapper{}
}

func (m *merchantProtoMapper) ToProtoResponseMerchant(status string, message string, pbResponse *response.MerchantResponse) *merchantpb.ApiResponseMerchant {
	return &merchantpb.ApiResponseMerchant{
		Status:  status,
		Message: message,
		Data:    m.mapResponseMerchant(pbResponse),
	}
}

func (m *merchantProtoMapper) ToProtoResponsesMerchant(status string, message string, pbResponse []*response.MerchantResponse) *merchantpb.ApiResponsesMerchant {
	return &merchantpb.ApiResponsesMerchant{
		Status:  status,
		Message: message,
		Data:    m.mapResponsesMerchant(pbResponse),
	}
}

func (m *merchantProtoMapper) ToProtoResponseMerchantDeleteAt(status string, message string, pbResponse *response.MerchantResponseDeleteAt) *merchantpb.ApiResponseMerchantDeleteAt {
	return &merchantpb.ApiResponseMerchantDeleteAt{
		Status:  status,
		Message: message,
		Data:    m.mapResponseMerchantDeleteAt(pbResponse),
	}
}

func (m *merchantProtoMapper) ToProtoResponseMerchantDelete(status string, message string) *merchantpb.ApiResponseMerchantDelete {
	return &merchantpb.ApiResponseMerchantDelete{
		Status:  status,
		Message: message,
	}
}

func (m *merchantProtoMapper) ToProtoResponseMerchantAll(status string, message string) *merchantpb.ApiResponseMerchantAll {
	return &merchantpb.ApiResponseMerchantAll{
		Status:  status,
		Message: message,
	}
}

func (m *merchantProtoMapper) ToProtoResponsePaginationMerchantDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, merchants []*response.MerchantResponseDeleteAt) *merchantpb.ApiResponsePaginationMerchantDeleteAt {
	return &merchantpb.ApiResponsePaginationMerchantDeleteAt{
		Status:     status,
		Message:    message,
		Data:       m.mapResponsesMerchantDeleteAt(merchants),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (m *merchantProtoMapper) ToProtoResponsePaginationMerchant(pagination *commonpb.PaginationMeta, status string, message string, merchants []*response.MerchantResponse) *merchantpb.ApiResponsePaginationMerchant {
	return &merchantpb.ApiResponsePaginationMerchant{
		Status:     status,
		Message:    message,
		Data:       m.mapResponsesMerchant(merchants),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (m *merchantProtoMapper) mapResponseMerchant(merchant *response.MerchantResponse) *merchantpb.MerchantResponse {
	return &merchantpb.MerchantResponse{
		Id:           int32(merchant.ID),
		UserId:       int32(merchant.UserID),
		Name:         merchant.Name,
		Description:  merchant.Description,
		Address:      merchant.Address,
		ContactEmail: merchant.ContactEmail,
		ContactPhone: merchant.ContactPhone,
		Status:       merchant.Status,
		CreatedAt:    merchant.CreatedAt,
		UpdatedAt:    merchant.UpdatedAt,
	}
}

func (m *merchantProtoMapper) mapResponsesMerchant(merchants []*response.MerchantResponse) []*merchantpb.MerchantResponse {
	var mappedMerchants []*merchantpb.MerchantResponse

	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.mapResponseMerchant(merchant))
	}

	return mappedMerchants
}

func (m *merchantProtoMapper) mapResponseMerchantDeleteAt(merchant *response.MerchantResponseDeleteAt) *merchantpb.MerchantResponseDeleteAt {
	return &merchantpb.MerchantResponseDeleteAt{
		Id:           int32(merchant.ID),
		UserId:       int32(merchant.UserID),
		Name:         merchant.Name,
		Description:  merchant.Description,
		Address:      merchant.Address,
		ContactEmail: merchant.ContactEmail,
		ContactPhone: merchant.ContactPhone,
		Status:       merchant.Status,
		CreatedAt:    merchant.CreatedAt,
		UpdatedAt:    merchant.UpdatedAt,
		DeletedAt:    merchant.DeletedAt,
	}
}

func (m *merchantProtoMapper) mapResponsesMerchantDeleteAt(merchants []*response.MerchantResponseDeleteAt) []*merchantpb.MerchantResponseDeleteAt {
	var mappedMerchants []*merchantpb.MerchantResponseDeleteAt

	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.mapResponseMerchantDeleteAt(merchant))
	}

	return mappedMerchants
}
