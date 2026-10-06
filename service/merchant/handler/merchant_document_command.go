package handler

import (
	"context"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	merchantdocument_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_document_errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *merchantDocumentQueryHandleGrpc) Create(ctx context.Context, req *pb.CreateMerchantDocumentRequest) (*pb.ApiResponseMerchantDocument, error) {
	request := requests.CreateMerchantDocumentRequest{
		MerchantID: int(req.GetMerchantId()), DocumentType: req.GetDocumentType(), DocumentUrl: req.GetDocumentUrl(),
	}
	if err := request.Validate(); err != nil {
		return nil, merchantdocument_errors.ErrGrpcValidateCreateMerchantDocument
	}
	document, err := s.merchantDocumentCommandService.CreateMerchantDocument(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocument{Status: "success", Message: "Successfully created merchant document", Data: mapMerchantDocument(document)}, nil
}

func (s *merchantDocumentQueryHandleGrpc) Update(ctx context.Context, req *pb.UpdateMerchantDocumentRequest) (*pb.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}
	request := requests.UpdateMerchantDocumentRequest{
		DocumentID: &id, MerchantID: int(req.GetMerchantId()), DocumentType: req.GetDocumentType(),
		DocumentUrl: req.GetDocumentUrl(), Status: req.GetStatus(), Note: req.GetNote(),
	}
	if err := request.Validate(); err != nil {
		return nil, merchantdocument_errors.ErrGrpcFailedUpdateMerchantDocument
	}
	document, err := s.merchantDocumentCommandService.UpdateMerchantDocument(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocument{Status: "success", Message: "Successfully updated merchant document", Data: mapMerchantDocument(document)}, nil
}

func (s *merchantDocumentQueryHandleGrpc) UpdateStatus(ctx context.Context, req *pb.UpdateMerchantDocumentStatusRequest) (*pb.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}
	request := requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: &id, MerchantID: int(req.GetMerchantId()), Status: req.GetStatus(), Note: req.GetNote(),
	}
	if err := request.Validate(); err != nil {
		return nil, merchantdocument_errors.ErrGrpcFailedUpdateMerchantDocument
	}
	document, err := s.merchantDocumentCommandService.UpdateMerchantDocumentStatus(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocument{Status: "success", Message: "Successfully updated merchant document status", Data: mapMerchantDocument(document)}, nil
}

func (s *merchantDocumentQueryHandleGrpc) Trashed(ctx context.Context, req *pb.TrashedMerchantDocumentRequest) (*pb.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}
	document, err := s.merchantDocumentCommandService.TrashedMerchantDocument(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocument{Status: "success", Message: "Successfully trashed merchant document", Data: mapMerchantDocument(document)}, nil
}

func (s *merchantDocumentQueryHandleGrpc) Restore(ctx context.Context, req *pb.RestoreMerchantDocumentRequest) (*pb.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}
	document, err := s.merchantDocumentCommandService.RestoreMerchantDocument(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocument{Status: "success", Message: "Successfully restored merchant document", Data: mapMerchantDocument(document)}, nil
}

func (s *merchantDocumentQueryHandleGrpc) DeletePermanent(ctx context.Context, req *pb.DeleteMerchantDocumentPermanentRequest) (*pb.ApiResponseMerchantDocumentDelete, error) {
	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}
	_, err := s.merchantDocumentCommandService.DeleteMerchantDocumentPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocumentDelete{Status: "success", Message: "Successfully permanently deleted merchant document"}, nil
}

func (s *merchantDocumentQueryHandleGrpc) RestoreAll(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseMerchantDocumentAll, error) {
	_, err := s.merchantDocumentCommandService.RestoreAllMerchantDocument(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocumentAll{Status: "success", Message: "Successfully restored all merchant documents"}, nil
}

func (s *merchantDocumentQueryHandleGrpc) DeleteAllPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseMerchantDocumentAll, error) {
	_, err := s.merchantDocumentCommandService.DeleteAllMerchantDocumentPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocumentAll{Status: "success", Message: "Successfully permanently deleted all merchant documents"}, nil
}

func mapMerchantDocumentDeleteAt(doc *models.MerchantDocument) *pb.MerchantDocumentDeleteAt {
	if doc == nil {
		return nil
	}
	return &pb.MerchantDocumentDeleteAt{
		DocumentId: int32(doc.DocumentID), MerchantId: int32(doc.MerchantID),
		DocumentType: doc.DocumentType, DocumentUrl: doc.DocumentUrl, Status: doc.Status,
		Note: convert.StrVal(doc.Note), UploadedAt: convert.FormatTimePtr(doc.UploadedAt),
		UpdatedAt: convert.FormatTimePtr(doc.UpdatedAt), DeletedAt: convert.TimeToWrappers(doc.DeletedAt),
	}
}

func mapMerchantDocumentResponseDeleteAt(doc *pb.MerchantDocument) *pb.MerchantDocumentDeleteAt {
	if doc == nil {
		return nil
	}
	return &pb.MerchantDocumentDeleteAt{
		DocumentId: doc.DocumentId, MerchantId: doc.MerchantId,
		DocumentType: doc.DocumentType, DocumentUrl: doc.DocumentUrl, Status: doc.Status,
		Note: doc.Note, UploadedAt: doc.UploadedAt, UpdatedAt: doc.UpdatedAt,
		DeletedAt: nil,
	}
}

// wrapper to convert wrapperpb
func strValWrapper(s *string) *wrapperspb.StringValue {
	return convert.StrValToWrappers(s)
}
