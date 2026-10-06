package handler

import (
	"context"
	"math"

	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
	pbutils "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pb "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	merchantdocument_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_document_errors"
)

type merchantDocumentQueryHandleGrpc struct {
	pb.UnimplementedMerchantDocumentServiceServer

	merchantDocumentQuery service.MerchantDocumentQueryService

	merchantDocumentCommandService service.MerchantDocumentCommandService
}

func NewMerchantDocumentHandleGrpc(query service.MerchantDocumentQueryService, command service.MerchantDocumentCommandService) *merchantDocumentQueryHandleGrpc {
	return &merchantDocumentQueryHandleGrpc{
		merchantDocumentQuery:          query,
		merchantDocumentCommandService: command,
	}
}

func (s *merchantDocumentQueryHandleGrpc) FindAll(ctx context.Context, req *pb.FindAllMerchantDocumentsRequest) (*pb.ApiResponsePaginationMerchantDocument, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllMerchantDocuments{Page: page, PageSize: pageSize, Search: search}
	documents, totalRecords, err := s.merchantDocumentQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationMerchantDocument{
		Status: "success", Message: "Successfully fetched merchant documents",
		Data: mapResponsesGetMerchantDocumentsRow(documents), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *merchantDocumentQueryHandleGrpc) FindById(ctx context.Context, req *pb.FindMerchantDocumentByIdRequest) (*pb.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}
	document, err := s.merchantDocumentQuery.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseMerchantDocument{Status: "success", Message: "Successfully fetched merchant document", Data: mapMerchantDocument(document)}, nil
}

func (s *merchantDocumentQueryHandleGrpc) FindAllActive(ctx context.Context, req *pb.FindAllMerchantDocumentsRequest) (*pb.ApiResponsePaginationMerchantDocument, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllMerchantDocuments{Page: page, PageSize: pageSize, Search: search}
	documents, totalRecords, err := s.merchantDocumentQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationMerchantDocument{
		Status: "success", Message: "Successfully fetched active merchant documents",
		Data: mapResponsesGetActiveMerchantDocumentsRow(documents), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func (s *merchantDocumentQueryHandleGrpc) FindAllTrashed(ctx context.Context, req *pb.FindAllMerchantDocumentsRequest) (*pb.ApiResponsePaginationMerchantDocumentAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllMerchantDocuments{Page: page, PageSize: pageSize, Search: search}
	documents, totalRecords, err := s.merchantDocumentQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	return &pb.ApiResponsePaginationMerchantDocumentAt{
		Status: "success", Message: "Successfully fetched trashed merchant documents",
		Data: mapResponsesGetTrashedMerchantDocumentsRow(documents), Pagination: &pbutils.PaginationMeta{CurrentPage: int32(page), PageSize: int32(pageSize), TotalPages: int32(totalPages), TotalRecords: int32(*totalRecords)},
	}, nil
}

func mapMerchantDocument(doc *models.MerchantDocument) *pb.MerchantDocument {
	if doc == nil {
		return nil
	}
	return &pb.MerchantDocument{
		DocumentId: int32(doc.DocumentID), MerchantId: int32(doc.MerchantID),
		DocumentType: doc.DocumentType, DocumentUrl: doc.DocumentUrl, Status: doc.Status,
		Note: convert.StrVal(doc.Note), UploadedAt: convert.FormatTimePtr(doc.UploadedAt), UpdatedAt: convert.FormatTimePtr(doc.UpdatedAt),
	}
}

func mapResponsesGetMerchantDocumentsRow(docs []*repository.MerchantDocumentResult) []*pb.MerchantDocument {
	var res []*pb.MerchantDocument
	for _, doc := range docs {
		res = append(res, &pb.MerchantDocument{
			DocumentId: int32(doc.DocumentID), MerchantId: int32(doc.MerchantID),
			DocumentType: doc.DocumentType, DocumentUrl: doc.DocumentUrl, Status: doc.Status,
			Note: convert.StrVal(doc.Note), UploadedAt: doc.UploadedAt, UpdatedAt: doc.UpdatedAt,
		})
	}
	return res
}

func mapResponsesGetActiveMerchantDocumentsRow(docs []*repository.MerchantDocumentResultDeleteAt) []*pb.MerchantDocument {
	var res []*pb.MerchantDocument
	for _, doc := range docs {
		res = append(res, &pb.MerchantDocument{
			DocumentId: int32(doc.DocumentID), MerchantId: int32(doc.MerchantID),
			DocumentType: doc.DocumentType, DocumentUrl: doc.DocumentUrl, Status: doc.Status,
			Note: convert.StrVal(doc.Note), UploadedAt: doc.UploadedAt, UpdatedAt: doc.UpdatedAt,
		})
	}
	return res
}

func mapResponsesGetTrashedMerchantDocumentsRow(docs []*repository.MerchantDocumentResultDeleteAt) []*pb.MerchantDocumentDeleteAt {
	var res []*pb.MerchantDocumentDeleteAt
	for _, doc := range docs {
		res = append(res, &pb.MerchantDocumentDeleteAt{
			DocumentId: int32(doc.DocumentID), MerchantId: int32(doc.MerchantID),
			DocumentType: doc.DocumentType, DocumentUrl: doc.DocumentUrl, Status: doc.Status,
			Note: convert.StrVal(doc.Note), UploadedAt: doc.UploadedAt, UpdatedAt: doc.UpdatedAt,
			DeletedAt: convert.StrValToWrappers(&doc.DeletedAt),
		})
	}
	return res
}
