package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type userRoleCommandRepository struct {
	db *mongo.Database
}

func NewUserRoleCommandRepository(db *mongo.Database) UserRoleCommandRepository {
	return &userRoleCommandRepository{db: db}
}

func (r *userRoleCommandRepository) AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	urID, err := database.GetNextID(ctx, r.db, "user_roles")
	if err != nil {
		return nil, err
	}

	ur := &models.UserRole{
		UserRoleID: urID,
		UserID:     int32(req.UserId),
		RoleID:     int32(req.RoleId),
	}

	if _, err := r.db.Collection("user_roles").InsertOne(ctx, ur); err != nil {
		return nil, err
	}
	return ur, nil
}

func (r *userRoleCommandRepository) RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error {
	_, err := r.db.Collection("user_roles").DeleteOne(ctx, bson.M{
		"user_id": int32(req.UserId),
		"role_id": int32(req.RoleId),
	})
	return err
}
