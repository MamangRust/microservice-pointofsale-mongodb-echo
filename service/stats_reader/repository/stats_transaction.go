package repository

import "context"

// ── Transaction stats ───────────────────────────────────────────────────────

type TransactionMonthAmount struct {
	Year        string
	Month       string
	TotalCount  uint64
	TotalAmount int64
}

type TransactionYearAmount struct {
	Year        string
	TotalCount  uint64
	TotalAmount int64
}

type TransactionMonthMethod struct {
	Month             string
	PaymentMethod     string
	TotalTransactions uint64
	TotalAmount       int64
}

type TransactionYearMethod struct {
	Year              string
	PaymentMethod     string
	TotalTransactions uint64
	TotalAmount       int64
}

// TransactionStatsRepository serves the transaction status and payment-method
// stats rpc, reading transaction_daily.
type TransactionStatsRepository interface {
	// Status stats
	GetMonthStatusSuccess(ctx context.Context, year, month int) ([]TransactionMonthAmount, error)
	GetYearStatusSuccess(ctx context.Context, year int) ([]TransactionYearAmount, error)
	GetMonthStatusFailed(ctx context.Context, year, month int) ([]TransactionMonthAmount, error)
	GetYearStatusFailed(ctx context.Context, year int) ([]TransactionYearAmount, error)
	GetMonthStatusSuccessByMerchant(ctx context.Context, year, month, merchantID int) ([]TransactionMonthAmount, error)
	GetYearStatusSuccessByMerchant(ctx context.Context, year, merchantID int) ([]TransactionYearAmount, error)
	GetMonthStatusFailedByMerchant(ctx context.Context, year, month, merchantID int) ([]TransactionMonthAmount, error)
	GetYearStatusFailedByMerchant(ctx context.Context, year, merchantID int) ([]TransactionYearAmount, error)

	// Payment-method stats
	GetMonthMethodSuccess(ctx context.Context, year, month int) ([]TransactionMonthMethod, error)
	GetYearMethodSuccess(ctx context.Context, year int) ([]TransactionYearMethod, error)
	GetMonthMethodFailed(ctx context.Context, year, month int) ([]TransactionMonthMethod, error)
	GetYearMethodFailed(ctx context.Context, year int) ([]TransactionYearMethod, error)
	GetMonthMethodByMerchantSuccess(ctx context.Context, year, month, merchantID int) ([]TransactionMonthMethod, error)
	GetYearMethodByMerchantSuccess(ctx context.Context, year, merchantID int) ([]TransactionYearMethod, error)
	GetMonthMethodByMerchantFailed(ctx context.Context, year, month, merchantID int) ([]TransactionMonthMethod, error)
	GetYearMethodByMerchantFailed(ctx context.Context, year, merchantID int) ([]TransactionYearMethod, error)
}
