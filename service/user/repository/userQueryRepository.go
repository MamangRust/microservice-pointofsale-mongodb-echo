package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	user_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type userQueryRepository struct {
	db *mongo.Database
}

func NewUserQueryRepository(db *mongo.Database) UserQueryRepository {
	return &userQueryRepository{db: db}
}

func decodeUserFacetResult(cursor *mongo.Cursor, ctx context.Context) (totalCount int64, data []bson.M, err error) {
	var results []struct {
		Metadata []struct {
			TotalCount int64 `bson:"total_count"`
		} `bson:"metadata"`
		Data []bson.M `bson:"data"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return 0, nil, err
	}
	if len(results) == 0 {
		return 0, nil, nil
	}

	if len(results[0].Metadata) > 0 {
		totalCount = results[0].Metadata[0].TotalCount
	}
	return totalCount, results[0].Data, nil
}

func bsonMToTimeStr(v bson.M, key string) *string {
	if val, ok := v[key]; ok {
		if dt, ok := val.(bson.DateTime); ok {
			s := dt.Time().Format(time.RFC3339)
			return &s
		}
	}
	return nil
}

func bsonMToInt32Val(v bson.M, key string) int32 {
	if val, ok := v[key]; ok {
		switch n := val.(type) {
		case int32:
			return n
		case int64:
			return int32(n)
		case float64:
			return int32(n)
		}
	}
	return 0
}

func (r *userQueryRepository) FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"firstname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"lastname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"email": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": mongo.Pipeline{
				{{Key: "$count", Value: "total_count"}},
			},
			"data": mongo.Pipeline{
				{{Key: "$skip", Value: offset}},
				{{Key: "$limit", Value: limit}},
			},
		}}},
	}

	cursor, err := r.db.Collection("users").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeUserFacetResult(cursor, ctx)
	if err != nil {
		return nil, err
	}

	var results []*UserResult
	for _, d := range data {
		u := &UserResult{
			UserID:     bsonMToInt32Val(d, "_id"),
			Firstname:  fmt.Sprintf("%v", d["firstname"]),
			Lastname:   fmt.Sprintf("%v", d["lastname"]),
			Email:      fmt.Sprintf("%v", d["email"]),
			TotalCount: totalCount,
			CreatedAt:  bsonMToTimeStr(d, "created_at"),
			UpdatedAt:  bsonMToTimeStr(d, "updated_at"),
			DeletedAt:  bsonMToTimeStr(d, "deleted_at"),
		}
		results = append(results, u)
	}

	return results, nil
}

func (r *userQueryRepository) FindById(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{
		"_id":        int32(userID),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userQueryRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{
		"email":      email,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userQueryRepository) FindByEmailWithPassword(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{
		"email":      email,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userQueryRepository) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{
		"email":       email,
		"is_verified": true,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userQueryRepository) FindByVerificationCode(ctx context.Context, verificationCode string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{
		"verification_code": verificationCode,
		"deleted_at":        bson.M{"$exists": false},
	}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"firstname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"lastname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"email": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": mongo.Pipeline{
				{{Key: "$count", Value: "total_count"}},
			},
			"data": mongo.Pipeline{
				{{Key: "$skip", Value: offset}},
				{{Key: "$limit", Value: limit}},
			},
		}}},
	}

	cursor, err := r.db.Collection("users").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeUserFacetResult(cursor, ctx)
	if err != nil {
		return nil, err
	}

	var results []*UserResult
	for _, d := range data {
		u := &UserResult{
			UserID:     bsonMToInt32Val(d, "_id"),
			Firstname:  fmt.Sprintf("%v", d["firstname"]),
			Lastname:   fmt.Sprintf("%v", d["lastname"]),
			Email:      fmt.Sprintf("%v", d["email"]),
			TotalCount: totalCount,
			CreatedAt:  bsonMToTimeStr(d, "created_at"),
			UpdatedAt:  bsonMToTimeStr(d, "updated_at"),
			DeletedAt:  bsonMToTimeStr(d, "deleted_at"),
		}
		results = append(results, u)
	}

	return results, nil
}

func (r *userQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["$or"] = []bson.M{
			{"firstname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"lastname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"email": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": mongo.Pipeline{
				{{Key: "$count", Value: "total_count"}},
			},
			"data": mongo.Pipeline{
				{{Key: "$skip", Value: offset}},
				{{Key: "$limit", Value: limit}},
			},
		}}},
	}

	cursor, err := r.db.Collection("users").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeUserFacetResult(cursor, ctx)
	if err != nil {
		return nil, err
	}

	var results []*UserResult
	for _, d := range data {
		u := &UserResult{
			UserID:     bsonMToInt32Val(d, "_id"),
			Firstname:  fmt.Sprintf("%v", d["firstname"]),
			Lastname:   fmt.Sprintf("%v", d["lastname"]),
			Email:      fmt.Sprintf("%v", d["email"]),
			TotalCount: totalCount,
			CreatedAt:  bsonMToTimeStr(d, "created_at"),
			UpdatedAt:  bsonMToTimeStr(d, "updated_at"),
			DeletedAt:  bsonMToTimeStr(d, "deleted_at"),
		}
		results = append(results, u)
	}

	return results, nil
}
