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

type merchantQueryRepository struct {
	db *mongo.Database
}

func NewMerchantQueryRepository(db *mongo.Database) MerchantQueryRepository {
	return &merchantQueryRepository{db: db}
}

func (r *merchantQueryRepository) FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	count, _ := r.db.Collection("merchants").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "merchant_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("merchants").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var merchants []models.Merchant
	if err := cursor.All(ctx, &merchants); err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var results []*MerchantResult
	for _, m := range merchants {
		mr := &MerchantResult{
			MerchantID: m.MerchantID, UserID: m.UserID, Name: m.Name, Description: m.Description,
			Address: m.Address, ContactEmail: m.ContactEmail, ContactPhone: m.ContactPhone, Status: m.Status, TotalCount: count,
		}
		if m.CreatedAt != nil {
			s := m.CreatedAt.Format("2006-01-02 15:04:05")
			mr.CreatedAt = s
		}
		if m.UpdatedAt != nil {
			s := m.UpdatedAt.Format("2006-01-02 15:04:05")
			mr.UpdatedAt = s
		}
		results = append(results, mr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *merchantQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	count, _ := r.db.Collection("merchants").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "merchant_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("merchants").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var merchants []models.Merchant
	if err := cursor.All(ctx, &merchants); err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var results []*MerchantResultDeleteAt
	for _, m := range merchants {
		mr := &MerchantResultDeleteAt{
			MerchantID: m.MerchantID, UserID: m.UserID, Name: m.Name, Description: m.Description,
			Address: m.Address, ContactEmail: m.ContactEmail, ContactPhone: m.ContactPhone, Status: m.Status, TotalCount: count,
		}
		if m.CreatedAt != nil {
			s := m.CreatedAt.Format("2006-01-02 15:04:05")
			mr.CreatedAt = s
		}
		if m.UpdatedAt != nil {
			s := m.UpdatedAt.Format("2006-01-02 15:04:05")
			mr.UpdatedAt = s
		}
		if m.DeletedAt != nil {
			s := m.DeletedAt.Format("2006-01-02 15:04:05")
			mr.DeletedAt = s
		}
		results = append(results, mr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *merchantQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	count, _ := r.db.Collection("merchants").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "merchant_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("merchants").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var merchants []models.Merchant
	if err := cursor.All(ctx, &merchants); err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var results []*MerchantResultDeleteAt
	for _, m := range merchants {
		mr := &MerchantResultDeleteAt{
			MerchantID: m.MerchantID, UserID: m.UserID, Name: m.Name, Description: m.Description,
			Address: m.Address, ContactEmail: m.ContactEmail, ContactPhone: m.ContactPhone, Status: m.Status, TotalCount: count,
		}
		if m.CreatedAt != nil {
			s := m.CreatedAt.Format("2006-01-02 15:04:05")
			mr.CreatedAt = s
		}
		if m.UpdatedAt != nil {
			s := m.UpdatedAt.Format("2006-01-02 15:04:05")
			mr.UpdatedAt = s
		}
		if m.DeletedAt != nil {
			s := m.DeletedAt.Format("2006-01-02 15:04:05")
			mr.DeletedAt = s
		}
		results = append(results, mr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *merchantQueryRepository) FindById(ctx context.Context, userID int) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.Collection("merchants").FindOne(ctx, bson.M{
		"_id":         int32(userID),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&merchant)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}
