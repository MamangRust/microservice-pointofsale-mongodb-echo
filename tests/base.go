package tests

import (
	"context"
	"reflect"
	"time"

	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type BaseTestSuite struct {
	suite.Suite

	// Context is the bounded context under test. Embedding suites must set it
	// before calling SetupSuite, so the suite's primary database is the one the
	// service under test actually owns.
	Context string

	ts      *TestSuite
	Log     logger.LoggerInterface
	Obs     observability.TraceLoggerObservability
	Conns   map[string]*grpc.ClientConn
	Servers []*grpc.Server
	Ctx     context.Context
	Cancel  context.CancelFunc

	// statsContainer is the ClickHouse testcontainer started by SetupStatsEnv,
	// terminated in TearDownSuite.
	statsContainer testcontainers.Container
}

func (s *BaseTestSuite) SetupSuite() {
	s.Require().NotEmpty(s.Context, "embedding suite must set Context before calling BaseTestSuite.SetupSuite")

	s.Ctx, s.Cancel = context.WithCancel(context.Background())

	ts, err := SetupTestSuite(s.Context)
	s.ts = ts
	s.Require().NoError(err)

	s.Log, _ = logger.NewLogger("test", nil)

	if s.Log == nil || (reflect.ValueOf(s.Log).Kind() == reflect.Ptr && reflect.ValueOf(s.Log).IsNil()) {
		z, _ := zap.NewDevelopment()
		s.Log = &logger.Logger{Log: z}
	}

	s.Obs, err = observability.NewObservability("test", s.Log)
	s.Require().NoError(err)
	s.Require().NotNil(s.Obs)
	s.Conns = make(map[string]*grpc.ClientConn)
}

func (s *BaseTestSuite) TearDownSuite() {
	if s.statsContainer != nil {
		_ = s.statsContainer.Terminate(context.Background())
	}
	for _, conn := range s.Conns {
		conn.Close()
	}
	for _, server := range s.Servers {
		server.GracefulStop()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
	if s.Cancel != nil {
		s.Cancel()
	}
}

func (s *BaseTestSuite) MongoDB() *mongo.Database {
	return s.ts.MongoDB
}

// MongoURI is the testcontainers connection string, for components that open
// their own Mongo connection instead of accepting a *mongo.Database.
func (s *BaseTestSuite) MongoURI() string {
	return s.ts.MongoURI
}

func (s *BaseTestSuite) RedisClient() *goredis.Client {
	return s.ts.RedisClient()
}

func (s *BaseTestSuite) RegisterServer(server *grpc.Server) string {
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Servers = append(s.Servers, server)
	return addr
}

func (s *BaseTestSuite) GetConnection(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	s.Require().NoError(err)
	return conn
}

func (s *BaseTestSuite) SeedUser(ctx context.Context) int {
	// Seed roles if not present (roles belong to the identity context)
	coll := s.Coll("roles")
	now := time.Now().UTC()
	for _, name := range []string{"ROLE_ADMIN", "Admin Access 1"} {
		count, _ := coll.CountDocuments(ctx, bson.M{"role_name": name})
		if count == 0 {
			id, _ := database.GetNextID(ctx, s.DBForCollection("roles"), "roles")
			coll.InsertOne(ctx, bson.M{"_id": id, "role_name": name, "created_at": now, "updated_at": now})
		}
	}

	res, err := pbuser.NewUserServiceClient(s.Conns["user"]).Create(ctx, &pbuser.CreateUserRequest{
		Firstname:       "Seed",
		Lastname:        "User",
		Email:           "seed.user@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedCategory(ctx context.Context) int {
	res, err := pbcategory.NewCategoryServiceClient(s.Conns["category"]).Create(ctx, &pbcategory.CreateCategoryRequest{
		Name:        "Seed Category",
		Description: "Seed Description",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedMerchant(ctx context.Context, userID int) int {
	res, err := pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]).Create(ctx, &pbmerchant.CreateMerchantRequest{
		UserId:       int32(userID),
		Name:         "Seed Merchant",
		Description:  "Seed Description",
		Address:      "Seed Address",
		ContactEmail: "merchant@example.com",
		ContactPhone: "08123456789",
		Status:       "active",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedProduct(ctx context.Context, merchantID int, categoryID int) int {
	res, err := pbproduct.NewProductServiceClient(s.Conns["product"]).Create(ctx, &pbproduct.CreateProductRequest{
		MerchantId:   int32(merchantID),
		CategoryId:   int32(categoryID),
		Name:         "Seed Product",
		Description:  "Seed Description",
		Price:        10000,
		CountInStock: 100,
		Brand:        "Seed Brand",
		Weight:       1000,
		ImageProduct: "seed.jpg",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedOrder(ctx context.Context, userID int, merchID int, prodID int) int {
	// Find or create cashier (cashiers belong to the merchant context)
	coll := s.Coll("cashiers")
	var cashierDoc bson.M
	err := coll.FindOne(ctx, bson.M{"user_id": userID, "merchant_id": merchID}).Decode(&cashierDoc)
	if err != nil {
		cashierID, _ := database.GetNextID(ctx, s.DBForCollection("cashiers"), "cashiers")
		now := time.Now().UTC()
		coll.InsertOne(ctx, bson.M{
			"_id":         cashierID,
			"merchant_id": merchID,
			"user_id":     userID,
			"name":        "Seed Cashier",
			"created_at":  now,
			"updated_at":  now,
		})
	}

	// Create order via gRPC
	res, err := pborder.NewOrderServiceClient(s.Conns["order"]).Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(merchID),
		CashierId:  int32(userID), // will be replaced with actual cashier_id
		Items: []*pborder.CreateOrderItemRequest{
			{
				ProductId: int32(prodID),
				Quantity:  1,
			},
		},
	})
	if err != nil {
		return 0
	}
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedOrderItem(ctx context.Context, orderID, productID int) int {
	coll := s.Coll("order_items")
	orderItemID, _ := database.GetNextID(ctx, s.DBForCollection("order_items"), "order_items")
	now := time.Now().UTC()
	_, err := coll.InsertOne(ctx, bson.M{
		"_id":        orderItemID,
		"order_id":   int32(orderID),
		"product_id": int32(productID),
		"quantity":   int32(1),
		"price":      int32(10000),
		"created_at": now,
		"updated_at": now,
	})
	if err != nil {
		return 0
	}
	return int(orderItemID)
}
