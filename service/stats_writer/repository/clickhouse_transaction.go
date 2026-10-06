package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Transaction stats ───────────────────────────────────────────────────────

func (r *clickhouseRepository) InsertTransactionEvent(ctx context.Context, eventID string, eventVersion uint64, event events.TransactionEvent) error {
	query := `INSERT INTO transaction_daily (
		event_id, event_time, transaction_id, order_id, cashier_id, merchant_id,
		payment_method, status, amount, event_version
	)`
	return r.appendToBatch(ctx, "transaction", query,
		toUUID(eventID), parseEventTime(event.EventTime), uint64(event.TransactionID), uint64(event.OrderID),
		uint64(event.CashierID), uint64(event.MerchantID), event.PaymentMethod, event.Status,
		int64(event.Amount), eventVersion,
	)
}
