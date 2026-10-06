package database

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

// NewMongoClient connects to MongoDB using the MONGO_* environment variables
// and returns a *mongo.Client instance.
func NewMongoClient(log logger.LoggerInterface) (*mongo.Client, error) {
	return NewMongoClientWithPrefix(log, "MONGO")
}

// NewMongoClientForContext connects to the MongoDB instance that owns the given
// context database (see ContextPrefix). Setiap database punya server sendiri,
// jadi pemanggil yang butuh lebih dari satu database harus membuka lebih dari
// satu client.
//
// Bila key per-context tidak ada, rantai fallback di NewMongoClientWithPrefix
// jatuh ke MONGO_URI / MONGO_HOST — sehingga deployment instance-tunggal lama
// tetap jalan tanpa perubahan env.
func NewMongoClientForContext(log logger.LoggerInterface, databaseName string) (*mongo.Client, error) {
	return NewMongoClientWithPrefix(log, ContextPrefix(databaseName))
}

// NewMongoClientWithPrefix connects to MongoDB configured via the given
// prefix keys (e.g. MONGO_HOST, MONGO_PORT) with fallback to base MONGO_* keys.
func NewMongoClientWithPrefix(log logger.LoggerInterface, prefix string) (*mongo.Client, error) {
	if prefix == "" {
		prefix = "MONGO"
	}

	host := viper.GetString(fmt.Sprintf("%s_HOST", prefix))
	if host == "" {
		host = viper.GetString("MONGO_HOST")
	}
	if host == "" {
		host = "localhost"
	}

	port := viper.GetString(fmt.Sprintf("%s_PORT", prefix))
	if port == "" {
		port = viper.GetString("MONGO_PORT")
	}
	if port == "" {
		port = "27017"
	}

	username := viper.GetString(fmt.Sprintf("%s_USERNAME", prefix))
	if username == "" {
		username = viper.GetString("MONGO_USERNAME")
	}

	password := viper.GetString(fmt.Sprintf("%s_PASSWORD", prefix))
	if password == "" {
		password = viper.GetString("MONGO_PASSWORD")
	}

	var uri string
	if username != "" && password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s", username, password, host, port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s", host, port)
	}

	// Allow explicit URI override
	if envURI := viper.GetString(fmt.Sprintf("%s_URI", prefix)); envURI != "" {
		uri = envURI
	} else if envURI := viper.GetString("MONGO_URI"); envURI != "" {
		uri = envURI
	}

	maxPoolSize := viper.GetUint64("MONGO_MAX_POOL_SIZE")
	if maxPoolSize == 0 {
		maxPoolSize = 100
	}
	minPoolSize := viper.GetUint64("MONGO_MIN_POOL_SIZE")
	if minPoolSize == 0 {
		minPoolSize = 10
	}

	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(maxPoolSize).
		SetMinPoolSize(minPoolSize).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Error("Failed to ping MongoDB", zap.Error(err))
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Debug("MongoDB connection established",
		zap.String("prefix", prefix),
		zap.String("host", host),
	)

	return client, nil
}

// NewMongoDatabase returns the named database from the given client.
//
// Nama database wajib eksplisit — tidak ada fallback ke MONGO_NAME/env lagi.
// Pemanggil bertanggung jawab memastikan nama tidak kosong (pkg/server
// memvalidasinya lewat Config.DBName; seeder/backfill memakai konstanta di
// names.go yang tidak mungkin kosong).
func NewMongoDatabase(client *mongo.Client, database string) *mongo.Database {
	return client.Database(database)
}

// GetNextID atomically increments a counter and returns the next int32 ID
// for the given collection. It uses a "counters" collection to generate
// sequential IDs compatible with the existing int32 primary key scheme.
func GetNextID(ctx context.Context, db *mongo.Database, collection string) (int32, error) {
	counters := db.Collection("counters")

	filter := bson.M{"_id": collection}
	update := bson.M{"$inc": bson.M{"seq": int32(1)}}
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetUpsert(true)

	var result struct {
		Seq int32 `bson:"seq"`
	}
	err := counters.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to get next ID for %s: %w", collection, err)
	}
	return result.Seq, nil
}
