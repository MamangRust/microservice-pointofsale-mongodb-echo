package repository

import "context"

// ── Category stats ──────────────────────────────────────────────────────────

type CategoryMonthTotalPrice struct {
	Year         string
	Month        string
	TotalRevenue int64
}

type CategoryYearTotalPrice struct {
	Year         string
	TotalRevenue int64
}

type CategoryMonthPrice struct {
	Month        string
	CategoryID   uint64
	OrderCount   uint64
	ItemsSold    uint64
	TotalRevenue int64
}

type CategoryYearPrice struct {
	Year               string
	CategoryID         uint64
	OrderCount         uint64
	ItemsSold          uint64
	TotalRevenue       int64
	UniqueProductsSold uint64
}

// CategoryStatsRepository serves the category stats rpc, reading
// order_item_daily.
type CategoryStatsRepository interface {
	GetMonthlyTotalPrices(ctx context.Context, year, month int) ([]CategoryMonthTotalPrice, error)
	GetYearlyTotalPrices(ctx context.Context, year int) ([]CategoryYearTotalPrice, error)
	GetMonthPrice(ctx context.Context, year int) ([]CategoryMonthPrice, error)
	GetYearPrice(ctx context.Context, year int) ([]CategoryYearPrice, error)
	GetMonthlyTotalPricesByMerchant(ctx context.Context, year, month, merchantID int) ([]CategoryMonthTotalPrice, error)
	GetYearlyTotalPricesByMerchant(ctx context.Context, year, merchantID int) ([]CategoryYearTotalPrice, error)
	GetMonthPriceByMerchant(ctx context.Context, year, merchantID int) ([]CategoryMonthPrice, error)
	GetYearPriceByMerchant(ctx context.Context, year, merchantID int) ([]CategoryYearPrice, error)
	GetMonthlyTotalPricesById(ctx context.Context, year, month, categoryID int) ([]CategoryMonthTotalPrice, error)
	GetYearlyTotalPricesById(ctx context.Context, year, categoryID int) ([]CategoryYearTotalPrice, error)
	GetMonthPriceById(ctx context.Context, year, categoryID int) ([]CategoryMonthPrice, error)
	GetYearPriceById(ctx context.Context, year, categoryID int) ([]CategoryYearPrice, error)
}
