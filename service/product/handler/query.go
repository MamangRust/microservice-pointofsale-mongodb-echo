package handler

import (
	"context"
	"math"

	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pb "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-product/repository"
	"github.com/MamangRust/microservice-point-of-sale-product/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	product_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
)

type productQueryHandleGrpc struct {
	pb.UnimplementedProductServiceServer

	productQueryService service.ProductQueryService

	productCommandService service.ProductCommandService
}

func NewProductHandleGrpc(query service.ProductQueryService, command service.ProductCommandService) *productQueryHandleGrpc {
	return &productQueryHandleGrpc{
		productQueryService:   query,
		productCommandService: command,
	}
}

func (s *productQueryHandleGrpc) FindAll(ctx context.Context, request *pb.FindAllProductRequest) (*pb.ApiResponsePaginationProduct, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllProducts{Page: page, PageSize: pageSize, Search: search}
	products, totalRecords, err := s.productQueryService.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationProduct{
		Status: "success", Message: "Successfully fetched product",
		Data: mapResponsesProduct(products), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *productQueryHandleGrpc) FindByMerchant(ctx context.Context, request *pb.FindAllProductMerchantRequest) (*pb.ApiResponsePaginationProduct, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	merchant_id := int(request.GetMerchantId())
	min_price := int(request.GetMinPrice())
	max_price := int(request.GetMaxPrice())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if min_price <= 0 {
		min_price = 0
	}
	if max_price <= 0 {
		max_price = 0
	}
	reqService := requests.ProductByMerchantRequest{
		MerchantID: merchant_id, Page: page, PageSize: pageSize, Search: search,
		MinPrice: &min_price, MaxPrice: &max_price,
	}
	products, totalRecords, err := s.productQueryService.FindByMerchant(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationProduct{
		Status: "success", Message: "Successfully fetched product",
		Data: mapResponsesProductByMerchant(products, int32(merchant_id)), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *productQueryHandleGrpc) FindByCategory(ctx context.Context, request *pb.FindAllProductCategoryRequest) (*pb.ApiResponsePaginationProduct, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	category_name := request.GetCategoryName()
	min_price := int(request.GetMinprice())
	max_price := int(request.GetMaxprice())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if min_price <= 0 {
		min_price = 0
	}
	if max_price <= 0 {
		max_price = 0
	}
	reqService := requests.ProductByCategoryRequest{
		Page: page, PageSize: pageSize, Search: search, CategoryName: category_name,
		MinPrice: &min_price, MaxPrice: &max_price,
	}
	products, totalRecords, err := s.productQueryService.FindByCategory(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationProduct{
		Status: "success", Message: "Successfully fetched product",
		Data: mapResponsesProductByCategory(products), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *productQueryHandleGrpc) FindById(ctx context.Context, request *pb.FindByIdProductRequest) (*pb.ApiResponseProduct, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}
	product, err := s.productQueryService.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProduct{Status: "success", Message: "Successfully fetched product", Data: mapResponseProduct(product)}, nil
}

func (s *productQueryHandleGrpc) FindByActive(ctx context.Context, request *pb.FindAllProductRequest) (*pb.ApiResponsePaginationProductDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllProducts{Page: page, PageSize: pageSize, Search: search}
	products, totalRecords, err := s.productQueryService.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationProductDeleteAt{
		Status: "success", Message: "Successfully fetched active product",
		Data: mapResponsesProductActive(products), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *productQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pb.FindAllProductRequest) (*pb.ApiResponsePaginationProductDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllProducts{Page: page, PageSize: pageSize, Search: search}
	products, totalRecords, err := s.productQueryService.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationProductDeleteAt{
		Status: "success", Message: "Successfully fetched trashed product",
		Data: mapResponsesProductTrashed(products), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func mapResponseProduct(product *models.Product) *pb.ProductResponse {
	if product == nil {
		return nil
	}
	return &pb.ProductResponse{
		Id: int32(product.ProductID), MerchantId: int32(product.MerchantID), CategoryId: int32(product.CategoryID),
		Name: product.Name, Description: convert.StrVal(product.Description),
		Price: int32(product.Price), CountInStock: int32(product.CountInStock),
		Brand: convert.StrVal(product.Brand), Weight: convert.Int32Val(product.Weight),
		SlugProduct: convert.StrVal(product.SlugProduct), ImageProduct: convert.StrVal(product.ImageProduct),
		Barcode:   convert.StrVal(product.Barcode),
		CreatedAt: convert.FormatTimePtr(product.CreatedAt), UpdatedAt: convert.FormatTimePtr(product.UpdatedAt),
	}
}

func mapResponseProductDeleteAt(product *models.Product) *pb.ProductResponseDeleteAt {
	if product == nil {
		return nil
	}
	return &pb.ProductResponseDeleteAt{
		Id: int32(product.ProductID), MerchantId: int32(product.MerchantID), CategoryId: int32(product.CategoryID),
		Name: product.Name, Description: convert.StrVal(product.Description),
		Price: int32(product.Price), CountInStock: int32(product.CountInStock),
		Brand: convert.StrVal(product.Brand), Weight: convert.Int32Val(product.Weight),
		SlugProduct: convert.StrVal(product.SlugProduct), ImageProduct: convert.StrVal(product.ImageProduct),
		Barcode:   convert.StrVal(product.Barcode),
		CreatedAt: convert.FormatTimePtr(product.CreatedAt), UpdatedAt: convert.FormatTimePtr(product.UpdatedAt),
		DeletedAt: convert.TimeToWrappers(product.DeletedAt),
	}
}

func mapResponsesProduct(products []*repository.ProductResult) []*pb.ProductResponse {
	var mappedProducts []*pb.ProductResponse
	for _, p := range products {
		if p == nil {
			continue
		}
		mappedProducts = append(mappedProducts, &pb.ProductResponse{
			Id: int32(p.ProductID), MerchantId: int32(p.MerchantID), CategoryId: int32(p.CategoryID),
			Name: p.Name, Description: convert.StrVal(p.Description),
			Price: int32(p.Price), CountInStock: int32(p.CountInStock),
			Brand: convert.StrVal(p.Brand), Weight: convert.Int32Val(p.Weight),
			SlugProduct: convert.StrVal(p.SlugProduct), ImageProduct: convert.StrVal(p.ImageProduct),
			Barcode: convert.StrVal(p.Barcode), CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
		})
	}
	return mappedProducts
}

func mapResponsesProductByMerchant(products []*repository.ProductByMerchantResult, merchantId int32) []*pb.ProductResponse {
	var mappedProducts []*pb.ProductResponse
	for _, p := range products {
		if p == nil {
			continue
		}
		mappedProducts = append(mappedProducts, &pb.ProductResponse{
			Id: int32(p.ProductID), MerchantId: merchantId,
			Name: p.Name, Description: convert.StrVal(p.Description),
			Price: int32(p.Price), CountInStock: int32(p.CountInStock),
			Brand: convert.StrVal(p.Brand), ImageProduct: convert.StrVal(p.ImageProduct),
			CreatedAt: p.CreatedAt,
		})
	}
	return mappedProducts
}

func mapResponsesProductByCategory(products []*repository.ProductByCategoryResult) []*pb.ProductResponse {
	var mappedProducts []*pb.ProductResponse
	for _, p := range products {
		if p == nil {
			continue
		}
		mappedProducts = append(mappedProducts, &pb.ProductResponse{
			Id: int32(p.ProductID), MerchantId: int32(p.MerchantID), CategoryId: int32(p.CategoryID),
			Name: p.Name, Description: convert.StrVal(p.Description),
			Price: int32(p.Price), CountInStock: int32(p.CountInStock),
			Brand: convert.StrVal(p.Brand), Weight: convert.Int32Val(p.Weight),
			SlugProduct: convert.StrVal(p.SlugProduct), ImageProduct: convert.StrVal(p.ImageProduct),
			Barcode: convert.StrVal(p.Barcode), CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
		})
	}
	return mappedProducts
}

func mapResponsesProductActive(products []*repository.ProductResultDeleteAt) []*pb.ProductResponseDeleteAt {
	var mappedProducts []*pb.ProductResponseDeleteAt
	for _, p := range products {
		if p == nil {
			continue
		}
		mappedProducts = append(mappedProducts, &pb.ProductResponseDeleteAt{
			Id: int32(p.ProductID), MerchantId: int32(p.MerchantID), CategoryId: int32(p.CategoryID),
			Name: p.Name, Description: convert.StrVal(p.Description),
			Price: int32(p.Price), CountInStock: int32(p.CountInStock),
			Brand: convert.StrVal(p.Brand), Weight: convert.Int32Val(p.Weight),
			SlugProduct: convert.StrVal(p.SlugProduct), ImageProduct: convert.StrVal(p.ImageProduct),
			Barcode: convert.StrVal(p.Barcode), CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
			DeletedAt: convert.StrValToWrappers(&p.DeletedAt),
		})
	}
	return mappedProducts
}

func mapResponsesProductTrashed(products []*repository.ProductResultDeleteAt) []*pb.ProductResponseDeleteAt {
	return mapResponsesProductActive(products)
}
