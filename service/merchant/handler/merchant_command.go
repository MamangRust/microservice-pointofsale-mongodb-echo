package handler

import (
	"context"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	merchant_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *merchantQueryHandleGrpc) Create(ctx context.Context, request *pb.CreateMerchantRequest) (*pb.ApiResponseMerchant, error) {
	req := &requests.CreateMerchantRequest{
		UserID: int(request.GetUserId()), Name: request.GetName(), Description: request.GetDescription(),
		Address: request.GetAddress(), ContactEmail: request.GetContactEmail(), ContactPhone: request.GetContactPhone(), Status: request.GetStatus(),
	}
	if err := req.Validate(); err != nil {
		return nil, merchant_errors.ErrGrpcValidateCreateMerchant
	}
	merchant, err := s.merchantCommandService.CreateMerchant(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchant{Status: "success", Message: "Successfully created merchant", Data: mapResponseMerchant(merchant)}, nil
}

func (s *merchantQueryHandleGrpc) Update(ctx context.Context, request *pb.UpdateMerchantRequest) (*pb.ApiResponseMerchant, error) {
	id := int(request.GetMerchantId())
	if id <= 0 {
		return nil, merchant_errors.ErrGrpcInvalidID
	}
	req := &requests.UpdateMerchantRequest{
		MerchantID: &id, UserID: int(request.GetUserId()), Name: request.GetName(), Description: request.GetDescription(),
		Address: request.GetAddress(), ContactEmail: request.GetContactEmail(), ContactPhone: request.GetContactPhone(), Status: request.GetStatus(),
	}
	if err := req.Validate(); err != nil {
		return nil, merchant_errors.ErrGrpcValidateUpdateMerchant
	}
	merchant, err := s.merchantCommandService.UpdateMerchant(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchant{Status: "success", Message: "Successfully updated merchant", Data: mapResponseMerchant(merchant)}, nil
}

func (s *merchantQueryHandleGrpc) UpdateMerchantStatus(ctx context.Context, req *pb.UpdateMerchantStatusRequest) (*pb.ApiResponseMerchant, error) {
	id := int(req.GetMerchantId())
	if id <= 0 {
		return nil, merchant_errors.ErrGrpcInvalidID
	}
	request := requests.UpdateMerchantStatusRequest{MerchantID: &id, Status: req.GetStatus()}
	if err := request.Validate(); err != nil {
		return nil, merchant_errors.ErrGrpcValidateUpdateMerchantStatus
	}
	merchant, err := s.merchantCommandService.UpdateMerchantStatus(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchant{Status: "success", Message: "Successfully updated merchant status", Data: mapResponseMerchant(merchant)}, nil
}

func (s *merchantQueryHandleGrpc) TrashedMerchant(ctx context.Context, request *pb.FindByIdMerchantRequest) (*pb.ApiResponseMerchantDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, merchant_errors.ErrGrpcInvalidID
	}
	merchant, err := s.merchantCommandService.TrashedMerchant(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDeleteAt{Status: "success", Message: "Successfully trashed merchant", Data: mapResponseMerchantDeleteAt(merchant)}, nil
}

func (s *merchantQueryHandleGrpc) RestoreMerchant(ctx context.Context, request *pb.FindByIdMerchantRequest) (*pb.ApiResponseMerchant, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, merchant_errors.ErrGrpcInvalidID
	}
	merchant, err := s.merchantCommandService.RestoreMerchant(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchant{Status: "success", Message: "Successfully restored merchant", Data: mapResponseMerchant(merchant)}, nil
}

func (s *merchantQueryHandleGrpc) DeleteMerchantPermanent(ctx context.Context, request *pb.FindByIdMerchantRequest) (*pb.ApiResponseMerchantDelete, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, merchant_errors.ErrGrpcInvalidID
	}
	_, err := s.merchantCommandService.DeleteMerchantPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDelete{Status: "success", Message: "Successfully deleted merchant permanently"}, nil
}

func (s *merchantQueryHandleGrpc) RestoreAllMerchant(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseMerchantAll, error) {
	_, err := s.merchantCommandService.RestoreAllMerchant(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantAll{Status: "success", Message: "Successfully restore all merchant"}, nil
}

func (s *merchantQueryHandleGrpc) DeleteAllMerchantPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseMerchantAll, error) {
	_, err := s.merchantCommandService.DeleteAllMerchantPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantAll{Status: "success", Message: "Successfully delete merchant permanen"}, nil
}
