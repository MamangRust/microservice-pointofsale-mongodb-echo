package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type merchantDocumentCommandRepository struct {
	db *mongo.Database
}

func NewMerchantDocumentCommandRepository(db *mongo.Database) MerchantDocumentCommandRepository {
	return &merchantDocumentCommandRepository{db: db}
}

func (r *merchantDocumentCommandRepository) CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*models.MerchantDocument, error) {
	docID, err := database.GetNextID(ctx, r.db, "merchant_documents")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now().UTC()
	emptyNote := ""
	doc := &models.MerchantDocument{
		DocumentID:   docID,
		MerchantID:   int32(request.MerchantID),
		DocumentType: request.DocumentType,
		DocumentUrl:  request.DocumentUrl,
		Status:       "pending",
		Note:         &emptyNote,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if _, err := r.db.Collection("merchant_documents").InsertOne(ctx, doc); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*models.MerchantDocument, error) {
	if request.DocumentID == nil {
		return nil, sharedErrors.ErrBadRequest.WithMessage("document id is required")
	}

	var doc models.MerchantDocument
	coll := r.db.Collection("merchant_documents")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.DocumentID)}).Decode(&doc)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	doc.DocumentType = request.DocumentType
	doc.DocumentUrl = request.DocumentUrl
	doc.Status = request.Status
	doc.Note = &request.Note
	doc.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": doc.DocumentID}, doc); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*models.MerchantDocument, error) {
	if request.DocumentID == nil {
		return nil, sharedErrors.ErrBadRequest.WithMessage("document id is required")
	}

	var doc models.MerchantDocument
	coll := r.db.Collection("merchant_documents")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.DocumentID)}).Decode(&doc)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	doc.Status = request.Status
	doc.Note = &request.Note
	doc.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": doc.DocumentID}, doc); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) TrashedMerchantDocument(ctx context.Context, documentID int) (*models.MerchantDocument, error) {
	var doc models.MerchantDocument
	coll := r.db.Collection("merchant_documents")
	err := coll.FindOne(ctx, bson.M{"_id": int32(documentID)}).Decode(&doc)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": doc.DocumentID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) RestoreMerchantDocument(ctx context.Context, documentID int) (*models.MerchantDocument, error) {
	var doc models.MerchantDocument
	coll := r.db.Collection("merchant_documents")
	err := coll.FindOne(ctx, bson.M{"_id": int32(documentID)}).Decode(&doc)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": doc.DocumentID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) DeleteMerchantDocumentPermanent(ctx context.Context, documentID int) (bool, error) {
	result, err := r.db.Collection("merchant_documents").DeleteOne(ctx, bson.M{"_id": int32(documentID)})
	if err != nil {
		return false, sharedErrors.ErrInternal.WithInternal(err)
	}
	return result.DeletedCount > 0, nil
}

func (r *merchantDocumentCommandRepository) RestoreAllMerchantDocument(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("merchant_documents").UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrInternal.WithInternal(err)
	}
	return true, nil
}

func (r *merchantDocumentCommandRepository) DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("merchant_documents").DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	if err != nil {
		return false, sharedErrors.ErrInternal.WithInternal(err)
	}
	return true, nil
}
