package handler

import (
	"context"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	cashier_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *cashierQueryHandleGrpc) CreateCashier(ctx context.Context, request *pb.CreateCashierRequest) (*pb.ApiResponseCashier, error) {
	req := &requests.CreateCashierRequest{Name: request.GetName(), MerchantID: int(request.GetMerchantId()), UserID: int(request.GetUserId())}
	if err := req.Validate(); err != nil {
		return nil, cashier_errors.ErrGrpcValidateCreateCashier
	}
	cashier, err := s.cashierCommandService.CreateCashier(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashier{Status: "success", Message: "Successfully created cashier", Data: mapResponseCashier(cashier)}, nil
}

func (s *cashierQueryHandleGrpc) UpdateCashier(ctx context.Context, request *pb.UpdateCashierRequest) (*pb.ApiResponseCashier, error) {
	id := int(request.GetCashierId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}
	req := &requests.UpdateCashierRequest{CashierID: &id, Name: request.GetName()}
	if err := req.Validate(); err != nil {
		return nil, cashier_errors.ErrGrpcValidateUpdateCashier
	}
	cashier, err := s.cashierCommandService.UpdateCashier(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashier{Status: "success", Message: "Successfully updated cashier", Data: mapResponseCashier(cashier)}, nil
}

func (s *cashierQueryHandleGrpc) TrashedCashier(ctx context.Context, request *pb.FindByIdCashierRequest) (*pb.ApiResponseCashierDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}
	cashier, err := s.cashierCommandService.TrashedCashier(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashierDeleteAt{Status: "success", Message: "Successfully trashed cashier", Data: mapResponseCashierDeleteAt(cashier)}, nil
}

func (s *cashierQueryHandleGrpc) RestoreCashier(ctx context.Context, request *pb.FindByIdCashierRequest) (*pb.ApiResponseCashierDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}
	cashier, err := s.cashierCommandService.RestoreCashier(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashierDeleteAt{Status: "success", Message: "Successfully restored cashier", Data: mapResponseCashierDeleteAt(cashier)}, nil
}

func (s *cashierQueryHandleGrpc) DeleteCashierPermanent(ctx context.Context, request *pb.FindByIdCashierRequest) (*pb.ApiResponseCashierDelete, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}
	_, err := s.cashierCommandService.DeleteCashierPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashierDelete{Status: "success", Message: "Successfully deleted cashier permanently"}, nil
}

func (s *cashierQueryHandleGrpc) RestoreAllCashier(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseCashierAll, error) {
	_, err := s.cashierCommandService.RestoreAllCashier(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashierAll{Status: "success", Message: "Successfully restore all cashier"}, nil
}

func (s *cashierQueryHandleGrpc) DeleteAllCashierPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseCashierAll, error) {
	_, err := s.cashierCommandService.DeleteAllCashierPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseCashierAll{Status: "success", Message: "Successfully delete cashier permanen"}, nil
}
