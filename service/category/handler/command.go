package handler

import (
	"context"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	category_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *categoryQueryHandleGrpc) Create(ctx context.Context, request *pb.CreateCategoryRequest) (*pb.ApiResponseCategory, error) {
	req := &requests.CreateCategoryRequest{
		Name:        request.GetName(),
		Description: request.GetDescription(),
	}
	if err := req.Validate(); err != nil {
		return nil, category_errors.ErrGrpcValidateCreateCategory
	}
	category, err := s.categoryCommandService.CreateCategory(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategory{Status: "success", Message: "Successfully created category", Data: mapResponseCategory(category)}, nil
}

func (s *categoryQueryHandleGrpc) Update(ctx context.Context, request *pb.UpdateCategoryRequest) (*pb.ApiResponseCategory, error) {
	id := int(request.GetCategoryId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}
	req := &requests.UpdateCategoryRequest{
		CategoryID:  &id,
		Name:        request.GetName(),
		Description: request.GetDescription(),
	}
	if err := req.Validate(); err != nil {
		return nil, category_errors.ErrGrpcValidateUpdateCategory
	}
	category, err := s.categoryCommandService.UpdateCategory(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategory{Status: "success", Message: "Successfully updated category", Data: mapResponseCategory(category)}, nil
}

func (s *categoryQueryHandleGrpc) TrashedCategory(ctx context.Context, request *pb.FindByIdCategoryRequest) (*pb.ApiResponseCategoryDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}
	category, err := s.categoryCommandService.TrashedCategory(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategoryDeleteAt{Status: "success", Message: "Successfully trashed category", Data: mapResponseCategoryDeleteAt(category)}, nil
}

func (s *categoryQueryHandleGrpc) RestoreCategory(ctx context.Context, request *pb.FindByIdCategoryRequest) (*pb.ApiResponseCategoryDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}
	category, err := s.categoryCommandService.RestoreCategory(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategoryDeleteAt{Status: "success", Message: "Successfully restored category", Data: mapResponseCategoryDeleteAt(category)}, nil
}

func (s *categoryQueryHandleGrpc) DeleteCategoryPermanent(ctx context.Context, request *pb.FindByIdCategoryRequest) (*pb.ApiResponseCategoryDelete, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}
	_, err := s.categoryCommandService.DeleteCategoryPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategoryDelete{Status: "success", Message: "Successfully deleted category permanently"}, nil
}

func (s *categoryQueryHandleGrpc) RestoreAllCategory(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseCategoryAll, error) {
	_, err := s.categoryCommandService.RestoreAllCategories(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategoryAll{Status: "success", Message: "Successfully restore all category"}, nil
}

func (s *categoryQueryHandleGrpc) DeleteAllCategoryPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseCategoryAll, error) {
	_, err := s.categoryCommandService.DeleteAllCategoriesPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCategoryAll{Status: "success", Message: "Successfully delete category permanen"}, nil
}
