package repository

import "context"

// ── Cashier stats ───────────────────────────────────────────────────────────

type CashierMonthTotalSales struct {
	Year       string
	Month      string
	TotalSales int64
}

type CashierYearTotalSales struct {
	Year       string
	TotalSales int64
}

type CashierMonthSales struct {
	Month      string
	CashierID  uint64
	OrderCount uint64
	TotalSales int64
}

type CashierYearSales struct {
	Year       string
	CashierID  uint64
	OrderCount uint64
	TotalSales int64
}

// CashierStatsRepository serves the cashier stats rpc, reading order_daily.
type CashierStatsRepository interface {
	GetMonthlyTotalSales(ctx context.Context, year, month int) ([]CashierMonthTotalSales, error)
	GetYearlyTotalSales(ctx context.Context, year int) ([]CashierYearTotalSales, error)
	GetMonthSales(ctx context.Context, year int) ([]CashierMonthSales, error)
	GetYearSales(ctx context.Context, year int) ([]CashierYearSales, error)
	GetMonthlyTotalSalesByMerchant(ctx context.Context, year, month, merchantID int) ([]CashierMonthTotalSales, error)
	GetYearlyTotalSalesByMerchant(ctx context.Context, year, merchantID int) ([]CashierYearTotalSales, error)
	GetMonthSalesByMerchant(ctx context.Context, year, merchantID int) ([]CashierMonthSales, error)
	GetYearSalesByMerchant(ctx context.Context, year, merchantID int) ([]CashierYearSales, error)
	GetMonthlyTotalSalesById(ctx context.Context, year, month, cashierID int) ([]CashierMonthTotalSales, error)
	GetYearlyTotalSalesById(ctx context.Context, year, cashierID int) ([]CashierYearTotalSales, error)
	GetMonthSalesById(ctx context.Context, year, cashierID int) ([]CashierMonthSales, error)
	GetYearSalesById(ctx context.Context, year, cashierID int) ([]CashierYearSales, error)
}
