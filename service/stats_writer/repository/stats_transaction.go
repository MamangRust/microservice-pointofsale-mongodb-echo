package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
)

// ── Transaction stats ───────────────────────────────────────────────────────
//
// TransactionStatsRepository writes the events that feed the transaction
// statistics (transaction_daily), mirroring the reader's
// TransactionStatsRepository.
type TransactionStatsRepository interface {
	InsertTransactionEvent(ctx context.Context, eventID string, eventVersion uint64, event events.TransactionEvent) error
}
