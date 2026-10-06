package resilience

import (
	"testing"
	"time"
)

func TestCircuitBreakerInitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 60, noopLogger{})
	if cb.IsOpen() {
		t.Fatal("expected circuit breaker to start closed")
	}
	if !cb.ShouldAllowRequest() {
		t.Fatal("expected requests to be allowed while closed")
	}
}

func TestCircuitBreakerOpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 60, noopLogger{})

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.IsOpen() {
		t.Fatal("expected circuit to stay closed below threshold")
	}

	cb.RecordFailure()
	if !cb.IsOpen() {
		t.Fatal("expected circuit to open after reaching threshold")
	}
	if cb.ShouldAllowRequest() {
		t.Fatal("expected requests to be rejected while open")
	}
}

func TestCircuitBreakerHalfOpenRecoversAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 1, noopLogger{})
	cb.RecordFailure()
	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open")
	}

	// After the timeout window elapses, the breaker enters half-open state
	// and allows a limited number of probe requests.
	time.Sleep(1100 * time.Millisecond)

	if !cb.ShouldAllowRequest() {
		t.Fatal("expected probe request to be allowed in half-open state")
	}

	// Enough successful probes close the circuit again.
	for i := 0; i < 5; i++ {
		cb.RecordSuccess()
	}
	if cb.IsOpen() {
		t.Fatal("expected circuit to close after enough successes")
	}
}

func TestCircuitBreakerResetCountersOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 60, noopLogger{})

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	cb.RecordFailure()
	if cb.IsOpen() {
		t.Fatal("expected success to reset failure count below threshold")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := NewCircuitBreaker(1, 60, noopLogger{})
	cb.RecordFailure()
	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open")
	}

	cb.Reset()
	if cb.IsOpen() {
		t.Fatal("expected Reset to close the circuit")
	}
	if !cb.ShouldAllowRequest() {
		t.Fatal("expected requests allowed after reset")
	}
}
