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

type resetTokenRepository struct {
	db *mongo.Database
}

func NewResetTokenRepository(db *mongo.Database) *resetTokenRepository {
	return &resetTokenRepository{db: db}
}

func (r *resetTokenRepository) FindByToken(ctx context.Context, code string) (*models.ResetToken, error) {
	var rt models.ResetToken
	err := r.db.Collection("reset_tokens").FindOne(ctx, bson.M{
		"token": code,
	}).Decode(&rt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *resetTokenRepository) CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*models.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rtID, err := database.GetNextID(ctx, r.db, "reset_tokens")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rt := &models.ResetToken{
		ResetTokenID: rtID,
		UserID:       int64(req.UserID),
		Token:        req.ResetToken,
		ExpiryDate:   expiryDate,
	}

	if _, err := r.db.Collection("reset_tokens").InsertOne(ctx, rt); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return rt, nil
}

func (r *resetTokenRepository) DeleteResetToken(ctx context.Context, user_id int) error {
	_, err := r.db.Collection("reset_tokens").DeleteMany(ctx, bson.M{
		"user_id": int64(user_id),
	})
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}
