package handler

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/events"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/usecase"
	"go.uber.org/zap"
)

// Topic names for the stats pipeline (pattern stats.pos.<domain>.event).
const (
	TopicOrder       = "stats.pos.order.event"
	TopicOrderItem   = "stats.pos.order_item.event"
	TopicTransaction = "stats.pos.transaction.event"
	TopicCashier     = "stats.pos.cashier.event"
)

// StatsTopics returns every topic stats-writer consumes.
func StatsTopics() []string {
	return []string{TopicOrder, TopicOrderItem, TopicTransaction, TopicCashier}
}

// statEnvelope mirrors the envelope producers publish for dedup.
type statEnvelope struct {
	EventID string          `json:"event_id"`
	Payload json.RawMessage `json:"payload"`
}

type StatsHandler struct {
	useCase usecase.UseCase
	log     logger.LoggerInterface

	// dedup + mu guard the 24h event_id window. ConsumeClaim runs one goroutine
	// per Kafka partition (and per topic), so the map must be mutex-protected;
	// without it a concurrent iteration+write can panic the whole consumer.
	mu    sync.Mutex
	dedup map[string]time.Time
}

func NewStatsHandler(useCase usecase.UseCase, log logger.LoggerInterface) *StatsHandler {
	return &StatsHandler{
		useCase: useCase,
		log:     log,
		dedup:   make(map[string]time.Time),
	}
}

func (h *StatsHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *StatsHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return h.useCase.Close()
}

func (h *StatsHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// Unwrap the envelope and check dedup before processing. At-least-once
		// delivery can redeliver the same message after a rebalance; the
		// event_id window guards against double materialization.
		raw := msg.Value
		eventID := ""
		if env := h.tryUnwrap(raw); env != nil {
			eventID = env.EventID
			if h.isDuplicate(eventID) {
				session.MarkMessage(msg, "")
				continue
			}
			raw = env.Payload
		}

		switch msg.Topic {
		case TopicOrder:
			var event events.OrderEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal order event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveOrderEvent(session.Context(), eventID, event); err != nil {
				h.log.Error("Failed to save order event", zap.Error(err))
				continue
			}
		case TopicOrderItem:
			var event events.OrderItemEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal order item event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveOrderItemEvent(session.Context(), eventID, event); err != nil {
				h.log.Error("Failed to save order item event", zap.Error(err))
				continue
			}
		case TopicTransaction:
			var event events.TransactionEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal transaction event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveTransactionEvent(session.Context(), eventID, event); err != nil {
				h.log.Error("Failed to save transaction event", zap.Error(err))
				continue
			}
		case TopicCashier:
			var event events.CashierEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal cashier event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveCashierEvent(session.Context(), eventID, event); err != nil {
				h.log.Error("Failed to save cashier event", zap.Error(err))
				continue
			}
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *StatsHandler) tryUnwrap(raw []byte) *statEnvelope {
	var env statEnvelope
	if err := json.Unmarshal(raw, &env); err != nil || env.EventID == "" {
		return nil
	}
	return &env
}

// isDuplicate reports whether eventID was seen in the last 24h window and
// records it. The map is bounded by the window sweep on each call.
func (h *StatsHandler) isDuplicate(eventID string) bool {
	if eventID == "" {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)
	if seen, ok := h.dedup[eventID]; ok && seen.After(cutoff) {
		return true
	}
	for id, seen := range h.dedup {
		if seen.Before(cutoff) {
			delete(h.dedup, id)
		}
	}
	h.dedup[eventID] = now
	return false
}
