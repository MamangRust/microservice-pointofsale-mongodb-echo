package handler

import (
	"context"
	"math"

	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"

	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-cashier/service"
	pb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	cashier_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
)

type cashierQueryHandleGrpc struct {
	pb.UnimplementedCashierServiceServer

	cashierQuery service.CashierQueryService

	cashierCommandService service.CashierCommandService
}

func NewCashierHandleGrpc(query service.CashierQueryService, command service.CashierCommandService) *cashierQueryHandleGrpc {
	return &cashierQueryHandleGrpc{
		cashierQuery:          query,
		cashierCommandService: command,
	}
}

func (s *cashierQueryHandleGrpc) FindAll(ctx context.Context, request *pb.FindAllCashierRequest) (*pb.ApiResponsePaginationCashier, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllCashiers{Search: search, Page: page, PageSize: pageSize}
	cashier, totalRecords, err := s.cashierQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationCashier{
		Status: "success", Message: "Successfully fetched cashier",
		Data: mapResponsesCashier(cashier), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *cashierQueryHandleGrpc) FindById(ctx context.Context, request *pb.FindByIdCashierRequest) (*pb.ApiResponseCashier, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}
	cashier, err := s.cashierQuery.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashier{Status: "success", Message: "Successfully fetched cashier", Data: mapResponseCashier(cashier)}, nil
}

func (s *cashierQueryHandleGrpc) FindByActive(ctx context.Context, request *pb.FindAllCashierRequest) (*pb.ApiResponsePaginationCashierDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllCashiers{Search: search, Page: page, PageSize: pageSize}
	cashier, totalRecords, err := s.cashierQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationCashierDeleteAt{
		Status: "success", Message: "Successfully fetched active cashier",
		Data: mapResponsesCashierActive(cashier), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *cashierQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pb.FindAllCashierRequest) (*pb.ApiResponsePaginationCashierDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllCashiers{Search: search, Page: page, PageSize: pageSize}
	users, totalRecords, err := s.cashierQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationCashierDeleteAt{
		Status: "success", Message: "Successfully fetched trashed cashier",
		Data: mapResponsesCashierTrashed(users), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *cashierQueryHandleGrpc) FindByMerchant(ctx context.Context, request *pb.FindByMerchantCashierRequest) (*pb.ApiResponsePaginationCashier, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	merchantID := int(request.GetMerchantId())
	if merchantID <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidMerchantId
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllCashierMerchant{Search: search, Page: page, PageSize: pageSize, MerchantID: merchantID}
	cashier, totalRecords, err := s.cashierQuery.FindByMerchant(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationCashier{
		Status: "success", Message: "Successfully fetched cashier",
		Data: mapResponsesCashierByMerchant(cashier), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func mapResponseCashier(cashier *models.Cashier) *pb.CashierResponse {
	if cashier == nil {
		return nil
	}
	return &pb.CashierResponse{
		Id: cashier.CashierID, MerchantId: cashier.MerchantID, Name: cashier.Name,
		CreatedAt: convert.FormatTimePtr(cashier.CreatedAt), UpdatedAt: convert.FormatTimePtr(cashier.UpdatedAt),
	}
}

func mapResponsesCashier(cashiers []*repository.CashierResult) []*pb.CashierResponse {
	var res []*pb.CashierResponse
	for _, c := range cashiers {
		res = append(res, &pb.CashierResponse{
			Id: c.CashierID, MerchantId: c.MerchantID, Name: c.Name,
			CreatedAt: convert.StrVal(c.CreatedAt), UpdatedAt: convert.StrVal(c.UpdatedAt),
		})
	}
	return res
}

func mapResponsesCashierByMerchant(cashiers []*repository.CashierResult) []*pb.CashierResponse {
	return mapResponsesCashier(cashiers)
}

func mapResponseCashierDeleteAt(cashier *models.Cashier) *pb.CashierResponseDeleteAt {
	if cashier == nil {
		return nil
	}
	return &pb.CashierResponseDeleteAt{
		Id: cashier.CashierID, MerchantId: cashier.MerchantID, Name: cashier.Name,
		CreatedAt: convert.FormatTimePtr(cashier.CreatedAt), UpdatedAt: convert.FormatTimePtr(cashier.UpdatedAt), DeletedAt: convert.TimeToWrappers(cashier.DeletedAt),
	}
}

func mapResponsesCashierActive(cashiers []*repository.CashierResultDeleteAt) []*pb.CashierResponseDeleteAt {
	var res []*pb.CashierResponseDeleteAt
	for _, c := range cashiers {
		res = append(res, &pb.CashierResponseDeleteAt{
			Id: c.CashierID, MerchantId: c.MerchantID, Name: c.Name,
			CreatedAt: convert.StrVal(c.CreatedAt), UpdatedAt: convert.StrVal(c.UpdatedAt), DeletedAt: convert.StrValToWrappers(c.DeletedAt),
		})
	}
	return res
}

func mapResponsesCashierTrashed(cashiers []*repository.CashierResultDeleteAt) []*pb.CashierResponseDeleteAt {
	return mapResponsesCashierActive(cashiers)
}
