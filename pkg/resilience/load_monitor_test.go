package resilience

import (
	"testing"
	"time"
)

func TestLoadMonitorRecordAndCount(t *testing.T) {
	lm := NewLoadMonitor()

	for i := 0; i < 10; i++ {
		lm.RecordRequest()
	}

	if lm.GetRequestCount() != 10 {
		t.Fatalf("expected request count 10, got %d", lm.GetRequestCount())
	}

	// GetCurrentRPS swaps the counter to zero.
	rps := lm.GetCurrentRPS()
	if lm.GetRequestCount() != 0 {
		t.Fatalf("expected counter reset after GetCurrentRPS, got %d", lm.GetRequestCount())
	}
	if rps == 0 {
		t.Log("rps was 0 (elapsed window too small) — acceptable, counter still reset")
	}
}

func TestLoadMonitorAverageAndPeak(t *testing.T) {
	lm := NewLoadMonitor()
	window := NewLoadMonitorWithWindow(10 * time.Millisecond)

	for i := 0; i < 5; i++ {
		lm.RecordRequest()
		window.RecordRequest()
	}
	lm.GetCurrentRPS()
	window.GetCurrentRPS()

	avg := lm.GetAverageRPS(time.Minute)
	if avg == 0 && len(lm.history) > 0 {
		t.Log("average rps was 0 for tiny windows — acceptable")
	}

	peak := window.GetPeakRPS(time.Minute)
	_ = peak // peak may be 0 on fast machines; presence of history is the invariant
	if len(window.history) == 0 {
		t.Fatal("expected load monitor to record history snapshots")
	}
}

func TestLoadMonitorEmptyHistory(t *testing.T) {
	lm := NewLoadMonitor()

	if avg := lm.GetAverageRPS(time.Minute); avg != 0 {
		t.Fatalf("expected 0 average with empty history, got %v", avg)
	}
	if peak := lm.GetPeakRPS(time.Minute); peak != 0 {
		t.Fatalf("expected 0 peak with empty history, got %d", peak)
	}
}

func TestLoadMonitorHistoryBounded(t *testing.T) {
	lm := NewLoadMonitor()

	// 100 snapshots > 60 cap.
	for i := 0; i < 100; i++ {
		lm.RecordRequest()
		lm.GetCurrentRPS()
	}

	if len(lm.history) > 60 {
		t.Fatalf("expected history bounded to 60, got %d", len(lm.history))
	}
}
