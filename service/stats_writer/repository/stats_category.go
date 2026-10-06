package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Category stats ──────────────────────────────────────────────────────────
//
// CategoryStatsRepository writes the events that feed the category statistics
// (order_item_daily), mirroring the reader's CategoryStatsRepository. The
// category_id is denormalized onto each row by the producer/backfill so the
// reader never needs a cross-service join.
type CategoryStatsRepository interface {
	InsertOrderItemEvent(ctx context.Context, eventID string, eventVersion uint64, event events.OrderItemEvent) error
}
