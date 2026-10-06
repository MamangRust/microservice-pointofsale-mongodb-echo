package protomapper

import (
	commonpb "github.com/MamangRust/microservice-point-of-sale-pb/common"
	productpb "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type productProtoMapper struct{}

func NewProductProtoMapper() *productProtoMapper {
	return &productProtoMapper{}
}

func (p *productProtoMapper) ToProtoResponseProduct(status string, message string, pbResponse *response.ProductResponse) *productpb.ApiResponseProduct {
	return &productpb.ApiResponseProduct{
		Status:  status,
		Message: message,
		Data:    p.mapResponseProduct(pbResponse),
	}
}

func (p *productProtoMapper) ToProtoResponsesProduct(status string, message string, pbResponse []*response.ProductResponse) *productpb.ApiResponsesProduct {
	return &productpb.ApiResponsesProduct{
		Status:  status,
		Message: message,
		Data:    p.mapResponsesProduct(pbResponse),
	}
}

func (p *productProtoMapper) ToProtoResponseProductDeleteAt(status string, message string, pbResponse *response.ProductResponseDeleteAt) *productpb.ApiResponseProductDeleteAt {
	return &productpb.ApiResponseProductDeleteAt{
		Status:  status,
		Message: message,
		Data:    p.mapResponseProductDeleteAt(pbResponse),
	}
}

func (p *productProtoMapper) ToProtoResponseProductDelete(status string, message string) *productpb.ApiResponseProductDelete {
	return &productpb.ApiResponseProductDelete{
		Status:  status,
		Message: message,
	}
}

func (p *productProtoMapper) ToProtoResponseProductAll(status string, message string) *productpb.ApiResponseProductAll {
	return &productpb.ApiResponseProductAll{
		Status:  status,
		Message: message,
	}
}

func (p *productProtoMapper) ToProtoResponsePaginationProductDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, products []*response.ProductResponseDeleteAt) *productpb.ApiResponsePaginationProductDeleteAt {
	return &productpb.ApiResponsePaginationProductDeleteAt{
		Status:     status,
		Message:    message,
		Data:       p.mapResponsesProductDeleteAt(products),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (p *productProtoMapper) ToProtoResponsePaginationProduct(pagination *commonpb.PaginationMeta, status string, message string, products []*response.ProductResponse) *productpb.ApiResponsePaginationProduct {
	return &productpb.ApiResponsePaginationProduct{
		Status:     status,
		Message:    message,
		Data:       p.mapResponsesProduct(products),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (p *productProtoMapper) mapResponseProduct(product *response.ProductResponse) *productpb.ProductResponse {
	return &productpb.ProductResponse{
		Id:           int32(product.ID),
		MerchantId:   int32(product.MerchantID),
		CategoryId:   int32(product.CategoryID),
		Name:         product.Name,
		Description:  product.Description,
		Price:        int32(product.Price),
		CountInStock: int32(product.CountInStock),
		Brand:        product.Brand,
		Weight:       int32(product.Weight),
		SlugProduct:  product.SlugProduct,
		ImageProduct: product.ImageProduct,
		Barcode:      product.Barcode,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}
}

func (p *productProtoMapper) mapResponsesProduct(products []*response.ProductResponse) []*productpb.ProductResponse {
	var mappedProducts []*productpb.ProductResponse

	for _, product := range products {
		mappedProducts = append(mappedProducts, p.mapResponseProduct(product))
	}

	return mappedProducts
}

func (p *productProtoMapper) mapResponseProductDeleteAt(product *response.ProductResponseDeleteAt) *productpb.ProductResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if product.DeleteAt != nil {
		deletedAt = wrapperspb.String(*product.DeleteAt)
	}

	return &productpb.ProductResponseDeleteAt{
		Id:           int32(product.ID),
		MerchantId:   int32(product.MerchantID),
		CategoryId:   int32(product.CategoryID),
		Name:         product.Name,
		Description:  product.Description,
		Price:        int32(product.Price),
		CountInStock: int32(product.CountInStock),
		Brand:        product.Brand,
		Weight:       int32(product.Weight),
		SlugProduct:  product.SlugProduct,
		ImageProduct: product.ImageProduct,
		Barcode:      product.Barcode,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

func (p *productProtoMapper) mapResponsesProductDeleteAt(products []*response.ProductResponseDeleteAt) []*productpb.ProductResponseDeleteAt {
	var mappedProducts []*productpb.ProductResponseDeleteAt

	for _, product := range products {
		mappedProducts = append(mappedProducts, p.mapResponseProductDeleteAt(product))
	}

	return mappedProducts
}
