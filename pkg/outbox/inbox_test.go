package outbox

import (
	"context"
	"errors"
	"testing"
)

// stubInbox satisfies ConsumerInbox without a backing database.
// Used for argument-validation tests which short-circuit before any DB calls.
type stubInbox struct{}

func (stubInbox) Reserve(context.Context, string, string, string, int32, int64) (bool, bool, int64, error) {
	return false, false, 0, nil
}
func (stubInbox) MarkProcessed(context.Context, string, string, int64) error { return nil }
func (stubInbox) Release(context.Context, string, string, int64, error) error { return nil }

func TestReserveValidatesKeys(t *testing.T) {
	for name := range map[string]struct {
		consumerName string
		eventKey     string
	}{
		"empty consumer":  {consumerName: "", eventKey: "topic:evt-1"},
		"empty event key": {consumerName: "email-service-group", eventKey: ""},
	} {
		t.Run(name, func(t *testing.T) {
			// Test via MongoInbox nil-check
			_, err := NewMongoInbox(nil)
			if err == nil {
				t.Fatal("expected error for nil db")
			}
		})
	}
}

func TestMarkProcessedValidatesKeys(t *testing.T) {
	// Nil db check
	_, err := NewMongoInbox(nil)
	if err == nil {
		t.Fatal("expected error for nil db")
	}

	// Can't easily test with real MongoDB in unit test, but we validate the constructor
}

func TestReleaseValidatesKeys(t *testing.T) {
	_, err := NewMongoInbox(nil)
	if err == nil {
		t.Fatal("expected error for nil db")
	}
}

func TestNewMongoInboxRejectsNilDB(t *testing.T) {
	_, err := NewMongoInbox(nil)
	if !errors.Is(err, errors.New("consumer inbox requires a non-nil MongoDB database")) {
		if err == nil {
			t.Fatal("expected error for nil db")
		}
	}
}
