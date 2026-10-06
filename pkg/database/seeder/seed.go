package seeder

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Databases holds one handle per bounded context. Seeder steps read and write
// the database that owns each collection.
type Databases struct {
	Identity *mongo.Database
	Catalog  *mongo.Database
	Sales    *mongo.Database
	Merchant *mongo.Database
	Email    *mongo.Database
}

type Deps struct {
	DBs    Databases
	Ctx    context.Context
	Logger logger.LoggerInterface
	Hash   hash.HashPassword
}

type Seeder struct {
	User        *userSeeder
	Role        *roleSeeder
	UserRole    *userRoleSeeder
	Cashier     *cashierSeeder
	Category    *categorySeeder
	Product     *productSeeder
	Merchant    *merchantSeeder
	Order       *orderSeeder
	Transaction *transactionSeeder
}

func NewSeeder(deps Deps) *Seeder {
	dbs := deps.DBs
	return &Seeder{
		User:        NewUserSeeder(dbs, deps.Hash, deps.Ctx, deps.Logger),
		Role:        NewRoleSeeder(dbs, deps.Ctx, deps.Logger),
		UserRole:    NewUserRoleSeeder(dbs, deps.Ctx, deps.Logger),
		Merchant:    NewMerchantSeeder(dbs, deps.Ctx, deps.Logger),
		Cashier:     NewCashierSeeder(dbs, deps.Ctx, deps.Logger),
		Category:    NewCategorySeeder(dbs, deps.Ctx, deps.Logger),
		Product:     NewProductSeeder(dbs, deps.Ctx, deps.Logger),
		Order:       NewOrderSeeder(dbs, deps.Ctx, deps.Logger),
		Transaction: NewTransactionSeeder(dbs, deps.Ctx, deps.Logger),
	}
}

func (s *Seeder) Run() error {
	if err := s.seedWithDelay("users", s.User.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("roles", s.Role.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("user_roles", s.UserRole.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("merchant", s.Merchant.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("cashier", s.Cashier.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("category", s.Category.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("product", s.Product.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("order", s.Order.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("transaction", s.Transaction.Seed); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedWithDelay(entityName string, seedFunc func() error) error {
	if err := seedFunc(); err != nil {
		return fmt.Errorf("failed to seed %s: %w", entityName, err)
	}
	time.Sleep(seedDelay())
	return nil
}

func seedDelay() time.Duration {
	if raw := os.Getenv("SEED_DELAY_SECONDS"); raw != "" {
		if secs, err := strconv.Atoi(raw); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return 30 * time.Second
}

func ptrString(s string) *string {
	return &s
}

func ptrInt32(i int32) *int32 {
	return &i
}
