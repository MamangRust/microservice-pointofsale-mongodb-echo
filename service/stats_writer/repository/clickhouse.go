package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	chDriver "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultBatchSize     = 1000
	defaultFlushInterval = 5 * time.Second
)

// clickhouseBatch is a local interface matching the methods we use from
// driver.Batch so the batch path stays unit-testable.
type clickhouseBatch interface {
	Append(v ...interface{}) error
	Send() error
}

type batchEntry struct {
	batch clickhouseBatch
	count int
}

// clickhouseRepository holds the shared batching machinery used by every
// domain implementation (clickhouse_order.go, clickhouse_cashier.go,
// clickhouse_category.go, clickhouse_transaction.go). Only the Insert* methods
// live in those files.
type clickhouseRepository struct {
	conn chDriver.Conn
	log  logger.LoggerInterface

	mu        sync.Mutex
	batches   map[string]*batchEntry
	batchSize int

	flushTicker *time.Ticker
	flushDone   chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewClickhouseRepository(conn chDriver.Conn, log logger.LoggerInterface) Repository {
	ctx, cancel := context.WithCancel(context.Background())

	r := &clickhouseRepository{
		conn:        conn,
		log:         log,
		batches:     make(map[string]*batchEntry),
		batchSize:   defaultBatchSize,
		flushTicker: time.NewTicker(defaultFlushInterval),
		flushDone:   make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	go r.flushLoop()

	return r
}

func (r *clickhouseRepository) Flush(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for key, entry := range r.batches {
		if entry.count == 0 {
			continue
		}
		if err := entry.batch.Send(); err != nil {
			r.log.Error("Failed to flush ClickHouse batch",
				zap.String("batch", key),
				zap.Int("rows", entry.count),
				zap.Error(err),
			)
			lastErr = fmt.Errorf("flush batch %s: %w", key, err)
		} else {
			r.log.Debug("Flushed ClickHouse batch",
				zap.String("batch", key),
				zap.Int("rows", entry.count),
			)
		}
		delete(r.batches, key)
	}
	return lastErr
}

func (r *clickhouseRepository) Close() error {
	r.flushTicker.Stop()
	r.cancel()
	<-r.flushDone

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return r.Flush(ctx)
}

func (r *clickhouseRepository) appendToBatch(ctx context.Context, key, query string, args ...interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.batches[key]
	if !exists {
		batch, err := r.conn.PrepareBatch(ctx, query)
		if err != nil {
			r.log.Error("Failed to prepare ClickHouse batch",
				zap.String("batch", key),
				zap.Error(err),
			)
			return fmt.Errorf("prepare batch %s: %w", key, err)
		}
		entry = &batchEntry{batch: batch}
		r.batches[key] = entry
	}

	if err := entry.batch.Append(args...); err != nil {
		r.log.Error("Failed to append to ClickHouse batch",
			zap.String("batch", key),
			zap.Error(err),
		)
		return fmt.Errorf("append to batch %s: %w", key, err)
	}

	entry.count++
	if entry.count >= r.batchSize {
		if err := entry.batch.Send(); err != nil {
			r.log.Error("Failed to send full ClickHouse batch",
				zap.String("batch", key),
				zap.Int("rows", entry.count),
				zap.Error(err),
			)
			delete(r.batches, key)
			return fmt.Errorf("send batch %s: %w", key, err)
		}
		r.log.Debug("Sent full ClickHouse batch",
			zap.String("batch", key),
			zap.Int("rows", entry.count),
		)
		delete(r.batches, key)
	}

	return nil
}

func (r *clickhouseRepository) flushLoop() {
	defer close(r.flushDone)

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.flushTicker.C:
			ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
			_ = r.Flush(ctx)
			cancel()
		}
	}
}

// toUUID parses a string UUID; an empty or malformed value yields the nil UUID
// so a producer that omits event_id still inserts a valid ClickHouse UUID. The
// return type is uuid.UUID (not [16]byte) because clickhouse-go's batch
// Append only accepts the named type.
func toUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// parseEventTime parses an RFC3339 event time; empty or malformed values fall
// back to now() so ClickHouse always receives a valid DateTime.
func parseEventTime(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now()
	}
	return t
}
