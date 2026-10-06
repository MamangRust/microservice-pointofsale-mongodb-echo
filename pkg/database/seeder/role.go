package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type roleSeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewRoleSeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *roleSeeder {
	return &roleSeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *roleSeeder) Seed() error {
	coll := r.dbs.Identity.Collection("roles")

	randomRoles := []string{
		"ROLE_ADMIN", "Admin Access 1", "Super Admin", "Admin",
		"Store Manager", "Cashier", "Inventory Staff", "Support", "Auditor", "Viewer",
	}

	now := time.Now().UTC()
	for i, roleName := range randomRoles {
		// Check if role already exists
		var existing bson.M
		err := coll.FindOne(r.ctx, bson.M{"role_name": roleName}).Decode(&existing)
		if err == nil {
			r.logger.Debug("role already exists, skipping", zap.String("roleName", roleName))
			continue
		}

		roleID, err := database.GetNextID(r.ctx, r.dbs.Identity, "roles")
		if err != nil {
			r.logger.Error("failed to get next role id", zap.Int("role", i+1), zap.Error(err))
			return fmt.Errorf("failed to get next role id: %w", err)
		}

		doc := bson.M{
			"_id":        roleID,
			"role_name":  roleName,
			"created_at": now,
			"updated_at": now,
		}

		if _, err := coll.InsertOne(r.ctx, doc); err != nil {
			r.logger.Error("failed to seed role", zap.Int("role", i+1), zap.String("roleName", roleName), zap.Error(err))
			return fmt.Errorf("failed to seed role %d (%s): %w", i+1, roleName, err)
		}
	}

	r.logger.Info("role seeded successfully", zap.Int("totalRoles", len(randomRoles)))
	return nil
}
