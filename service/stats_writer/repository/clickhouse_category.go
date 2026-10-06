package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Category stats ──────────────────────────────────────────────────────────

func (r *clickhouseRepository) InsertOrderItemEvent(ctx context.Context, eventID string, eventVersion uint64, event events.OrderItemEvent) error {
	query := `INSERT INTO order_item_daily (
		event_id, event_time, order_item_id, order_id, product_id, category_id,
		quantity, unit_price, subtotal, event_version
	)`
	return r.appendToBatch(ctx, "order_item", query,
		toUUID(eventID), parseEventTime(event.EventTime), uint64(event.OrderItemID), uint64(event.OrderID),
		uint64(event.ProductID), uint64(event.CategoryID), uint32(event.Quantity),
		int64(event.UnitPrice), int64(event.Subtotal), eventVersion,
	)
}
