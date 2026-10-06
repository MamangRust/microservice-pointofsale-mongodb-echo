package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Order stats ─────────────────────────────────────────────────────────────
//
// OrderStatsRepository writes the events that feed the order statistics
// (order_daily), mirroring the reader's OrderStatsRepository.
//
// eventID is the idempotency key stored as part of the ClickHouse primary key;
// eventVersion feeds the ReplacingMergeTree version column so a newer delivery
// (or a re-backfill with a newer run timestamp) supersedes older rows.
type OrderStatsRepository interface {
	InsertOrderEvent(ctx context.Context, eventID string, eventVersion uint64, event events.OrderEvent) error
}
