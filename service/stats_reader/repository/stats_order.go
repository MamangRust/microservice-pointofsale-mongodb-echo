package repository

import "context"

// ── Order stats ─────────────────────────────────────────────────────────────
//
// Result models mirror the ClickHouse aggregations in clickhouse.go. Counts and
// amounts are int64/uint64: ClickHouse UInt64/Int64 columns scan into them
// directly (clickhouse-go v2 rejects implicit UInt64→int64 conversion).

type OrderMonthlyTotalRevenue struct {
	Year           string
	Month          string
	OrderCount     uint64
	TotalRevenue   int64
	TotalItemsSold uint64
}

type OrderYearlyTotalRevenue struct {
	Year               string
	OrderCount         uint64
	TotalRevenue       int64
	TotalItemsSold     uint64
	ActiveCashiers     uint64
	UniqueProductsSold uint64
}

type OrderMonthly struct {
	Month          string
	OrderCount     uint64
	TotalRevenue   int64
	TotalItemsSold uint64
}

type OrderYearly struct {
	Year               string
	OrderCount         uint64
	TotalRevenue       int64
	TotalItemsSold     uint64
	ActiveCashiers     uint64
	UniqueProductsSold uint64
}

// OrderStatsRepository serves the order stats rpc, reading order_daily (with
// total_items_sold/unique_products_sold merged in from order_item_daily).
type OrderStatsRepository interface {
	GetMonthlyTotalRevenue(ctx context.Context, year, month int) ([]OrderMonthlyTotalRevenue, error)
	GetYearlyTotalRevenue(ctx context.Context, year int) ([]OrderYearlyTotalRevenue, error)
	GetMonthlyRevenue(ctx context.Context, year int) ([]OrderMonthly, error)
	GetYearlyRevenue(ctx context.Context, year int) ([]OrderYearly, error)
	GetMonthlyTotalRevenueByMerchant(ctx context.Context, year, month, merchantID int) ([]OrderMonthlyTotalRevenue, error)
	GetYearlyTotalRevenueByMerchant(ctx context.Context, year, merchantID int) ([]OrderYearlyTotalRevenue, error)
	GetMonthlyRevenueByMerchant(ctx context.Context, year, merchantID int) ([]OrderMonthly, error)
	GetYearlyRevenueByMerchant(ctx context.Context, year, merchantID int) ([]OrderYearly, error)
	GetMonthlyTotalRevenueById(ctx context.Context, year, month, orderID int) ([]OrderMonthlyTotalRevenue, error)
	GetYearlyTotalRevenueById(ctx context.Context, year, orderID int) ([]OrderYearlyTotalRevenue, error)
}
