// Package stats_writer_test exercises the stats-writer backfill end to end: it
// seeds the owning MongoDB collections (catalog + sales contexts), starts the
// real order/transaction/product/order_item gRPC services, runs the backfiller
// — which reads every source through those services — against a real
// ClickHouse, and asserts the materialized rows.
package stats_writer_test

import (
	"fmt"
	"testing"
	"time"

	chDriver "github.com/ClickHouse/clickhouse-go/v2"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	orderadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	transactionadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/transaction"
	pkgclickhouse "github.com/MamangRust/microservice-point-of-sale-pkg/clickhouse"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/backfill"
	statsrepo "github.com/MamangRust/microservice-point-of-sale-stats-writer/repository"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	// validItemCount is deliberately larger than the backfiller's 1000-row page
	// size so the walk has to cross the page boundary and stop on a short page.
	validItemCount = 1005
	orphanProduct  = 999999

	productOnePrice = 1000
	productTwoPrice = 2000
	orderTotalPrice = 1507000

	// seedCashierID is the cashier on the seeded order. Transactions carry no
	// cashier_id themselves; the backfill must recover it from the order.
	seedCashierID = 7
)

type BackfillSuite struct {
	tests.BaseTestSuite

	chConn chDriver.Conn
	repo   statsrepo.Repository

	orderID    int32
	categoryA  int32
	categoryB  int32
	productOne int32
	productTwo int32
	cashierOne int32
	cashierTwo int32

	// expectations derived from the seeded documents (the source of truth)
	expMinItemID   int32
	expMaxItemID   int32
	expSumSubtotal int64
	expCategoryA   uint64
	expCategoryB   uint64
}

func (s *BackfillSuite) SetupSuite() {
	// Orders, order items and transactions live in the sales context, which is
	// the suite's primary database; catalog data is seeded cross-context through
	// Coll/DBForCollection.
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()
	s.SetupStatsEnv()

	// Brings up order (and its cashier/merchant/product/order_item deps),
	// transaction and product services over real gRPC.
	s.SetupTransactionService()
	s.SetupProductService()

	conn, err := pkgclickhouse.NewClient(s.Log)
	s.Require().NoError(err)
	s.Require().NoError(pkgclickhouse.ApplySchema(s.Ctx, conn, s.Log))

	s.chConn = conn
	s.repo = statsrepo.NewClickhouseRepository(conn, s.Log)

	s.seedMongo()
}

func (s *BackfillSuite) TearDownSuite() {
	if s.repo != nil {
		_ = s.repo.Close()
	}
	if s.chConn != nil {
		_ = s.chConn.Close()
	}
	s.BaseTestSuite.TearDownSuite()
}

// seedMongo writes the fixture rows into the database that owns each
// collection: categories/products into the catalog context,
// orders/order_items/transactions into the sales context, and cashiers into the
// merchant context. The backfill itself never reads Mongo — it goes through the
// owning services — so this is purely a data setup step.
func (s *BackfillSuite) seedMongo() {
	ctx := s.Ctx
	now := time.Now().UTC()

	// Catalog context: categories → products.
	catColl := s.Coll("categories")
	s.categoryA, _ = database.GetNextID(ctx, s.DBForCollection("categories"), "categories")
	s.categoryB, _ = database.GetNextID(ctx, s.DBForCollection("categories"), "categories")
	_, err := catColl.InsertMany(ctx, []interface{}{
		bson.M{"_id": s.categoryA, "name": "Cat A", "description": "a", "created_at": now, "updated_at": now},
		bson.M{"_id": s.categoryB, "name": "Cat B", "description": "b", "created_at": now, "updated_at": now},
	})
	s.Require().NoError(err)

	prodColl := s.Coll("products")
	s.productOne, _ = database.GetNextID(ctx, s.DBForCollection("products"), "products")
	s.productTwo, _ = database.GetNextID(ctx, s.DBForCollection("products"), "products")
	_, err = prodColl.InsertMany(ctx, []interface{}{
		bson.M{"_id": s.productOne, "merchant_id": int32(1), "category_id": s.categoryA, "name": "P1", "price": int32(productOnePrice), "count_in_stock": int32(100), "created_at": now, "updated_at": now},
		bson.M{"_id": s.productTwo, "merchant_id": int32(1), "category_id": s.categoryB, "name": "P2", "price": int32(productTwoPrice), "count_in_stock": int32(100), "created_at": now, "updated_at": now},
	})
	s.Require().NoError(err)

	// Sales context: one order.
	orderColl := s.Coll("orders")
	s.orderID, _ = database.GetNextID(ctx, s.DBForCollection("orders"), "orders")
	_, err = orderColl.InsertOne(ctx, bson.M{
		"_id": s.orderID, "merchant_id": int32(1), "cashier_id": int32(seedCashierID),
		"total_price": int64(orderTotalPrice), "created_at": now, "updated_at": now,
	})
	s.Require().NoError(err)

	// Sales context: order items. Odd rows use product one, even rows product
	// two, so both catalog categories are exercised; the orphan points at a
	// product that does not exist and must be skipped.
	itemColl := s.Coll("order_items")
	items := make([]interface{}, 0, validItemCount+1)
	for i := 0; i < validItemCount; i++ {
		id, idErr := database.GetNextID(ctx, s.DBForCollection("order_items"), "order_items")
		s.Require().NoError(idErr)

		productID := s.productOne
		price := int64(productOnePrice)
		if i%2 == 1 {
			productID = s.productTwo
			price = productTwoPrice
		}

		if i == 0 {
			s.expMinItemID = id
		}
		s.expMaxItemID = id
		s.expSumSubtotal += price
		if productID == s.productOne {
			s.expCategoryA++
		} else {
			s.expCategoryB++
		}

		items = append(items, bson.M{
			"_id": id, "order_id": s.orderID, "product_id": productID,
			"quantity": int32(1), "price": price, "created_at": now, "updated_at": now,
		})
	}
	orphanID, idErr := database.GetNextID(ctx, s.DBForCollection("order_items"), "order_items")
	s.Require().NoError(idErr)
	items = append(items, bson.M{
		"_id": orphanID, "order_id": s.orderID, "product_id": int32(orphanProduct),
		"quantity": int32(1), "price": int64(500), "created_at": now, "updated_at": now,
	})
	_, err = itemColl.InsertMany(ctx, items)
	s.Require().NoError(err)

	// Sales context: two transactions for the order.
	txColl := s.Coll("transactions")
	txOne, _ := database.GetNextID(ctx, s.DBForCollection("transactions"), "transactions")
	txTwo, _ := database.GetNextID(ctx, s.DBForCollection("transactions"), "transactions")
	paid := "success"
	_, err = txColl.InsertMany(ctx, []interface{}{
		bson.M{"_id": txOne, "order_id": s.orderID, "merchant_id": int32(1), "payment_method": "cash", "amount": int32(1000000), "payment_status": paid, "created_at": now, "updated_at": now},
		bson.M{"_id": txTwo, "order_id": s.orderID, "merchant_id": int32(1), "payment_method": "transfer", "amount": int32(500000), "payment_status": paid, "created_at": now, "updated_at": now},
	})
	s.Require().NoError(err)

	// Merchant context: two cashiers, materialized into cashier_daily.
	cashierColl := s.Coll("cashiers")
	s.cashierOne, _ = database.GetNextID(ctx, s.DBForCollection("cashiers"), "cashiers")
	s.cashierTwo, _ = database.GetNextID(ctx, s.DBForCollection("cashiers"), "cashiers")
	_, err = cashierColl.InsertMany(ctx, []interface{}{
		bson.M{"_id": s.cashierOne, "merchant_id": int32(1), "user_id": int32(1), "name": "C1", "created_at": now, "updated_at": now},
		bson.M{"_id": s.cashierTwo, "merchant_id": int32(1), "user_id": int32(1), "name": "C2", "created_at": now, "updated_at": now},
	})
	s.Require().NoError(err)
}

func (s *BackfillSuite) TestBackfillMaterializesEverySource() {
	orderRepo := orderadapter.New(pborder.NewOrderServiceClient(s.Conns["order"]))
	orderItemRepo := orderitemadapter.New(pborderitem.NewOrderItemServiceClient(s.Conns["order-item"]))
	productRepo := productadapter.New(pbproduct.NewProductServiceClient(s.Conns["product"]))
	transactionRepo := transactionadapter.New(pbtransaction.NewTransactionServiceClient(s.Conns["transaction"]))
	cashierRepo := cashieradapter.New(pbcashier.NewCashierServiceClient(s.Conns["cashier"]))

	bf := backfill.New(s.Log, s.repo, orderRepo, orderItemRepo, productRepo, transactionRepo, cashierRepo)
	s.Require().NoError(bf.Run(s.Ctx))

	s.Equal(uint64(validItemCount), s.count(`SELECT count() FROM order_item_daily FINAL`),
		"the orphan item must be skipped and both pages must be walked")

	var minID, maxID uint64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx,
		`SELECT min(order_item_id), max(order_item_id) FROM order_item_daily FINAL`,
	).Scan(&minID, &maxID))
	s.Equal(uint64(s.expMinItemID), minID, "first page must be reached")
	s.Equal(uint64(s.expMaxItemID), maxID, "second page must be reached")

	s.Equal(s.expCategoryA,
		s.count(fmt.Sprintf(`SELECT count() FROM order_item_daily FINAL WHERE category_id = %d`, s.categoryA)),
		"category must be denormalized from the catalog")
	s.Equal(s.expCategoryB,
		s.count(fmt.Sprintf(`SELECT count() FROM order_item_daily FINAL WHERE category_id = %d`, s.categoryB)),
		"category must be denormalized from the catalog")

	s.Equal(uint64(0), s.count(
		`SELECT count() FROM order_item_daily FINAL WHERE subtotal != toInt64(quantity) * unit_price`),
		"subtotal must equal quantity * unit_price")

	var sumSubtotal int64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx,
		`SELECT sum(subtotal) FROM order_item_daily FINAL`).Scan(&sumSubtotal))
	s.Equal(s.expSumSubtotal, sumSubtotal)

	s.Equal(uint64(0),
		s.count(fmt.Sprintf(`SELECT count() FROM order_item_daily FINAL WHERE product_id = %d`, orphanProduct)),
		"orphan order item must not be materialized")

	s.Equal(uint64(1), s.count(`SELECT count() FROM order_daily FINAL`))

	var totalPrice int64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx,
		`SELECT total_price FROM order_daily FINAL LIMIT 1`).Scan(&totalPrice))
	s.Equal(int64(orderTotalPrice), totalPrice)

	s.Equal(uint64(2), s.count(`SELECT count() FROM transaction_daily FINAL`))

	// The transaction rows must carry the cashier recovered from their order.
	s.Equal(uint64(2),
		s.count(fmt.Sprintf(`SELECT count() FROM transaction_daily FINAL WHERE cashier_id = %d`, seedCashierID)),
		"cashier_id must be recovered from the owning order")

	// Cashiers are materialized from the cashier service into cashier_daily.
	s.Equal(uint64(2), s.count(`SELECT count() FROM cashier_daily FINAL`),
		"both cashiers must be materialized")
	s.Equal(uint64(2), s.count(
		`SELECT count() FROM cashier_daily FINAL WHERE merchant_id = 1 AND status = 'active'`),
		"cashier rows must carry merchant and active status")
}

func (s *BackfillSuite) count(query string) uint64 {
	var n uint64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx, query).Scan(&n))
	return n
}

func TestBackfillSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	suite.Run(t, new(BackfillSuite))
}
