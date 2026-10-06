package main

import (
	"context"
	"fmt"
	"log"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// dbIndexes describes the indexes to create inside one bounded context database.
type dbIndexes struct {
	name    string
	indexes map[string][]mongo.IndexModel
}

func main() {
	if err := dotenv.Viper(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	l, err := logger.NewLogger("migrate", nil)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}

	ctx := context.Background()

	// Satu database = satu MongoDB, jadi tiap context dibuka dengan client
	// sendiri (MONGO_<PREFIX>_URI) dan ditutup setelah indeksnya dibuat.
	for _, target := range indexPlan() {
		client, err := database.NewMongoClientForContext(l, target.name)
		if err != nil {
			log.Fatalf("Error connecting to MongoDB for %s: %v", target.name, err)
		}

		db := database.NewMongoDatabase(client, target.name)
		if err := ensureIndexes(ctx, db, target.indexes, l); err != nil {
			_ = client.Disconnect(ctx)
			log.Fatalf("Migration failed for %s: %v", target.name, err)
		}

		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("Error disconnecting MongoDB for %s: %v", target.name, err)
		}
	}

	l.Info("Migration completed successfully.")
}

// indexPlan lists the collections each bounded context owns. outbox_events is
// written by every service that publishes events, so its index is created in
// each of those context databases.
func indexPlan() []dbIndexes {
	return []dbIndexes{
		{name: database.IdentityDB, indexes: mergeIndexes(identityIndexes(), outboxIndexes())},
		{name: database.CatalogDB, indexes: catalogIndexes()},
		{name: database.SalesDB, indexes: mergeIndexes(salesIndexes(), outboxIndexes())},
		{name: database.MerchantDB, indexes: mergeIndexes(merchantIndexes(), outboxIndexes())},
		{name: database.EmailDB, indexes: emailIndexes()},
	}
}

func mergeIndexes(groups ...map[string][]mongo.IndexModel) map[string][]mongo.IndexModel {
	merged := make(map[string][]mongo.IndexModel)
	for _, group := range groups {
		for coll, models := range group {
			merged[coll] = models
		}
	}
	return merged
}

func identityIndexes() map[string][]mongo.IndexModel {
	return map[string][]mongo.IndexModel{
		"users": {
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
		"roles": {
			{
				Keys:    bson.D{{Key: "role_name", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
		"user_roles": {
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "role_id", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
		"refresh_tokens": {
			{
				Keys:    bson.D{{Key: "token", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "user_id", Value: 1}},
			},
		},
		"reset_tokens": {
			{
				Keys:    bson.D{{Key: "token", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "user_id", Value: 1}},
			},
		},
	}
}

func catalogIndexes() map[string][]mongo.IndexModel {
	return map[string][]mongo.IndexModel{
		"categories": {
			{
				Keys:    bson.D{{Key: "name", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
		"products": {
			{
				Keys: bson.D{{Key: "merchant_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "category_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
	}
}

func salesIndexes() map[string][]mongo.IndexModel {
	return map[string][]mongo.IndexModel{
		"orders": {
			{
				Keys: bson.D{{Key: "merchant_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "cashier_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
		"order_items": {
			{
				Keys: bson.D{{Key: "order_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "product_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
		"transactions": {
			{
				Keys: bson.D{{Key: "order_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "merchant_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
	}
}

func merchantIndexes() map[string][]mongo.IndexModel {
	return map[string][]mongo.IndexModel{
		"merchants": {
			{
				Keys: bson.D{{Key: "user_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "status", Value: 1}},
			},
		},
		"merchant_documents": {
			{
				Keys: bson.D{{Key: "merchant_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
		"cashiers": {
			{
				Keys: bson.D{{Key: "merchant_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "user_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "deleted_at", Value: 1}},
			},
		},
	}
}

func emailIndexes() map[string][]mongo.IndexModel {
	return map[string][]mongo.IndexModel{
		"consumer_inbox": {
			{
				Keys:    bson.D{{Key: "consumer_name", Value: 1}, {Key: "event_key", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "status", Value: 1}, {Key: "lease_until", Value: 1}},
			},
		},
	}
}

func outboxIndexes() map[string][]mongo.IndexModel {
	return map[string][]mongo.IndexModel{
		"outbox_events": {
			{
				Keys: bson.D{{Key: "status", Value: 1}, {Key: "next_attempt_at", Value: 1}},
			},
		},
	}
}

func ensureIndexes(ctx context.Context, db *mongo.Database, indexes map[string][]mongo.IndexModel, log logger.LoggerInterface) error {
	for collName, models := range indexes {
		coll := db.Collection(collName)
		if _, err := coll.Indexes().CreateMany(ctx, models); err != nil {
			log.Error("Failed to create indexes", zap.String("database", db.Name()), zap.String("collection", collName), zap.Error(err))
			return fmt.Errorf("failed to create indexes for %s.%s: %w", db.Name(), collName, err)
		}
		log.Info("Indexes created", zap.String("database", db.Name()), zap.String("collection", collName))
	}

	return nil
}
