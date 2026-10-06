package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type userCommandRepository struct {
	db *mongo.Database
}

func NewUserCommandRepository(db *mongo.Database) UserCommandRepository {
	return &userCommandRepository{db: db}
}

func timePtr(t time.Time) *time.Time { return &t }

func (r *userCommandRepository) CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*models.User, error) {
	userID, err := database.GetNextID(ctx, r.db, "users")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &models.User{
		UserID:    userID,
		Firstname: request.FirstName,
		Lastname:  request.LastName,
		Email:     request.Email,
		Password:  request.Password,
		CreatedAt: timePtr(now),
		UpdatedAt: timePtr(now),
	}

	if _, err := r.db.Collection("users").InsertOne(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userCommandRepository) UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*models.User, error) {
	var user models.User
	coll := r.db.Collection("users")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.UserID)}).Decode(&user)
	if err != nil {
		return nil, err
	}

	user.Firstname = request.FirstName
	user.Lastname = request.LastName
	user.Email = request.Email
	user.Password = request.Password
	user.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": user.UserID}, user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userCommandRepository) TrashedUser(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	coll := r.db.Collection("users")
	err := coll.FindOne(ctx, bson.M{"_id": int32(userID)}).Decode(&user)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": user.UserID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, err
	}
	user.DeletedAt = &now
	return &user, nil
}

func (r *userCommandRepository) RestoreUser(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	coll := r.db.Collection("users")
	err := coll.FindOne(ctx, bson.M{"_id": int32(userID)}).Decode(&user)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": user.UserID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
		"$set":   bson.M{"updated_at": now},
	})
	if err != nil {
		return nil, err
	}
	user.DeletedAt = nil
	return &user, nil
}

func (r *userCommandRepository) DeleteUserPermanent(ctx context.Context, userID int) (bool, error) {
	coll := r.db.Collection("users")
	result, err := coll.DeleteOne(ctx, bson.M{"_id": int32(userID)})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}

func (r *userCommandRepository) RestoreAllUser(ctx context.Context) (bool, error) {
	coll := r.db.Collection("users")
	_, err := coll.UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	return err == nil, err
}

func (r *userCommandRepository) DeleteAllUserPermanent(ctx context.Context) (bool, error) {
	coll := r.db.Collection("users")
	_, err := coll.DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	return err == nil, err
}

func (r *userCommandRepository) CreateUserRecord(ctx context.Context, request *requests.CreateUserRecordRequest) (*models.User, error) {
	userID, err := database.GetNextID(ctx, r.db, "users")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	isVerified := request.IsVerified
	user := &models.User{
		UserID:           userID,
		Firstname:        request.FirstName,
		Lastname:         request.LastName,
		Email:            request.Email,
		Password:         request.Password,
		VerificationCode: request.VerificationCode,
		IsVerified:       &isVerified,
		CreatedAt:        timePtr(now),
		UpdatedAt:        timePtr(now),
	}

	if _, err := r.db.Collection("users").InsertOne(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userCommandRepository) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error) {
	var user models.User
	coll := r.db.Collection("users")
	err := coll.FindOne(ctx, bson.M{"_id": int32(userID)}).Decode(&user)
	if err != nil {
		return nil, err
	}

	user.IsVerified = &isVerified
	user.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": user.UserID}, user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userCommandRepository) UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error) {
	var user models.User
	coll := r.db.Collection("users")
	err := coll.FindOne(ctx, bson.M{"_id": int32(userID)}).Decode(&user)
	if err != nil {
		return nil, err
	}

	user.Password = password
	user.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": user.UserID}, user); err != nil {
		return nil, err
	}
	return &user, nil
}
