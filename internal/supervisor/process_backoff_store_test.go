package supervisor

import (
	"testing"
	"time"
)

func TestProcessBackoffStore_InitialAttempt(t *testing.T) {
	store := NewProcessBackoffStore(DefaultBackoffPolicy())
	delay := store.NextDelay("web")
	if delay < 0 {
		t.Fatalf("expected non-negative delay, got %v", delay)
	}
}

func TestProcessBackoffStore_DelayGrowsWithAttempts(t *testing.T) {
	policy := DefaultBackoffPolicy()
	policy.Kind = "exponential"
	policy.BaseDelay = 100 * time.Millisecond
	policy.Multiplier = 2.0
	policy.MaxDelay = 10 * time.Second

	store := NewProcessBackoffStore(policy)

	d1 := store.NextDelay("svc")
	d2 := store.NextDelay("svc")
	d3 := store.NextDelay("svc")

	if d2 <= d1 {
		t.Errorf("expected d2 > d1, got d1=%v d2=%v", d1, d2)
	}
	if d3 <= d2 {
		t.Errorf("expected d3 > d2, got d2=%v d3=%v", d2, d3)
	}
}

func TestProcessBackoffStore_ResetClearsAttempts(t *testing.T) {
	policy := DefaultBackoffPolicy()
	policy.Kind = "exponential"
	policy.BaseDelay = 100 * time.Millisecond
	policy.Multiplier = 2.0
	policy.MaxDelay = 10 * time.Second

	store := NewProcessBackoffStore(policy)

	// Advance a few attempts
	store.NextDelay("svc")
	store.NextDelay("svc")
	dBefore := store.NextDelay("svc")

	store.Reset("svc")
	dAfter := store.NextDelay("svc")

	if dAfter >= dBefore {
		t.Errorf("expected delay after reset to be less than before reset: before=%v after=%v", dBefore, dAfter)
	}
}

func TestProcessBackoffStore_IndependentPerProcess(t *testing.T) {
	policy := DefaultBackoffPolicy()
	policy.Kind = "exponential"
	policy.BaseDelay = 50 * time.Millisecond
	policy.Multiplier = 2.0
	policy.MaxDelay = 5 * time.Second

	store := NewProcessBackoffStore(policy)

	// Advance "a" several times
	store.NextDelay("a")
	store.NextDelay("a")
	store.NextDelay("a")
	dA := store.NextDelay("a")

	// "b" should start fresh
	dB := store.NextDelay("b")

	if dA <= dB {
		t.Errorf("expected process 'a' to have higher delay than fresh 'b': dA=%v dB=%v", dA, dB)
	}
}

func TestProcessBackoffStore_AttemptCount(t *testing.T) {
	store := NewProcessBackoffStore(DefaultBackoffPolicy())

	if store.Attempts("x") != 0 {
		t.Fatalf("expected 0 attempts initially, got %d", store.Attempts("x"))
	}

	store.NextDelay("x")
	store.NextDelay("x")

	if store.Attempts("x") != 2 {
		t.Fatalf("expected 2 attempts, got %d", store.Attempts("x"))
	}
}

func TestProcessBackoffStore_ResetUnknownProcess(t *testing.T) {
	store := NewProcessBackoffStore(DefaultBackoffPolicy())
	// Should not panic when resetting a process that was never tracked
	store.Reset("nonexistent")
	if store.Attempts("nonexistent") != 0 {
		t.Fatalf("expected 0 attempts after reset of unknown process")
	}
}

func TestProcessBackoffStore_MaxDelayRespected(t *testing.T) {
	policy := DefaultBackoffPolicy()
	policy.Kind = "exponential"
	policy.BaseDelay = 100 * time.Millisecond
	policy.Multiplier = 10.0
	policy.MaxDelay = 500 * time.Millisecond

	store := NewProcessBackoffStore(policy)

	var maxSeen time.Duration
	for i := 0; i < 10; i++ {
		d := store.NextDelay("proc")
		if d > maxSeen {
			maxSeen = d
		}
	}

	if maxSeen > policy.MaxDelay {
		t.Errorf("delay exceeded MaxDelay: got %v, max allowed %v", maxSeen, policy.MaxDelay)
	}
}
