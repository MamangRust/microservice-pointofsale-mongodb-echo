// Package backfill implements the stats-writer `backfill` command: it walks the
// historical OLTP data through the owning services' gRPC APIs and materializes
// it into ClickHouse through the same batch repository used for live events.
//
// Every source is read from the service that owns it — orders from order,
// transactions from transaction, order items from order_item, products (for
// the product → category denormalization) from product, and cashiers from
// cashier. No collection is touched directly, mirroring the "one owner per
// collection" rule enforced for the live path. This is the bootstrap path for
// the stats pipeline: it lets the ClickHouse tables reflect pre-existing data
// without replaying every domain event.
package backfill

import (
	"context"
	"fmt"
	"time"

	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	orderadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	transactionadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// backfillPageSize is how many rows are requested per gRPC call while walking
// a collection.
const backfillPageSize = 1000

// backfillEventID derives a deterministic UUID per entity so re-running the
// backfill replaces the same ReplacingMergeTree key (with a newer version)
// instead of appending duplicates.
func backfillEventID(kind string, id int32) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(fmt.Sprintf("backfill:%s:%d", kind, id))).String()
}

// Backfiller streams historical data into ClickHouse through the owning
// services' bulk (enumeration) gRPC APIs.
type Backfiller struct {
	log          logger.LoggerInterface
	repo         repository.Repository
	orders       orderadapter.BulkRepository
	orderItems   orderitemadapter.BulkRepository
	products     productadapter.BulkRepository
	transactions transactionadapter.BulkRepository
	cashiers     cashieradapter.BulkRepository
}

// New returns a ready Backfiller. Each adapter is the narrow bulk surface of
// the service that owns the corresponding collection.
func New(
	log logger.LoggerInterface,
	repo repository.Repository,
	orders orderadapter.BulkRepository,
	orderItems orderitemadapter.BulkRepository,
	products productadapter.BulkRepository,
	transactions transactionadapter.BulkRepository,
	cashiers cashieradapter.BulkRepository,
) *Backfiller {
	return &Backfiller{
		log:          log,
		repo:         repo,
		orders:       orders,
		orderItems:   orderItems,
		products:     products,
		transactions: transactions,
		cashiers:     cashiers,
	}
}

// Run streams all stats sources into ClickHouse.
func (b *Backfiller) Run(ctx context.Context) error {
	version := uint64(time.Now().Unix())
	counts := map[string]int{}

	// The transactions collection does not store cashier_id (the transaction
	// service resolves it through the order it belongs to). While walking the
	// orders we record order_id → cashier_id so the transaction pass can fill
	// it without an extra RPC per row.
	cashiers := map[int32]int32{}

	if err := b.backfillOrders(ctx, version, counts, cashiers); err != nil {
		return err
	}
	if err := b.backfillOrderItems(ctx, version, counts); err != nil {
		return err
	}
	if err := b.backfillTransactions(ctx, version, counts, cashiers); err != nil {
		return err
	}
	if err := b.backfillCashiers(ctx, version, counts); err != nil {
		return err
	}

	if err := b.repo.Flush(ctx); err != nil {
		return fmt.Errorf("flush backfill batches: %w", err)
	}

	b.log.Info("backfill complete",
		zap.Int("orders", counts["order"]),
		zap.Int("order_items", counts["order_item"]),
		zap.Int("transactions", counts["transaction"]),
		zap.Int("cashiers", counts["cashier"]),
	)
	return nil
}

func (b *Backfiller) backfillOrders(ctx context.Context, version uint64, counts map[string]int, cashiers map[int32]int32) error {
	for page := 1; ; page++ {
		orders, total, err := b.orders.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query orders: %w", err)
		}
		if page == 1 {
			b.log.Info("backfilling orders", zap.Int("total_active", total))
		}

		for _, o := range orders {
			eventTime := ""
			if o.CreatedAt != nil {
				eventTime = o.CreatedAt.UTC().Format(time.RFC3339)
			}
			event := events.OrderEvent{
				OrderID:    o.OrderID,
				CashierID:  o.CashierID,
				MerchantID: o.MerchantID,
				TotalPrice: o.TotalPrice,
				Status:     "created",
				EventTime:  eventTime,
			}
			if err := b.repo.InsertOrderEvent(ctx, backfillEventID("order", o.OrderID), version, event); err != nil {
				return fmt.Errorf("insert order %d: %w", o.OrderID, err)
			}
			cashiers[o.OrderID] = o.CashierID
			counts["order"]++
		}

		if len(orders) < backfillPageSize {
			break
		}
	}
	return nil
}

// backfillOrderItems walks every active order item through the order_item
// service. The product → category mapping is loaded from the product service so
// the category_id denormalization still happens here (ClickHouse rows carry it
// for the category statistics).
func (b *Backfiller) backfillOrderItems(ctx context.Context, version uint64, counts map[string]int) error {
	categories, err := b.loadProductCategories(ctx)
	if err != nil {
		return err
	}

	for page := 1; ; page++ {
		items, total, err := b.orderItems.FindAllOrderItems(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query order items: %w", err)
		}
		if page == 1 {
			b.log.Info("backfilling order items", zap.Int("total_active", total))
		}

		for _, item := range items {
			categoryID, ok := categories[item.ProductID]
			if !ok {
				// Product is missing or soft-deleted; skip, matching the previous
				// $lookup + $unwind (preserveNullAndEmptyArrays: false) behaviour.
				continue
			}

			eventTime := ""
			if item.CreatedAt != nil {
				eventTime = item.CreatedAt.UTC().Format(time.RFC3339)
			}

			unitPrice := int32(item.Price)
			event := events.OrderItemEvent{
				OrderItemID: item.OrderItemID,
				OrderID:     item.OrderID,
				ProductID:   item.ProductID,
				CategoryID:  categoryID,
				Quantity:    item.Quantity,
				UnitPrice:   unitPrice,
				Subtotal:    item.Quantity * unitPrice,
				EventTime:   eventTime,
			}
			if err := b.repo.InsertOrderItemEvent(ctx, backfillEventID("order_item", item.OrderItemID), version, event); err != nil {
				return fmt.Errorf("insert order item %d: %w", item.OrderItemID, err)
			}
			counts["order_item"]++
		}

		// Stop on the first short page. Terminating on the reported total instead
		// would silently truncate the backfill if the server ever stopped sending
		// pagination metadata.
		if len(items) < backfillPageSize {
			break
		}
	}
	return nil
}

// loadProductCategories returns product_id → category_id for every active
// product, read through the product service that owns the catalog.
func (b *Backfiller) loadProductCategories(ctx context.Context) (map[int32]int32, error) {
	categories := make(map[int32]int32)
	for page := 1; ; page++ {
		products, _, err := b.products.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return nil, fmt.Errorf("query products: %w", err)
		}
		for _, p := range products {
			categories[p.ProductID] = p.CategoryID
		}
		if len(products) < backfillPageSize {
			break
		}
	}
	return categories, nil
}

func (b *Backfiller) backfillTransactions(ctx context.Context, version uint64, counts map[string]int, cashiers map[int32]int32) error {
	for page := 1; ; page++ {
		transactions, total, err := b.transactions.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query transactions: %w", err)
		}
		if page == 1 {
			b.log.Info("backfilling transactions", zap.Int("total_active", total))
		}

		for _, t := range transactions {
			eventTime := ""
			if t.CreatedAt != nil {
				eventTime = t.CreatedAt.UTC().Format(time.RFC3339)
			}
			status := ""
			if t.PaymentStatus != nil {
				status = *t.PaymentStatus
			}

			event := events.TransactionEvent{
				TransactionID: t.TransactionID,
				OrderID:       t.OrderID,
				CashierID:     cashiers[t.OrderID],
				MerchantID:    t.MerchantID,
				PaymentMethod: t.PaymentMethod,
				Amount:        t.Amount,
				Status:        status,
				EventTime:     eventTime,
			}
			if err := b.repo.InsertTransactionEvent(ctx, backfillEventID("transaction", t.TransactionID), version, event); err != nil {
				return fmt.Errorf("insert transaction %d: %w", t.TransactionID, err)
			}
			counts["transaction"]++
		}

		if len(transactions) < backfillPageSize {
			break
		}
	}
	return nil
}

// backfillCashiers walks every active cashier through the cashier service and
// materializes the cashier dimension into cashier_daily. The cashier *sales*
// aggregates are served from order_daily; this pass only records who the
// cashiers are (id, merchant, status).
func (b *Backfiller) backfillCashiers(ctx context.Context, version uint64, counts map[string]int) error {
	for page := 1; ; page++ {
		cashiers, total, err := b.cashiers.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query cashiers: %w", err)
		}
		if page == 1 {
			b.log.Info("backfilling cashiers", zap.Int("total_active", total))
		}

		for _, c := range cashiers {
			eventTime := ""
			if c.CreatedAt != nil {
				eventTime = c.CreatedAt.UTC().Format(time.RFC3339)
			}
			event := events.CashierEvent{
				CashierID:  c.CashierID,
				MerchantID: c.MerchantID,
				Status:     "active",
				EventTime:  eventTime,
			}
			if err := b.repo.InsertCashierEvent(ctx, backfillEventID("cashier", c.CashierID), version, event); err != nil {
				return fmt.Errorf("insert cashier %d: %w", c.CashierID, err)
			}
			counts["cashier"]++
		}

		if len(cashiers) < backfillPageSize {
			break
		}
	}
	return nil
}
