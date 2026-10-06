package handler

import (
	"context"
	"math"

	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"

	"github.com/MamangRust/microservice-point-of-sale-category/repository"
	"github.com/MamangRust/microservice-point-of-sale-category/service"
	pb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	category_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
)

type categoryQueryHandleGrpc struct {
	pb.UnimplementedCategoryServiceServer

	categoryQuery service.CategoryQueryService

	categoryCommandService service.CategoryCommandService
}

func NewCategoryHandleGrpc(query service.CategoryQueryService, command service.CategoryCommandService) *categoryQueryHandleGrpc {
	return &categoryQueryHandleGrpc{
		categoryQuery:          query,
		categoryCommandService: command,
	}
}

func (s *categoryQueryHandleGrpc) FindAll(ctx context.Context, request *pb.FindAllCategoryRequest) (*pb.ApiResponsePaginationCategory, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllCategory{Page: page, PageSize: pageSize, Search: search}
	category, totalRecords, err := s.categoryQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationCategory{
		Status: "success", Message: "Successfully fetched categories",
		Data: mapResponsesCategory(category), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *categoryQueryHandleGrpc) FindById(ctx context.Context, request *pb.FindByIdCategoryRequest) (*pb.ApiResponseCategory, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}
	category, err := s.categoryQuery.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategory{Status: "success", Message: "Successfully fetched category", Data: mapResponseCategory(category)}, nil
}

func (s *categoryQueryHandleGrpc) FindByName(ctx context.Context, request *pb.FindByNameCategoryRequest) (*pb.ApiResponseCategory, error) {
	name := request.GetName()
	if name == "" {
		return nil, category_errors.ErrGrpcFailedInvalidName
	}
	category, err := s.categoryQuery.FindByName(ctx, name)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategory{Status: "success", Message: "Successfully fetched category", Data: mapResponseCategory(category)}, nil
}

func (s *categoryQueryHandleGrpc) FindByIds(ctx context.Context, request *pb.FindByIdsCategoryRequest) (*pb.ApiResponsesCategory, error) {
	ids := make([]int, 0, len(request.GetIds()))
	for _, id := range request.GetIds() {
		if id <= 0 {
			return nil, category_errors.ErrGrpcFailedInvalidId
		}
		ids = append(ids, int(id))
	}

	categories, err := s.categoryQuery.FindByIds(ctx, ids)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponsesCategory{Status: "success", Message: "Successfully fetched categories", Data: mapResponsesCategoryModels(categories)}, nil
}

func (s *categoryQueryHandleGrpc) FindByActive(ctx context.Context, request *pb.FindAllCategoryRequest) (*pb.ApiResponsePaginationCategoryDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllCategory{Page: page, PageSize: pageSize, Search: search}
	categories, totalRecords, err := s.categoryQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationCategoryDeleteAt{
		Status: "success", Message: "Successfully fetched active categories",
		Data: mapResponsesCategoryActive(categories), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *categoryQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pb.FindAllCategoryRequest) (*pb.ApiResponsePaginationCategoryDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllCategory{Page: page, PageSize: pageSize, Search: search}
	categories, totalRecords, err := s.categoryQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationCategoryDeleteAt{
		Status: "success", Message: "Successfully fetched trashed categories",
		Data: mapResponsesCategoryTrashed(categories), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func mapResponseCategory(category *models.Category) *pb.CategoryResponse {
	if category == nil {
		return nil
	}
	return &pb.CategoryResponse{
		Id:           category.CategoryID,
		Name:         category.Name,
		Description:  convert.StrVal(category.Description),
		SlugCategory: convert.StrVal(category.SlugCategory),
		CreatedAt:    convert.FormatTimePtr(category.CreatedAt),
		UpdatedAt:    convert.FormatTimePtr(category.UpdatedAt),
	}
}

func mapResponseGetCategory(category *repository.CategoryResult) *pb.CategoryResponse {
	if category == nil {
		return nil
	}
	return &pb.CategoryResponse{
		Id:           category.CategoryID,
		Name:         category.Name,
		Description:  convert.StrVal(category.Description),
		SlugCategory: convert.StrVal(category.SlugCategory),
		CreatedAt:    convert.StrVal(category.CreatedAt),
		UpdatedAt:    convert.StrVal(category.UpdatedAt),
	}
}

func mapResponsesCategory(categories []*repository.CategoryResult) []*pb.CategoryResponse {
	var res []*pb.CategoryResponse
	for _, c := range categories {
		res = append(res, mapResponseGetCategory(c))
	}
	return res
}

func mapResponsesCategoryModels(categories []*models.Category) []*pb.CategoryResponse {
	var res []*pb.CategoryResponse
	for _, c := range categories {
		res = append(res, mapResponseCategory(c))
	}
	return res
}

func mapResponseCategoryDeleteAt(category *models.Category) *pb.CategoryResponseDeleteAt {
	if category == nil {
		return nil
	}
	return &pb.CategoryResponseDeleteAt{
		Id:           category.CategoryID,
		Name:         category.Name,
		Description:  convert.StrVal(category.Description),
		SlugCategory: convert.StrVal(category.SlugCategory),
		CreatedAt:    convert.FormatTimePtr(category.CreatedAt),
		UpdatedAt:    convert.FormatTimePtr(category.UpdatedAt),
		DeletedAt:    convert.TimeToWrappers(category.DeletedAt),
	}
}

func mapResponseGetCategoryActive(category *repository.CategoryResultDeleteAt) *pb.CategoryResponseDeleteAt {
	if category == nil {
		return nil
	}
	return &pb.CategoryResponseDeleteAt{
		Id:           category.CategoryID,
		Name:         category.Name,
		Description:  convert.StrVal(category.Description),
		SlugCategory: convert.StrVal(category.SlugCategory),
		CreatedAt:    convert.StrVal(category.CreatedAt),
		UpdatedAt:    convert.StrVal(category.UpdatedAt),
		DeletedAt:    convert.StrValToWrappers(category.DeletedAt),
	}
}

func mapResponsesCategoryActive(categories []*repository.CategoryResultDeleteAt) []*pb.CategoryResponseDeleteAt {
	var res []*pb.CategoryResponseDeleteAt
	for _, c := range categories {
		res = append(res, mapResponseGetCategoryActive(c))
	}
	return res
}

func mapResponseGetCategoryTrashed(category *repository.CategoryResultDeleteAt) *pb.CategoryResponseDeleteAt {
	if category == nil {
		return nil
	}
	return &pb.CategoryResponseDeleteAt{
		Id:           category.CategoryID,
		Name:         category.Name,
		Description:  convert.StrVal(category.Description),
		SlugCategory: convert.StrVal(category.SlugCategory),
		CreatedAt:    convert.StrVal(category.CreatedAt),
		UpdatedAt:    convert.StrVal(category.UpdatedAt),
		DeletedAt:    convert.StrValToWrappers(category.DeletedAt),
	}
}

func mapResponsesCategoryTrashed(categories []*repository.CategoryResultDeleteAt) []*pb.CategoryResponseDeleteAt {
	var res []*pb.CategoryResponseDeleteAt
	for _, c := range categories {
		res = append(res, mapResponseGetCategoryTrashed(c))
	}
	return res
}
