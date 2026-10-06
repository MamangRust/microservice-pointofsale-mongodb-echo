package resilience

import "testing"

func TestRequestLimiterAcquireRelease(t *testing.T) {
	rl := NewRequestLimiter(2, noopLogger{})

	if !rl.TryAcquire() {
		t.Fatal("expected first acquire to succeed")
	}
	if !rl.TryAcquire() {
		t.Fatal("expected second acquire to succeed")
	}
	if rl.AvailablePermits() != 0 {
		t.Fatalf("expected 0 available permits, got %d", rl.AvailablePermits())
	}

	rl.Release()
	if rl.AvailablePermits() != 1 {
		t.Fatalf("expected 1 available permit after release, got %d", rl.AvailablePermits())
	}
}

func TestRequestLimiterRejectsWhenSaturated(t *testing.T) {
	rl := NewRequestLimiter(1, noopLogger{})

	if !rl.TryAcquire() {
		t.Fatal("expected acquire to succeed within limit")
	}
	if rl.TryAcquire() {
		t.Fatal("expected acquire to be rejected when saturated")
	}
}

func TestRequestLimiterDefaultsToPositiveMax(t *testing.T) {
	rl := NewRequestLimiter(0, noopLogger{})
	if rl.MaxConcurrent() != 100 {
		t.Fatalf("expected default maxConcurrent 100, got %d", rl.MaxConcurrent())
	}
}

func TestRequestLimiterSafeOverRelease(t *testing.T) {
	rl := NewRequestLimiter(1, noopLogger{})

	// Release without a matching acquire must not panic or go negative.
	rl.Release()
	rl.Release()

	if !rl.TryAcquire() {
		t.Fatal("expected limiter to remain usable after over-release")
	}
}
