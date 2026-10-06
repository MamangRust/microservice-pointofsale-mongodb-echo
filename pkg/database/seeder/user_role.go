package seeder

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type userRoleSeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserRoleSeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *userRoleSeeder {
	return &userRoleSeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *userRoleSeeder) Seed() error {
	usersColl := r.dbs.Identity.Collection("users")
	rolesColl := r.dbs.Identity.Collection("roles")
	userRolesColl := r.dbs.Identity.Collection("user_roles")

	// Fetch active users
	cursor, err := usersColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		r.logger.Error("failed to fetch users", zap.Error(err))
		return err
	}
	defer cursor.Close(r.ctx)

	var users []bson.M
	if err := cursor.All(r.ctx, &users); err != nil {
		return err
	}

	// Fetch roles
	roleCursor, err := rolesColl.Find(r.ctx, bson.M{})
	if err != nil {
		r.logger.Error("failed to fetch roles", zap.Error(err))
		return err
	}
	defer roleCursor.Close(r.ctx)

	var roles []bson.M
	if err := roleCursor.All(r.ctx, &roles); err != nil {
		return err
	}

	if len(users) == 0 || len(roles) == 0 {
		r.logger.Debug("no users or roles available for seeding")
		return nil
	}

	now := time.Now().UTC()

	for _, user := range users {
		// Pick a random role
		roleIdx := int(user["_id"].(int32)) % len(roles)
		role := roles[roleIdx]

		userID := user["_id"].(int32)
		roleID := role["_id"].(int32)

		// Check if already assigned
		var existing bson.M
		err := userRolesColl.FindOne(r.ctx, bson.M{
			"user_id": userID,
			"role_id": roleID,
		}).Decode(&existing)
		if err == nil {
			continue // Already assigned
		}

		userRoleID, err := database.GetNextID(r.ctx, r.dbs.Identity, "user_roles")
		if err != nil {
			r.logger.Error("failed to get next user_role id", zap.Error(err))
			return err
		}

		doc := bson.M{
			"_id":        userRoleID,
			"user_id":    userID,
			"role_id":    roleID,
			"created_at": now,
			"updated_at": now,
		}

		if _, err := userRolesColl.InsertOne(r.ctx, doc); err != nil {
			r.logger.Error("failed to assign role to user", zap.Error(err))
			return err
		}
	}

	r.logger.Info("user roles assigned successfully", zap.Int("users", len(users)), zap.Int("roles", len(roles)))
	return nil
}
