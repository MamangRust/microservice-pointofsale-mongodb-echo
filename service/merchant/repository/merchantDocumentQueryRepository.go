package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type merchantDocumentQueryRepository struct {
	db *mongo.Database
}

func NewMerchantDocumentQueryRepository(db *mongo.Database) MerchantDocumentQueryRepository {
	return &merchantDocumentQueryRepository{db: db}
}

func (r *merchantDocumentQueryRepository) FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"document_type": bson.M{"$regex": req.Search, "$options": "i"}},
			{"document_url": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	count, _ := r.db.Collection("merchant_documents").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "document_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("merchant_documents").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var docs []models.MerchantDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var results []*MerchantDocumentResult
	for _, d := range docs {
		dr := &MerchantDocumentResult{
			DocumentID: d.DocumentID, MerchantID: d.MerchantID, DocumentType: d.DocumentType,
			DocumentUrl: d.DocumentUrl, Status: d.Status, Note: d.Note, TotalCount: count,
		}
		if d.UploadedAt != nil {
			s := d.UploadedAt.Format("2006-01-02 15:04:05")
			dr.UploadedAt = s
		}
		if d.CreatedAt != nil {
			s := d.CreatedAt.Format("2006-01-02 15:04:05")
			dr.CreatedAt = s
		}
		if d.UpdatedAt != nil {
			s := d.UpdatedAt.Format("2006-01-02 15:04:05")
			dr.UpdatedAt = s
		}
		results = append(results, dr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *merchantDocumentQueryRepository) FindById(ctx context.Context, id int) (*models.MerchantDocument, error) {
	var doc models.MerchantDocument
	err := r.db.Collection("merchant_documents").FindOne(ctx, bson.M{
		"_id":         int32(id),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&doc)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"document_type": bson.M{"$regex": req.Search, "$options": "i"}},
			{"document_url": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	count, _ := r.db.Collection("merchant_documents").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "document_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("merchant_documents").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var docs []models.MerchantDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var results []*MerchantDocumentResultDeleteAt
	for _, d := range docs {
		dr := &MerchantDocumentResultDeleteAt{
			DocumentID: d.DocumentID, MerchantID: d.MerchantID, DocumentType: d.DocumentType,
			DocumentUrl: d.DocumentUrl, Status: d.Status, Note: d.Note, TotalCount: count,
		}
		if d.UploadedAt != nil {
			s := d.UploadedAt.Format("2006-01-02 15:04:05")
			dr.UploadedAt = s
		}
		if d.CreatedAt != nil {
			s := d.CreatedAt.Format("2006-01-02 15:04:05")
			dr.CreatedAt = s
		}
		if d.UpdatedAt != nil {
			s := d.UpdatedAt.Format("2006-01-02 15:04:05")
			dr.UpdatedAt = s
		}
		if d.DeletedAt != nil {
			s := d.DeletedAt.Format("2006-01-02 15:04:05")
			dr.DeletedAt = s
		}
		results = append(results, dr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *merchantDocumentQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"document_type": bson.M{"$regex": req.Search, "$options": "i"}},
			{"document_url": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	count, _ := r.db.Collection("merchant_documents").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "document_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("merchant_documents").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var docs []models.MerchantDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var results []*MerchantDocumentResultDeleteAt
	for _, d := range docs {
		dr := &MerchantDocumentResultDeleteAt{
			DocumentID: d.DocumentID, MerchantID: d.MerchantID, DocumentType: d.DocumentType,
			DocumentUrl: d.DocumentUrl, Status: d.Status, Note: d.Note, TotalCount: count,
		}
		if d.UploadedAt != nil {
			s := d.UploadedAt.Format("2006-01-02 15:04:05")
			dr.UploadedAt = s
		}
		if d.CreatedAt != nil {
			s := d.CreatedAt.Format("2006-01-02 15:04:05")
			dr.CreatedAt = s
		}
		if d.UpdatedAt != nil {
			s := d.UpdatedAt.Format("2006-01-02 15:04:05")
			dr.UpdatedAt = s
		}
		if d.DeletedAt != nil {
			s := d.DeletedAt.Format("2006-01-02 15:04:05")
			dr.DeletedAt = s
		}
		results = append(results, dr)
	}
	total := int(count)
	return results, &total, nil
}
