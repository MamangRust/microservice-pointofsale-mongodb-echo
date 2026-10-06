package outbox

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrInvalidInboxKey = errors.New("invalid consumer inbox key")

// ConsumerInbox is the durable deduplication contract used by Kafka handlers.
type ConsumerInbox interface {
	Reserve(ctx context.Context, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error)
	MarkProcessed(ctx context.Context, consumerName, eventKey string, reservationVersion int64) error
	Release(ctx context.Context, consumerName, eventKey string, reservationVersion int64, processingErr error) error
}

// MongoInbox adapts the consumer-inbox pattern to MongoDB.
type MongoInbox struct {
	db *mongo.Database
}

func NewMongoInbox(db *mongo.Database) (*MongoInbox, error) {
	if db == nil {
		return nil, errors.New("consumer inbox requires a non-nil MongoDB database")
	}
	return &MongoInbox{db: db}, nil
}

type consumerInboxDoc struct {
	ConsumerName      string    `bson:"consumer_name"`
	EventKey          string    `bson:"event_key"`
	Topic             string    `bson:"topic"`
	PartitionID       int32     `bson:"partition_id"`
	MessageOffset     int64     `bson:"message_offset"`
	Status            string    `bson:"status"`
	Attempts          int32     `bson:"attempts"`
	ReservationVersion int64    `bson:"reservation_version"`
	LeaseUntil        time.Time `bson:"lease_until"`
	LastError         string    `bson:"last_error"`
	ProcessedAt       *time.Time `bson:"processed_at,omitempty"`
}

func (m *MongoInbox) Reserve(ctx context.Context, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error) {
	if consumerName == "" || eventKey == "" {
		return false, false, 0, ErrInvalidInboxKey
	}

	coll := m.db.Collection("consumer_inbox")
	filter := bson.M{
		"consumer_name": consumerName,
		"event_key":     eventKey,
	}

	// Try to find existing document
	var existing consumerInboxDoc
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil && err != mongo.ErrNoDocuments {
		return false, false, 0, err
	}

	now := time.Now().UTC()
	leaseUntil := now.Add(1 * time.Minute)

	if err == mongo.ErrNoDocuments {
		// New event — insert
		doc := consumerInboxDoc{
			ConsumerName:      consumerName,
			EventKey:          eventKey,
			Topic:             topic,
			PartitionID:       partition,
			MessageOffset:     offset,
			Status:            "processing",
			Attempts:          1,
			ReservationVersion: 1,
			LeaseUntil:        leaseUntil,
			LastError:         "",
		}
		if _, err := coll.InsertOne(ctx, doc); err != nil {
			return false, false, 0, err
		}
		return true, false, 1, nil
	}

	// Already processed — nothing to do
	if existing.Status == "processed" {
		return false, true, existing.ReservationVersion, nil
	}

	// Not processed — try to re-reserve if lease expired
	if existing.LeaseUntil.Before(now) || existing.Status == "pending" {
		update := bson.M{
			"$set": bson.M{
				"status":              "processing",
				"lease_until":         leaseUntil,
				"last_error":          "",
				"topic":               topic,
				"partition_id":        partition,
				"message_offset":      offset,
			},
			"$inc": bson.M{
				"attempts":           int32(1),
				"reservation_version": int64(1),
			},
		}
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		var updated consumerInboxDoc
		if err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated); err != nil {
			return false, false, 0, err
		}
		return true, false, updated.ReservationVersion, nil
	}

	// Still processing, not expired
	return false, false, 0, nil
}

func (m *MongoInbox) MarkProcessed(ctx context.Context, consumerName, eventKey string, reservationVersion int64) error {
	if consumerName == "" || eventKey == "" {
		return ErrInvalidInboxKey
	}

	coll := m.db.Collection("consumer_inbox")
	filter := bson.M{
		"consumer_name":      consumerName,
		"event_key":          eventKey,
		"status":             "processing",
		"reservation_version": reservationVersion,
	}
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":      "processed",
			"processed_at": now,
			"lease_until": now,
			"last_error":  "",
		},
	}
	_, err := coll.UpdateOne(ctx, filter, update)
	return err
}

func (m *MongoInbox) Release(ctx context.Context, consumerName, eventKey string, reservationVersion int64, processingErr error) error {
	if consumerName == "" || eventKey == "" {
		return ErrInvalidInboxKey
	}

	lastError := "consumer processing failed"
	if processingErr != nil {
		lastError = processingErr.Error()
	}

	coll := m.db.Collection("consumer_inbox")
	filter := bson.M{
		"consumer_name":      consumerName,
		"event_key":          eventKey,
		"status":             "processing",
		"reservation_version": reservationVersion,
	}
	update := bson.M{
		"$set": bson.M{
			"status":      "pending",
			"lease_until": time.Now().UTC(),
			"last_error":  lastError,
		},
	}
	_, err := coll.UpdateOne(ctx, filter, update)
	return err
}
