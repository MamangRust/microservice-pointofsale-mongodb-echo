package main

import (
	"context"
	"log"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/seeder"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	l, err := logger.NewLogger("seeder", nil)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}

	ctx := context.Background()

	// Satu database = satu MongoDB: buka satu client per context lewat
	// MONGO_<PREFIX>_URI, lalu tutup semuanya saat seeding selesai.
	clients := make([]*mongo.Client, 0, 5)
	openDB := func(name string) *mongo.Database {
		client, err := database.NewMongoClientForContext(l, name)
		if err != nil {
			log.Fatalf("Error connecting to MongoDB for %s: %v", name, err)
		}
		clients = append(clients, client)
		return database.NewMongoDatabase(client, name)
	}

	defer func() {
		for _, client := range clients {
			if err := client.Disconnect(ctx); err != nil {
				log.Printf("Error disconnecting MongoDB: %v", err)
			}
		}
	}()

	dbs := seeder.Databases{
		Identity: openDB(database.IdentityDB),
		Catalog:  openDB(database.CatalogDB),
		Sales:    openDB(database.SalesDB),
		Merchant: openDB(database.MerchantDB),
		Email:    openDB(database.EmailDB),
	}

	s := seeder.NewSeeder(seeder.Deps{
		DBs:    dbs,
		Ctx:    ctx,
		Logger: l,
		Hash:   hash.NewHashingPassword(),
	})

	if err := s.Run(); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	l.Info("Seeding completed successfully.")
}
