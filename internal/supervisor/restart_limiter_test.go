package supervisor

import (
	"testing"
	"time"
)

func TestRestartLimiter_AllowUnlimited(t *testing.T) {
	rl := NewRestartLimiter(0, time.Minute, time.Second, 30*time.Second)
	for i := 0; i < 100; i++ {
		if !rl.Allow() {
			t.Fatal("expected unlimited restarts to always be allowed")
		}
		rl.Record()
	}
}

func TestRestartLimiter_MaxRestartsEnforced(t *testing.T) {
	rl := NewRestartLimiter(3, time.Minute, time.Second, 30*time.Second)
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Fatalf("expected restart %d to be allowed", i+1)
		}
		rl.Record()
	}
	if rl.Allow() {
		t.Fatal("expected 4th restart to be denied")
	}
}

func TestRestartLimiter_Reset(t *testing.T) {
	rl := NewRestartLimiter(2, time.Minute, time.Second, 30*time.Second)
	rl.Record()
	rl.Record()
	if rl.Allow() {
		t.Fatal("expected restart to be denied before reset")
	}
	rl.Reset()
	if !rl.Allow() {
		t.Fatal("expected restart to be allowed after reset")
	}
}

func TestRestartLimiter_WindowExpiry(t *testing.T) {
	rl := NewRestartLimiter(2, 50*time.Millisecond, time.Millisecond, 100*time.Millisecond)
	rl.Record()
	rl.Record()
	if rl.Allow() {
		t.Fatal("expected restart to be denied within window")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.Allow() {
		t.Fatal("expected restart to be allowed after window expired")
	}
}

func TestRestartLimiter_BackoffGrowth(t *testing.T) {
	rl := NewRestartLimiter(0, time.Minute, time.Second, 16*time.Second)
	// 0 attempts → base backoff
	b0 := rl.Backoff()
	if b0 != time.Second {
		t.Fatalf("expected 1s backoff, got %v", b0)
	}
	rl.Record()
	b1 := rl.Backoff()
	if b1 != 2*time.Second {
		t.Fatalf("expected 2s backoff after 1 attempt, got %v", b1)
	}
	rl.Record()
	rl.Record()
	rl.Record()
	// 4 attempts: 1 → 2 → 4 → 8 → 16 capped
	b4 := rl.Backoff()
	if b4 != 16*time.Second {
		t.Fatalf("expected 16s (capped) backoff, got %v", b4)
	}
}

func TestRestartLimiter_ZeroBackoff(t *testing.T) {
	rl := NewRestartLimiter(0, time.Minute, 0, 0)
	rl.Record()
	if d := rl.Backoff(); d != 0 {
		t.Fatalf("expected zero backoff, got %v", d)
	}
}
