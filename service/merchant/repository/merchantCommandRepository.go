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

type merchantCommandRepository struct {
	db *mongo.Database
}

func NewMerchantCommandRepository(db *mongo.Database) MerchantCommandRepository {
	return &merchantCommandRepository{db: db}
}

func timePtr(t time.Time) *time.Time { return &t }

func (r *merchantCommandRepository) CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*models.Merchant, error) {
	merchantID, err := database.GetNextID(ctx, r.db, "merchants")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now().UTC()
	merchant := &models.Merchant{
		MerchantID:   merchantID,
		UserID:       int32(request.UserID),
		Name:         request.Name,
		Description:  &request.Description,
		Address:      &request.Address,
		ContactEmail: &request.ContactEmail,
		ContactPhone: &request.ContactPhone,
		Status:       "inactive",
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if _, err := r.db.Collection("merchants").InsertOne(ctx, merchant); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*models.Merchant, error) {
	var merchant models.Merchant
	coll := r.db.Collection("merchants")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.MerchantID)}).Decode(&merchant)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	merchant.Name = request.Name
	merchant.Description = &request.Description
	merchant.Address = &request.Address
	merchant.ContactEmail = &request.ContactEmail
	merchant.ContactPhone = &request.ContactPhone
	merchant.Status = request.Status
	merchant.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": merchant.MerchantID}, merchant); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*models.Merchant, error) {
	var merchant models.Merchant
	coll := r.db.Collection("merchants")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.MerchantID)}).Decode(&merchant)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	merchant.Status = request.Status
	merchant.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": merchant.MerchantID}, merchant); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) TrashedMerchant(ctx context.Context, merchantID int) (*models.Merchant, error) {
	var merchant models.Merchant
	coll := r.db.Collection("merchants")
	err := coll.FindOne(ctx, bson.M{"_id": int32(merchantID)}).Decode(&merchant)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": merchant.MerchantID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) RestoreMerchant(ctx context.Context, merchantID int) (*models.Merchant, error) {
	var merchant models.Merchant
	coll := r.db.Collection("merchants")
	err := coll.FindOne(ctx, bson.M{"_id": int32(merchantID)}).Decode(&merchant)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": merchant.MerchantID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) DeleteMerchantPermanent(ctx context.Context, merchantID int) (bool, error) {
	result, err := r.db.Collection("merchants").DeleteOne(ctx, bson.M{"_id": int32(merchantID)})
	if err != nil {
		return false, sharedErrors.ErrInternal.WithInternal(err)
	}
	return result.DeletedCount > 0, nil
}

func (r *merchantCommandRepository) RestoreAllMerchant(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("merchants").UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrInternal.WithInternal(err)
	}
	return true, nil
}

func (r *merchantCommandRepository) DeleteAllMerchantPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("merchants").DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	if err != nil {
		return false, sharedErrors.ErrInternal.WithInternal(err)
	}
	return true, nil
}
