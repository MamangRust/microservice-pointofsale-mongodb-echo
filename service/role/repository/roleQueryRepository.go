package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type roleQueryRepository struct {
	db *mongo.Database
}

func NewRoleQueryRepository(db *mongo.Database) RoleQueryRepository {
	return &roleQueryRepository{db: db}
}

func decodeFacetResult(cursor *mongo.Cursor, ctx context.Context) (totalCount int64, data []bson.M, err error) {
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

func bsonMToTime(v bson.M, key string) *string {
	if val, ok := v[key]; ok {
		if dt, ok := val.(bson.DateTime); ok {
			s := dt.Time().Format(time.RFC3339)
			return &s
		}
	}
	return nil
}

func bsonMToInt32(v bson.M, key string) int32 {
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

func (r *roleQueryRepository) FindAllRoles(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["role_name"] = bson.M{"$regex": req.Search, "$options": "i"}
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

	cursor, err := r.db.Collection("roles").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeFacetResult(cursor, ctx)
	if err != nil {
		return nil, err
	}

	var results []*RoleResult
	for _, d := range data {
		r := &RoleResult{
			RoleName:   fmt.Sprintf("%v", d["role_name"]),
			RoleID:     bsonMToInt32(d, "_id"),
			TotalCount: totalCount,
			CreatedAt:  bsonMToTime(d, "created_at"),
			UpdatedAt:  bsonMToTime(d, "updated_at"),
			DeletedAt:  bsonMToTime(d, "deleted_at"),
		}
		results = append(results, r)
	}

	return results, nil
}

func (r *roleQueryRepository) FindById(ctx context.Context, id int) (*models.Role, error) {
	var role models.Role
	err := r.db.Collection("roles").FindOne(ctx, bson.M{
		"_id":        int32(id),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Collection("roles").FindOne(ctx, bson.M{
		"role_name":  name,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByUserId(ctx context.Context, userID int) ([]*models.Role, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id":    int32(userID),
			"deleted_at": bson.M{"$exists": false},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "roles",
			"localField":   "role_id",
			"foreignField": "_id",
			"as":           "role",
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"deleted_at": bson.M{"$exists": false}}}},
			},
		}}},
		{{Key: "$unwind", Value: "$role"}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$role"}}},
	}

	cursor, err := r.db.Collection("user_roles").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*models.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleQueryRepository) FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["role_name"] = bson.M{"$regex": req.Search, "$options": "i"}
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

	cursor, err := r.db.Collection("roles").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeFacetResult(cursor, ctx)
	if err != nil {
		return nil, err
	}

	var results []*RoleResult
	for _, d := range data {
		rr := &RoleResult{
			RoleName:   fmt.Sprintf("%v", d["role_name"]),
			RoleID:     bsonMToInt32(d, "_id"),
			TotalCount: totalCount,
			CreatedAt:  bsonMToTime(d, "created_at"),
			UpdatedAt:  bsonMToTime(d, "updated_at"),
			DeletedAt:  bsonMToTime(d, "deleted_at"),
		}
		results = append(results, rr)
	}

	return results, nil
}

func (r *roleQueryRepository) FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["role_name"] = bson.M{"$regex": req.Search, "$options": "i"}
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

	cursor, err := r.db.Collection("roles").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeFacetResult(cursor, ctx)
	if err != nil {
		return nil, err
	}

	var results []*RoleResult
	for _, d := range data {
		rr := &RoleResult{
			RoleName:   fmt.Sprintf("%v", d["role_name"]),
			RoleID:     bsonMToInt32(d, "_id"),
			TotalCount: totalCount,
			CreatedAt:  bsonMToTime(d, "created_at"),
			UpdatedAt:  bsonMToTime(d, "updated_at"),
			DeletedAt:  bsonMToTime(d, "deleted_at"),
		}
		results = append(results, rr)
	}

	return results, nil
}
