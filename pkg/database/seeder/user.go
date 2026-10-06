package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type userSeeder struct {
	dbs    Databases
	hash   hash.HashPassword
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserSeeder(dbs Databases, hash hash.HashPassword, ctx context.Context, logger logger.LoggerInterface) *userSeeder {
	return &userSeeder{dbs: dbs, hash: hash, ctx: ctx, logger: logger}
}

func (r *userSeeder) Seed() error {
	coll := r.dbs.Identity.Collection("users")

	for i := 1; i <= 10; i++ {
		email := fmt.Sprintf("user_%s@example.com", uuid.NewString())
		rawPassword := fmt.Sprintf("password%d", i)

		hashedPassword, err := r.hash.HashPassword(rawPassword)
		if err != nil {
			r.logger.Error("failed to hash password", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to hash password for user %d: %w", i, err)
		}

		now := time.Now().UTC()

		userID, err := database.GetNextID(r.ctx, r.dbs.Identity, "users")
		if err != nil {
			r.logger.Error("failed to get next user id", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to get next user id: %w", err)
		}

		doc := bson.M{
			"_id":              userID,
			"firstname":        fmt.Sprintf("User%d", i),
			"lastname":         fmt.Sprintf("Last%d", i),
			"email":            email,
			"password":         hashedPassword,
			"verification_code": "",
			"is_verified":      false,
			"created_at":       now,
			"updated_at":       now,
		}

		if i > 5 {
			doc["deleted_at"] = now
		}

		if _, err := coll.InsertOne(r.ctx, doc); err != nil {
			r.logger.Error("failed to seed user", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to seed user %d: %w", i, err)
		}
	}

	r.logger.Info("User seeding completed successfully", zap.Int("totalUsers", 10))
	return nil
}
