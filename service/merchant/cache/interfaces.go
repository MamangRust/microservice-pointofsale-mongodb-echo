package mencache

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type MerchantQueryCache interface {
	GetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*repository.MerchantResult, *int, bool)
	SetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants, data []*repository.MerchantResult, total *int)

	GetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants) ([]*repository.MerchantResultDeleteAt, *int, bool)
	SetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants, data []*repository.MerchantResultDeleteAt, total *int)

	GetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*repository.MerchantResultDeleteAt, *int, bool)
	SetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants, data []*repository.MerchantResultDeleteAt, total *int)

	GetCachedMerchant(ctx context.Context, id int) (*models.Merchant, bool)
	SetCachedMerchant(ctx context.Context, data *models.Merchant)

	GetCachedMerchantsByUserId(ctx context.Context, id int) ([]*models.Merchant, bool)
	SetCachedMerchantsByUserId(ctx context.Context, userId int, data []*models.Merchant)
}

type MerchantCommandCache interface {
	DeleteCachedMerchant(ctx context.Context, id int)
	DeleteCachedMerchantAllCache(ctx context.Context)
}

type MerchantDocumentQueryCache interface {
	GetCachedMerchantDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*repository.MerchantDocumentResult, *int, bool)
	SetCachedMerchantDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*repository.MerchantDocumentResult, total *int)

	GetCachedMerchantDocumentsActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*repository.MerchantDocumentResultDeleteAt, *int, bool)
	SetCachedMerchantDocumentsActive(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*repository.MerchantDocumentResultDeleteAt, total *int)

	GetCachedMerchantDocumentsTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*repository.MerchantDocumentResultDeleteAt, *int, bool)
	SetCachedMerchantDocumentsTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*repository.MerchantDocumentResultDeleteAt, total *int)

	GetCachedMerchantDocument(ctx context.Context, id int) (*models.MerchantDocument, bool)
	SetCachedMerchantDocument(ctx context.Context, data *models.MerchantDocument)
}

type MerchantDocumentCommandCache interface {
	DeleteCachedMerchantDocuments(ctx context.Context, id int)
	DeleteCachedMerchantDocumentsAllCache(ctx context.Context)
}
