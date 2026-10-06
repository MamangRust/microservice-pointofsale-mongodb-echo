package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// MerchantResult is the result type for paginated merchant queries.
type MerchantResult struct {
	MerchantID   int32
	UserID       int32
	Name         string
	Description  *string
	Address      *string
	ContactEmail *string
	ContactPhone *string
	Status       string
	CreatedAt    string
	UpdatedAt    string
	TotalCount   int64
}

// MerchantResultDeleteAt is the result type for paginated merchant queries with deleted_at.
type MerchantResultDeleteAt struct {
	MerchantID   int32
	UserID       int32
	Name         string
	Description  *string
	Address      *string
	ContactEmail *string
	ContactPhone *string
	Status       string
	CreatedAt    string
	UpdatedAt    string
	DeletedAt    string
	TotalCount   int64
}

// MerchantDocumentResult is the result type for paginated merchant document queries.
type MerchantDocumentResult struct {
	DocumentID   int32
	MerchantID   int32
	DocumentType string
	DocumentUrl  string
	Status       string
	Note         *string
	UploadedAt   string
	CreatedAt    string
	UpdatedAt    string
	TotalCount   int64
}

// MerchantDocumentResultDeleteAt is the result type for paginated merchant document queries with deleted_at.
type MerchantDocumentResultDeleteAt struct {
	DocumentID   int32
	MerchantID   int32
	DocumentType string
	DocumentUrl  string
	Status       string
	Note         *string
	UploadedAt   string
	CreatedAt    string
	UpdatedAt    string
	DeletedAt    string
	TotalCount   int64
}

type MerchantDocumentQueryRepository interface {
	FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResult, *int, error)
	FindById(ctx context.Context, id int) (*models.MerchantDocument, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResultDeleteAt, *int, error)
}

type MerchantDocumentCommandRepository interface {
	CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*models.MerchantDocument, error)
	UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*models.MerchantDocument, error)
	UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*models.MerchantDocument, error)
	TrashedMerchantDocument(ctx context.Context, merchantDocumentID int) (*models.MerchantDocument, error)
	RestoreMerchantDocument(ctx context.Context, merchantDocumentID int) (*models.MerchantDocument, error)
	DeleteMerchantDocumentPermanent(ctx context.Context, merchantDocumentID int) (bool, error)
	RestoreAllMerchantDocument(ctx context.Context) (bool, error)
	DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error)
}

type MerchantQueryRepository interface {
	FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResult, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResultDeleteAt, *int, error)
	FindById(ctx context.Context, userID int) (*models.Merchant, error)
}

type MerchantCommandRepository interface {
	CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*models.Merchant, error)
	UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*models.Merchant, error)
	UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*models.Merchant, error)
	TrashedMerchant(ctx context.Context, merchantID int) (*models.Merchant, error)
	RestoreMerchant(ctx context.Context, merchantID int) (*models.Merchant, error)
	DeleteMerchantPermanent(ctx context.Context, merchantID int) (bool, error)
	RestoreAllMerchant(ctx context.Context) (bool, error)
	DeleteAllMerchantPermanent(ctx context.Context) (bool, error)
}
