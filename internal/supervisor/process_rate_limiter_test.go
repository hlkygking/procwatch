package supervisor

import (
	"testing"
	"time"
)

func TestProcessRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := NewProcessRateLimiter(time.Second, 3)

	for i := 0; i < 3; i++ {
		if !rl.Allow("svc") {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}
}

func TestProcessRateLimiter_DeniesWhenLimitReached(t *testing.T) {
	rl := NewProcessRateLimiter(time.Second, 2)

	rl.Allow("svc")
	rl.Allow("svc")

	if rl.Allow("svc") {
		t.Fatal("expected Allow to return false after limit reached")
	}
}

func TestProcessRateLimiter_Count(t *testing.T) {
	rl := NewProcessRateLimiter(time.Second, 10)

	rl.Allow("svc")
	rl.Allow("svc")
	rl.Allow("svc")

	if got := rl.Count("svc"); got != 3 {
		t.Fatalf("expected count 3, got %d", got)
	}
}

func TestProcessRateLimiter_Reset(t *testing.T) {
	rl := NewProcessRateLimiter(time.Second, 2)

	rl.Allow("svc")
	rl.Allow("svc")
	rl.Reset("svc")

	if got := rl.Count("svc"); got != 0 {
		t.Fatalf("expected count 0 after reset, got %d", got)
	}
	if !rl.Allow("svc") {
		t.Fatal("expected Allow to return true after reset")
	}
}

func TestProcessRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewProcessRateLimiter(50*time.Millisecond, 1)

	rl.Allow("svc") // fills the window

	if rl.Allow("svc") {
		t.Fatal("expected Allow to deny before window expires")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("svc") {
		t.Fatal("expected Allow to permit after window expires")
	}
}

func TestProcessRateLimiter_IndependentProcesses(t *testing.T) {
	rl := NewProcessRateLimiter(time.Second, 1)

	if !rl.Allow("alpha") {
		t.Fatal("expected alpha to be allowed")
	}
	if !rl.Allow("beta") {
		t.Fatal("expected beta to be allowed independently")
	}
	if rl.Allow("alpha") {
		t.Fatal("expected alpha to be denied after limit")
	}
}

func TestProcessRateLimiter_CountUnknownProcess(t *testing.T) {
	rl := NewProcessRateLimiter(time.Second, 5)

	if got := rl.Count("unknown"); got != 0 {
		t.Fatalf("expected 0 for unknown process, got %d", got)
	}
}
