package handler

import (
	"context"
	"math"

	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pb "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	merchant_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_errors"
)

type merchantQueryHandleGrpc struct {
	pb.UnimplementedMerchantServiceServer

	merchantQuery service.MerchantQueryService

	merchantCommandService service.MerchantCommandService
}

func NewMerchantHandleGrpc(query service.MerchantQueryService, command service.MerchantCommandService) *merchantQueryHandleGrpc {
	return &merchantQueryHandleGrpc{
		merchantQuery:          query,
		merchantCommandService: command,
	}
}

func (s *merchantQueryHandleGrpc) FindAll(ctx context.Context, req *pb.FindAllMerchantRequest) (*pb.ApiResponsePaginationMerchant, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllMerchants{Page: page, PageSize: pageSize, Search: search}
	merchants, totalRecords, err := s.merchantQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationMerchant{
		Status: "success", Message: "Successfully fetched merchant record",
		Data: mapResponsesGetMerchantsRow(merchants), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *merchantQueryHandleGrpc) FindByActive(ctx context.Context, req *pb.FindAllMerchantRequest) (*pb.ApiResponsePaginationMerchantDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllMerchants{Page: page, PageSize: pageSize, Search: search}
	res, totalRecords, err := s.merchantQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationMerchantDeleteAt{
		Status: "success", Message: "Successfully fetched merchant record",
		Data: mapResponsesGetMerchantsActiveRow(res), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *merchantQueryHandleGrpc) FindByTrashed(ctx context.Context, req *pb.FindAllMerchantRequest) (*pb.ApiResponsePaginationMerchantDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllMerchants{Page: page, PageSize: pageSize, Search: search}
	res, totalRecords, err := s.merchantQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationMerchantDeleteAt{
		Status: "success", Message: "Successfully fetched merchant record",
		Data: mapResponsesGetMerchantsTrashedRow(res), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *merchantQueryHandleGrpc) FindById(ctx context.Context, request *pb.FindByIdMerchantRequest) (*pb.ApiResponseMerchant, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, merchant_errors.ErrGrpcInvalidID
	}
	merchant, err := s.merchantQuery.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchant{Status: "success", Message: "Successfully fetched merchant", Data: mapResponseMerchant(merchant)}, nil
}

func mapResponseMerchant(merchant *models.Merchant) *pb.MerchantResponse {
	if merchant == nil {
		return nil
	}
	return &pb.MerchantResponse{
		Id:           int32(merchant.MerchantID),
		UserId:       int32(merchant.UserID),
		Name:         merchant.Name,
		Description:  convert.StrVal(merchant.Description),
		Address:      convert.StrVal(merchant.Address),
		ContactEmail: convert.StrVal(merchant.ContactEmail),
		ContactPhone: convert.StrVal(merchant.ContactPhone),
		Status:       merchant.Status,
		CreatedAt:    convert.FormatTimePtr(merchant.CreatedAt),
		UpdatedAt:    convert.FormatTimePtr(merchant.UpdatedAt),
	}
}

func mapResponsesGetMerchantsRow(merchants []*repository.MerchantResult) []*pb.MerchantResponse {
	var mapped []*pb.MerchantResponse
	for _, m := range merchants {
		mapped = append(mapped, &pb.MerchantResponse{
			Id: int32(m.MerchantID), UserId: int32(m.UserID), Name: m.Name,
			Description: convert.StrVal(m.Description), Address: convert.StrVal(m.Address),
			ContactEmail: convert.StrVal(m.ContactEmail), ContactPhone: convert.StrVal(m.ContactPhone),
			Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		})
	}
	return mapped
}

func mapResponseMerchantDeleteAt(merchant *models.Merchant) *pb.MerchantResponseDeleteAt {
	if merchant == nil {
		return nil
	}
	return &pb.MerchantResponseDeleteAt{
		Id: int32(merchant.MerchantID), UserId: int32(merchant.UserID), Name: merchant.Name,
		Description: convert.StrVal(merchant.Description), Address: convert.StrVal(merchant.Address),
		ContactEmail: convert.StrVal(merchant.ContactEmail), ContactPhone: convert.StrVal(merchant.ContactPhone),
		Status: merchant.Status, CreatedAt: convert.FormatTimePtr(merchant.CreatedAt),
		UpdatedAt: convert.FormatTimePtr(merchant.UpdatedAt), DeletedAt: convert.FormatTimePtr(merchant.DeletedAt),
	}
}

func mapResponsesGetMerchantsActiveRow(merchants []*repository.MerchantResultDeleteAt) []*pb.MerchantResponseDeleteAt {
	var mapped []*pb.MerchantResponseDeleteAt
	for _, m := range merchants {
		mapped = append(mapped, &pb.MerchantResponseDeleteAt{
			Id: int32(m.MerchantID), UserId: int32(m.UserID), Name: m.Name,
			Description: convert.StrVal(m.Description), Address: convert.StrVal(m.Address),
			ContactEmail: convert.StrVal(m.ContactEmail), ContactPhone: convert.StrVal(m.ContactPhone),
			Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, DeletedAt: m.DeletedAt,
		})
	}
	return mapped
}

func mapResponsesGetMerchantsTrashedRow(merchants []*repository.MerchantResultDeleteAt) []*pb.MerchantResponseDeleteAt {
	return mapResponsesGetMerchantsActiveRow(merchants)
}
