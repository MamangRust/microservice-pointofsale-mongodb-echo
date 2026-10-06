package repository

import "context"

// Repository materializes stats events into ClickHouse. Batches are flushed on
// size and on an interval so a burst of events does not hammer ClickHouse with
// single-row inserts.
//
// The write surface is grouped by the same four stats domains the reader serves
// (service/stats_reader/repository) — order, cashier, category, transaction —
// so both sides of the pipeline stay in sync. Repository is the full surface
// implemented by clickhouseRepository.
type Repository interface {
	OrderStatsRepository
	CashierStatsRepository
	CategoryStatsRepository
	TransactionStatsRepository

	// Flush sends every pending batch; Close stops the flush loop and flushes
	// one last time.
	Flush(ctx context.Context) error
	Close() error
}
