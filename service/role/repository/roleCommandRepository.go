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

type roleCommandRepository struct {
	db *mongo.Database
}

func NewRoleCommandRepository(db *mongo.Database) RoleCommandRepository {
	return &roleCommandRepository{db: db}
}

func timePtr(t time.Time) *time.Time { return &t }

func (r *roleCommandRepository) CreateRole(ctx context.Context, req *requests.CreateRoleRequest) (*models.Role, error) {
	roleID, err := database.GetNextID(ctx, r.db, "roles")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	role := &models.Role{
		RoleID:    roleID,
		RoleName:  req.Name,
		CreatedAt: timePtr(now),
		UpdatedAt: timePtr(now),
	}

	coll := r.db.Collection("roles")
	if _, err := coll.InsertOne(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleCommandRepository) UpdateRole(ctx context.Context, req *requests.UpdateRoleRequest) (*models.Role, error) {
	var role models.Role
	coll := r.db.Collection("roles")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*req.ID)}).Decode(&role)
	if err != nil {
		return nil, err
	}

	role.RoleName = req.Name
	role.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": role.RoleID}, role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleCommandRepository) TrashedRole(ctx context.Context, id int) (*models.Role, error) {
	var role models.Role
	coll := r.db.Collection("roles")
	err := coll.FindOne(ctx, bson.M{"_id": int32(id)}).Decode(&role)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": role.RoleID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, err
	}
	role.DeletedAt = &now
	return &role, nil
}

func (r *roleCommandRepository) RestoreRole(ctx context.Context, id int) (*models.Role, error) {
	var role models.Role
	coll := r.db.Collection("roles")
	err := coll.FindOne(ctx, bson.M{"_id": int32(id)}).Decode(&role)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": role.RoleID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
		"$set":   bson.M{"updated_at": now},
	})
	if err != nil {
		return nil, err
	}
	role.DeletedAt = nil
	return &role, nil
}

func (r *roleCommandRepository) DeleteRolePermanent(ctx context.Context, id int) (bool, error) {
	coll := r.db.Collection("roles")
	result, err := coll.DeleteOne(ctx, bson.M{"_id": int32(id)})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}

func (r *roleCommandRepository) RestoreAllRole(ctx context.Context) (bool, error) {
	coll := r.db.Collection("roles")
	_, err := coll.UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	return err == nil, err
}

func (r *roleCommandRepository) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	coll := r.db.Collection("roles")
	_, err := coll.DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	return err == nil, err
}
