package repository

// Repository serves the stats gRPC contracts (F4 §7.4) straight from
// ClickHouse — 50 methods, 1:1 with the rpc of the 11 stats services. Queries
// never join a main DB: order/cashier stats read order_daily, category stats
// read order_item_daily, and transaction stats read transaction_daily. All
// reads use FINAL so ReplacingMergeTree dedupes at-least-once redeliveries
// before aggregating.
//
// The methods are grouped by domain into the four embedded interfaces so a
// consumer can depend on just the slice it serves. Repository itself stays as
// the full surface implemented by ClickHouseReaderRepository.
type Repository interface {
	OrderStatsRepository
	CashierStatsRepository
	CategoryStatsRepository
	TransactionStatsRepository
}
