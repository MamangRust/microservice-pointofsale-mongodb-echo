package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestDependencyGuard_TransportFailures_OpenCircuitAndFailFast verifies that
// repeated transport failures open the circuit breaker and subsequent calls
// fail fast WITHOUT invoking the dependency (§8.1 poin 5).
func TestDependencyGuard_TransportFailures_OpenCircuitAndFailFast(t *testing.T) {
	guard := NewDependencyGuard("test", 5, 30, 100, 3*time.Second, nil)

	var calls int32
	fn := func(ctx context.Context) error {
		atomic.AddInt32(&calls, 1)
		return status.Error(codes.Unavailable, "down")
	}

	for i := 0; i < 5; i++ {
		err := guard.Call(context.Background(), fn)
		if status.Code(err) != codes.Unavailable {
			t.Fatalf("call %d: expected Unavailable, got %v", i+1, err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 5 {
		t.Fatalf("expected 5 dependency calls before circuit opens, got %d", got)
	}
	if !guard.breaker.IsOpen() {
		t.Fatal("circuit breaker should be open after 5 transport failures")
	}

	// The 6th call must fail fast without reaching the dependency.
	err := guard.Call(context.Background(), fn)
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("expected fail-fast Unavailable, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 5 {
		t.Fatalf("fail-fast must not call the dependency, calls=%d", got)
	}
}

// TestDependencyGuard_BusinessErrors_DoNotOpenCircuit verifies that business
// errors (not found, validation) are normal responses and never trip the
// breaker.
func TestDependencyGuard_BusinessErrors_DoNotOpenCircuit(t *testing.T) {
	guard := NewDependencyGuard("test", 2, 30, 100, 3*time.Second, nil)

	fn := func(ctx context.Context) error {
		return status.Error(codes.NotFound, "no such entity")
	}

	for i := 0; i < 10; i++ {
		if err := guard.Call(context.Background(), fn); err == nil {
			t.Fatal("expected business error")
		}
	}
	if guard.breaker.IsOpen() {
		t.Fatal("business errors must not open the circuit breaker")
	}
	if got := guard.breaker.GetFailureCount(); got != 0 {
		t.Fatalf("failure count must stay 0 for business errors, got %d", got)
	}
}

// TestDependencyGuard_CallTimeout verifies the per-call deadline: a slow
// dependency is cut off after callTimeout with DeadlineExceeded.
func TestDependencyGuard_CallTimeout(t *testing.T) {
	guard := NewDependencyGuard("test", 100, 30, 100, 50*time.Millisecond, nil)

	var sawDeadline bool
	err := guard.Call(context.Background(), func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			sawDeadline = true
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	})
	if !sawDeadline {
		t.Fatal("dependency context should carry the call deadline")
	}
	// The guard passes the dependency error through unchanged; a plain
	// context.DeadlineExceeded (non-gRPC) surfaces as-is.
	if !errors.Is(err, context.DeadlineExceeded) && status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

// TestDependencyGuard_BulkheadFull verifies the semaphore bulkhead rejects
// calls beyond maxConcurrent with ResourceExhausted.
func TestDependencyGuard_BulkheadFull(t *testing.T) {
	guard := NewDependencyGuard("test", 100, 30, 1, time.Second, nil)

	release := make(chan struct{})
	defer close(release)

	go func() {
		_ = guard.Call(context.Background(), func(ctx context.Context) error {
			<-release
			return nil
		})
	}()

	// Give the goroutine a moment to acquire the single permit.
	time.Sleep(50 * time.Millisecond)

	err := guard.Call(context.Background(), func(ctx context.Context) error { return nil })
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted (bulkhead full), got %v", err)
	}
}

// TestDependencyGuard_NilGuard_Passthrough verifies a nil guard is a no-op.
func TestDependencyGuard_NilGuard_Passthrough(t *testing.T) {
	var guard *DependencyGuard
	called := false
	err := guard.Call(context.Background(), func(ctx context.Context) error {
		called = true
		return errors.New("passthrough")
	})
	if !called {
		t.Fatal("nil guard must invoke fn")
	}
	if err == nil || err.Error() != "passthrough" {
		t.Fatalf("nil guard must pass the error through unchanged, got %v", err)
	}
}
