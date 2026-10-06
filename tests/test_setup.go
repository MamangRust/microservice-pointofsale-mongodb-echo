package tests

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongoopts "go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
)

type TestSuite struct {
	MongoContainer *mongodb.MongoDBContainer
	RedisContainer *redis.RedisContainer
	MongoClient    *mongo.Client

	// Databases holds one database per bounded context, keyed by the context
	// name (ContextIdentity, ContextCatalog, ...).
	Databases map[string]*mongo.Database

	// PrimaryContext is the context under test. MongoDB points at its database.
	PrimaryContext string

	// MongoDB is the database of the context under test. Tests that assert on
	// the service under test use this; cross-context seeds must go through
	// Coll/DBForCollection so they land in the owning database.
	MongoDB  *mongo.Database
	MongoURI string
	RedisURL string
	Ctx      context.Context
}

// SetupTestSuite starts the containers and creates one database per bounded
// context. primaryContext is the context under test; it becomes ts.MongoDB.
func SetupTestSuite(primaryContext string) (*TestSuite, error) {
	if _, ok := contextDatabase[primaryContext]; !ok {
		return nil, fmt.Errorf("unknown primary context %q", primaryContext)
	}

	ctx := context.Background()

	// Setup MongoDB
	mongoContainer, err := mongodb.Run(ctx,
		"mongo:7.0",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections").WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start mongo container: %w", err)
	}

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mongo connection string: %w", err)
	}

	// Setup Redis
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	redisURL, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis connection string: %w", err)
	}

	// Connect to MongoDB
	client, err := mongo.Connect(mongoopts.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	// Wait for connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	ts := &TestSuite{
		MongoContainer: mongoContainer,
		RedisContainer: redisContainer,
		MongoClient:    client,
		Databases:      make(map[string]*mongo.Database, len(contextDatabase)),
		PrimaryContext: primaryContext,
		MongoURI:       mongoURI,
		RedisURL:       redisURL,
		Ctx:            ctx,
	}

	for ctxName, dbName := range contextDatabase {
		ts.Databases[ctxName] = client.Database(dbName)
	}
	ts.MongoDB = ts.DB(primaryContext)

	// Run index creation
	if err := ts.EnsureIndexes(); err != nil {
		ts.Teardown()
		return nil, fmt.Errorf("failed to ensure indexes: %w", err)
	}

	// Seed minimal roles
	if err := ts.SeedMinimalRoles(); err != nil {
		ts.Teardown()
		return nil, fmt.Errorf("failed to seed minimal roles: %w", err)
	}

	return ts, nil
}

// contextIndexes declares the indexes each context needs on its own
// collections. Counters are created for every collection listed here, so a
// context's counters live in the same database as the documents they number.
var contextIndexes = map[string]map[string][]mongo.IndexModel{
	ContextIdentity: {
		"users":          {{Keys: bson.D{{Key: "email", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"roles":          {{Keys: bson.D{{Key: "role_name", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"user_roles":     {{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "role_id", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"refresh_tokens": {{Keys: bson.D{{Key: "user_id", Value: 1}}}},
		"reset_tokens":   {{Keys: bson.D{{Key: "user_id", Value: 1}}}},
	},
	ContextCatalog: {
		"categories": {{Keys: bson.D{{Key: "name", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"products":   {{Keys: bson.D{{Key: "merchant_id", Value: 1}}}, {Keys: bson.D{{Key: "category_id", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
	},
	ContextMerchant: {
		"merchants":          {{Keys: bson.D{{Key: "user_id", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"cashiers":           {{Keys: bson.D{{Key: "merchant_id", Value: 1}}}, {Keys: bson.D{{Key: "user_id", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"merchant_documents": {{Keys: bson.D{{Key: "merchant_id", Value: 1}}}},
	},
	ContextSales: {
		"orders":       {{Keys: bson.D{{Key: "merchant_id", Value: 1}}}, {Keys: bson.D{{Key: "cashier_id", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"order_items":  {{Keys: bson.D{{Key: "order_id", Value: 1}}}, {Keys: bson.D{{Key: "product_id", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
		"transactions": {{Keys: bson.D{{Key: "order_id", Value: 1}}}, {Keys: bson.D{{Key: "merchant_id", Value: 1}}}, {Keys: bson.D{{Key: "deleted_at", Value: 1}}}},
	},
	ContextEmail: {},
}

func (ts *TestSuite) EnsureIndexes() error {
	ctx := ts.Ctx

	for ctxName, collections := range contextIndexes {
		db := ts.DB(ctxName)

		for collName, models := range collections {
			coll := db.Collection(collName)
			if _, err := coll.Indexes().CreateMany(ctx, models); err != nil {
				log.Printf("Failed to create indexes for %s.%s: %v", ctxName, collName, err)
			}

			// Initialize the collection's counter in the owning database.
			if _, err := db.Collection("counters").UpdateOne(ctx,
				bson.M{"_id": collName},
				bson.M{"$setOnInsert": bson.M{"seq": int32(0)}},
				mongoopts.UpdateOne().SetUpsert(true),
			); err != nil {
				log.Printf("Failed to initialize counter for %s.%s: %v", ctxName, collName, err)
			}
		}
	}

	return nil
}

// SeedMinimalRoles inserts default roles for integration tests. Roles belong to
// the identity context, so they are written there regardless of which context
// the suite is exercising.
func (ts *TestSuite) SeedMinimalRoles() error {
	identityDB := ts.DB(ContextIdentity)
	coll := identityDB.Collection("roles")
	now := time.Now().UTC()

	for _, name := range []string{"ROLE_ADMIN", "ROLE_CASHIER"} {
		// Upsert: skip if already exists
		count, _ := coll.CountDocuments(ts.Ctx, bson.M{"role_name": name})
		if count > 0 {
			continue
		}

		roleID, err := database.GetNextID(ts.Ctx, identityDB, "roles")
		if err != nil {
			return err
		}

		_, err = coll.InsertOne(ts.Ctx, bson.M{
			"_id":        roleID,
			"role_name":  name,
			"created_at": now,
			"updated_at": now,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *TestSuite) RedisClient() *goredis.Client {
	opts, _ := goredis.ParseURL(ts.RedisURL)
	return goredis.NewClient(opts)
}

func (ts *TestSuite) Teardown() {
	if ts.MongoClient != nil {
		_ = ts.MongoClient.Disconnect(context.Background())
	}
	if ts.MongoContainer != nil {
		if err := ts.MongoContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate mongo container: %v", err)
		}
	}
	if ts.RedisContainer != nil {
		if err := ts.RedisContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate redis container: %v", err)
		}
	}
}

type GRPCServerRunner interface {
	Serve(lis net.Listener) error
	Stop()
	GracefulStop()
}

func RunGRPCServer(server *grpc.Server) (string, error) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("grpc server error: %v", err)
		}
	}()
	return lis.Addr().String(), nil
}
