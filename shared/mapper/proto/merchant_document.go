package protomapper

import (
	commonpb "github.com/MamangRust/microservice-point-of-sale-pb/common"
	merchantdocumentpb "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type merchantDocumentProtoMapper struct{}

func NewMerchantDocumentProtoMapper() *merchantDocumentProtoMapper {
	return &merchantDocumentProtoMapper{}
}

func (m *merchantDocumentProtoMapper) ToProtoResponseMerchantDocument(status string, message string, doc *response.MerchantDocumentResponse) *merchantdocumentpb.ApiResponseMerchantDocument {
	return &merchantdocumentpb.ApiResponseMerchantDocument{
		Status:  status,
		Message: message,
		Data:    m.mapMerchantDocument(doc),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponsesMerchantDocument(status string, message string, docs []*response.MerchantDocumentResponse) *merchantdocumentpb.ApiResponsesMerchantDocument {
	return &merchantdocumentpb.ApiResponsesMerchantDocument{
		Status:  status,
		Message: message,
		Data:    m.mapMerchantDocuments(docs),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponsePaginationMerchantDocument(pagination *commonpb.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponse) *merchantdocumentpb.ApiResponsePaginationMerchantDocument {
	return &merchantdocumentpb.ApiResponsePaginationMerchantDocument{
		Status:     status,
		Message:    message,
		Data:       m.mapMerchantDocuments(docs),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponsePaginationMerchantDocumentDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponseDeleteAt) *merchantdocumentpb.ApiResponsePaginationMerchantDocumentAt {
	return &merchantdocumentpb.ApiResponsePaginationMerchantDocumentAt{
		Status:     status,
		Message:    message,
		Data:       m.mapMerchantDocumentsDeleteAt(docs),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponseMerchantDocumentDelete(status string, message string) *merchantdocumentpb.ApiResponseMerchantDocumentDelete {
	return &merchantdocumentpb.ApiResponseMerchantDocumentDelete{
		Status:  status,
		Message: message,
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponseMerchantDocumentAll(status string, message string) *merchantdocumentpb.ApiResponseMerchantDocumentAll {
	return &merchantdocumentpb.ApiResponseMerchantDocumentAll{
		Status:  status,
		Message: message,
	}
}

func (m *merchantDocumentProtoMapper) mapMerchantDocument(doc *response.MerchantDocumentResponse) *merchantdocumentpb.MerchantDocument {
	return &merchantdocumentpb.MerchantDocument{
		DocumentId:   int32(doc.ID),
		MerchantId:   int32(doc.MerchantID),
		DocumentType: doc.DocumentType,
		DocumentUrl:  doc.DocumentURL,
		Status:       doc.Status,
		Note:         doc.Note,
		UploadedAt:   doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}

func (m *merchantDocumentProtoMapper) mapMerchantDocuments(docs []*response.MerchantDocumentResponse) []*merchantdocumentpb.MerchantDocument {
	var res []*merchantdocumentpb.MerchantDocument
	for _, doc := range docs {
		res = append(res, m.mapMerchantDocument(doc))
	}
	return res
}

func (m *merchantDocumentProtoMapper) mapMerchantDocumentDeleteAt(doc *response.MerchantDocumentResponseDeleteAt) *merchantdocumentpb.MerchantDocumentDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if doc.DeletedAt != nil {
		deletedAt = wrapperspb.String(*doc.DeletedAt)
	}

	return &merchantdocumentpb.MerchantDocumentDeleteAt{
		DocumentId:   int32(doc.ID),
		MerchantId:   int32(doc.MerchantID),
		DocumentType: doc.DocumentType,
		DocumentUrl:  doc.DocumentURL,
		Status:       doc.Status,
		Note:         doc.Note,
		UploadedAt:   doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

func (m *merchantDocumentProtoMapper) mapMerchantDocumentsDeleteAt(docs []*response.MerchantDocumentResponseDeleteAt) []*merchantdocumentpb.MerchantDocumentDeleteAt {
	var res []*merchantdocumentpb.MerchantDocumentDeleteAt
	for _, doc := range docs {
		res = append(res, m.mapMerchantDocumentDeleteAt(doc))
	}
	return res
}
