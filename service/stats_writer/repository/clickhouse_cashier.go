package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Cashier stats ───────────────────────────────────────────────────────────

func (r *clickhouseRepository) InsertCashierEvent(ctx context.Context, eventID string, eventVersion uint64, event events.CashierEvent) error {
	query := `INSERT INTO cashier_daily (
		event_id, event_time, cashier_id, merchant_id, status, event_version
	)`
	return r.appendToBatch(ctx, "cashier", query,
		toUUID(eventID), parseEventTime(event.EventTime), uint64(event.CashierID),
		uint64(event.MerchantID), event.Status, eventVersion,
	)
}
