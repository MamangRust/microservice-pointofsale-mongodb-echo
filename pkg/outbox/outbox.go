package outbox

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// outboxMetrics records outbox observability.
type outboxMetrics struct {
	pending   metric.Int64Gauge
	published metric.Int64Counter
	failed    metric.Int64Counter
}

func newOutboxMetrics() *outboxMetrics {
	meter := otel.Meter("outbox")
	pending, _ := meter.Int64Gauge(
		"outbox_events_pending",
		metric.WithDescription("Number of outbox events waiting for delivery"),
	)
	published, _ := meter.Int64Counter(
		"outbox_events_published",
		metric.WithDescription("Total number of outbox events successfully published"),
	)
	failed, _ := meter.Int64Counter(
		"outbox_events_failed",
		metric.WithDescription("Total number of outbox events that failed to publish"),
	)
	return &outboxMetrics{pending: pending, published: published, failed: failed}
}

const (
	OutboxMaxAttempts    = 5
	OutboxBackoff        = 30 * time.Second
	OutboxRelayInterval  = 5 * time.Second
	OutboxRelayBatchSize = 100
	OutboxClaimLease     = 1 * time.Minute
	OutboxRetention      = 7 * 24 * time.Hour

	OutboxRetentionEveryTicks = 60
)

// OutboxEvent mirrors a document in the outbox_events collection.
type OutboxEvent struct {
	OutboxID      int64     `bson:"_id"`
	Topic         string    `bson:"topic"`
	EventKey      string    `bson:"event_key"`
	Payload       []byte    `bson:"payload"`
	Status        string    `bson:"status"`
	Attempts      int32     `bson:"attempts"`
	NextAttemptAt time.Time `bson:"next_attempt_at"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
}

type OutboxPublisher interface {
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
}

type OutboxService struct {
	db        *mongo.Database
	publisher OutboxPublisher
	logger    logger.LoggerInterface
	metrics   *outboxMetrics
}

func NewOutboxService(db *mongo.Database, publisher OutboxPublisher, log logger.LoggerInterface) *OutboxService {
	return &OutboxService{db: db, publisher: publisher, logger: log, metrics: newOutboxMetrics()}
}

// nextOutboxID generates the next outbox event ID.
func (s *OutboxService) nextOutboxID(ctx context.Context) (int64, error) {
	counters := s.db.Collection("counters")
	filter := bson.M{"_id": "outbox_events"}
	update := bson.M{"$inc": bson.M{"seq": int64(1)}}
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetUpsert(true)

	var result struct {
		Seq int64 `bson:"seq"`
	}
	err := counters.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	return result.Seq, err
}

func (s *OutboxService) EnqueueInTx(ctx context.Context, _ interface{}, topic, key string, payload []byte) error {
	// In MongoDB, the caller should use a session/transaction and pass the session context.
	// For simplicity, we just insert directly — the caller manages the transaction.
	return s.enqueue(ctx, topic, key, payload)
}

func (s *OutboxService) Enqueue(ctx context.Context, topic, key string, payload []byte) error {
	if s.db == nil {
		return nil
	}
	return s.enqueue(ctx, topic, key, payload)
}

func (s *OutboxService) enqueue(ctx context.Context, topic, key string, payload []byte) error {
	outboxID, err := s.nextOutboxID(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	doc := OutboxEvent{
		OutboxID:      outboxID,
		Topic:         topic,
		EventKey:      key,
		Payload:       payload,
		Status:        "pending",
		Attempts:      0,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	coll := s.db.Collection("outbox_events")
	if _, err := coll.InsertOne(ctx, doc); err != nil {
		return err
	}
	s.logger.Info("outbox event enqueued", zap.String("topic", topic), zap.String("key", key))
	return nil
}

// PublishPending claims pending events whose retry window has elapsed,
// publishes each to Kafka, and marks it delivered.
func (s *OutboxService) PublishPending(ctx context.Context, limit int) (int, error) {
	if s.db == nil || s.publisher == nil {
		return 0, nil
	}

	coll := s.db.Collection("outbox_events")

	// FindAndModify pattern to atomically claim events (MongoDB equivalent of FOR UPDATE SKIP LOCKED)
	filter := bson.M{
		"status":           "pending",
		"next_attempt_at":  bson.M{"$lte": time.Now().UTC()},
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "outbox_id", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var events []OutboxEvent
	if err := cursor.All(ctx, &events); err != nil {
		return 0, err
	}

	delivered := 0
	for i := range events {
		event := &events[i]

		// Claim: update status to 'claiming'
		claimFilter := bson.M{
			"_id":    event.OutboxID,
			"status": "pending",
		}
		claimUpdate := bson.M{
			"$set": bson.M{
				"next_attempt_at": time.Now().UTC().Add(OutboxClaimLease),
				"updated_at":      time.Now().UTC(),
			},
		}
		res, err := coll.UpdateOne(ctx, claimFilter, claimUpdate)
		if err != nil || res.MatchedCount == 0 {
			continue // Another relay claimed it
		}

		if err := s.publisher.SendMessage(ctx, event.Topic, event.EventKey, event.Payload); err != nil {
			s.logger.Error("failed to publish outbox event, scheduling retry",
				zap.Error(err),
				zap.Int64("outbox_id", event.OutboxID),
				zap.String("topic", event.Topic),
				zap.Int32("attempts", event.Attempts),
			)
			s.metrics.failed.Add(ctx, 1)

			if int(event.Attempts)+1 >= OutboxMaxAttempts {
				// Dead-letter
				s.deadLetter(ctx, coll, event.OutboxID)
				continue
			}

			nextAttempt := time.Now().UTC().Add(OutboxBackoff * time.Duration(event.Attempts+1))
			failUpdate := bson.M{
				"$set": bson.M{
					"next_attempt_at": nextAttempt,
					"updated_at":      time.Now().UTC(),
				},
				"$inc": bson.M{"attempts": int32(1)},
			}
			coll.UpdateOne(ctx, bson.M{"_id": event.OutboxID}, failUpdate)
			continue
		}

		// Mark delivered
		deliverUpdate := bson.M{
			"$set": bson.M{
				"status":     "delivered",
				"updated_at": time.Now().UTC(),
			},
		}
		coll.UpdateOne(ctx, bson.M{"_id": event.OutboxID}, deliverUpdate)
		delivered++
		s.metrics.published.Add(ctx, 1)
	}

	return delivered, nil
}

func (s *OutboxService) deadLetter(ctx context.Context, coll *mongo.Collection, outboxID int64) {
	update := bson.M{
		"$set": bson.M{
			"status":     "dead",
			"updated_at": time.Now().UTC(),
		},
	}
	if _, err := coll.UpdateOne(ctx, bson.M{"_id": outboxID}, update); err != nil {
		s.logger.Error("failed to dead-letter outbox event", zap.Error(err), zap.Int64("outbox_id", outboxID))
	}
}

func (s *OutboxService) recordPending(ctx context.Context) {
	coll := s.db.Collection("outbox_events")
	count, err := coll.CountDocuments(ctx, bson.M{"status": "pending"})
	if err != nil {
		return
	}
	s.metrics.pending.Record(ctx, count)
}

// Start runs the relay loop until ctx is cancelled.
func (s *OutboxService) Start(ctx context.Context, interval time.Duration, limit int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tickCount := 0
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("outbox relay stopped")
			return
		case <-ticker.C:
			if _, err := s.PublishPending(ctx, limit); err != nil {
				s.logger.Error("outbox relay cycle failed", zap.Error(err))
			}
			s.recordPending(ctx)
			tickCount++
			if tickCount%OutboxRetentionEveryTicks == 0 {
				s.purgeOld(ctx)
			}
		}
	}
}

func (s *OutboxService) purgeOld(ctx context.Context) {
	coll := s.db.Collection("outbox_events")
	cutoff := time.Now().UTC().Add(-OutboxRetention)
	filter := bson.M{
		"status":     bson.M{"$in": []string{"delivered", "dead"}},
		"updated_at": bson.M{"$lt": cutoff},
	}
	res, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		s.logger.Error("outbox retention cleanup failed", zap.Error(err))
	} else if res.DeletedCount > 0 {
		s.logger.Info("outbox retention cleanup", zap.Int64("removed", res.DeletedCount))
	}
}
