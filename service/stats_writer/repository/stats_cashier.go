package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Cashier stats ───────────────────────────────────────────────────────────
//
// CashierStatsRepository writes the cashier events that feed cashier_daily,
// mirroring the reader's CashierStatsRepository. The cashier *sales* aggregates
// are derived from order_daily (written by OrderStatsRepository); cashier_daily
// carries the cashier dimension (id, merchant, status) on its own.
type CashierStatsRepository interface {
	InsertCashierEvent(ctx context.Context, eventID string, eventVersion uint64, event events.CashierEvent) error
}
