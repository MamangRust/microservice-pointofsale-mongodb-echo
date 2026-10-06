package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Order stats ─────────────────────────────────────────────────────────────

func (r *clickhouseRepository) InsertOrderEvent(ctx context.Context, eventID string, eventVersion uint64, event events.OrderEvent) error {
	query := `INSERT INTO order_daily (
		event_id, event_time, order_id, cashier_id, merchant_id, status, total_price, event_version
	)`
	return r.appendToBatch(ctx, "order", query,
		toUUID(eventID), parseEventTime(event.EventTime), uint64(event.OrderID), uint64(event.CashierID),
		uint64(event.MerchantID), event.Status, int64(event.TotalPrice), eventVersion,
	)
}
