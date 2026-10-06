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

func timePtr(t time.Time) *time.Time { return &t }

type refreshTokenRepository struct {
	db *mongo.Database
}

func NewRefreshTokenRepository(db *mongo.Database) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Collection("refresh_tokens").FindOne(ctx, bson.M{
		"token": token,
	}).Decode(&rt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) FindByUserId(ctx context.Context, user_id int) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Collection("refresh_tokens").FindOne(ctx, bson.M{
		"user_id": int32(user_id),
	}).Decode(&rt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*models.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rtID, err := database.GetNextID(ctx, r.db, "refresh_tokens")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now().UTC()
	rt := &models.RefreshToken{
		RefreshTokenID: rtID,
		UserID:         int32(req.UserId),
		Token:          req.Token,
		Expiration:     expirationTime,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	if _, err := r.db.Collection("refresh_tokens").InsertOne(ctx, rt); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return rt, nil
}

func (r *refreshTokenRepository) UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*models.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var rt models.RefreshToken
	coll := r.db.Collection("refresh_tokens")
	err = coll.FindOne(ctx, bson.M{"user_id": int32(req.UserId)}).Decode(&rt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rt.Token = req.Token
	rt.Expiration = expirationTime
	rt.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": rt.RefreshTokenID}, rt); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.Collection("refresh_tokens").DeleteOne(ctx, bson.M{
		"token": token,
	})
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error {
	_, err := r.db.Collection("refresh_tokens").DeleteMany(ctx, bson.M{
		"user_id": int32(user_id),
	})
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}
