package supervisor

import (
	"testing"
	"time"
)

func TestThrottle_InitialDelay(t *testing.T) {
	policy := ThrottlePolicy{
		MinDelay: 100 * time.Millisecond,
		MaxDelay: 10 * time.Second,
		Factor:   2.0,
	}
	th := NewThrottle(policy)
	if th.Current() != 100*time.Millisecond {
		t.Errorf("expected initial delay 100ms, got %v", th.Current())
	}
}

func TestThrottle_ExponentialGrowth(t *testing.T) {
	policy := ThrottlePolicy{
		MinDelay: 100 * time.Millisecond,
		MaxDelay: 10 * time.Second,
		Factor:   2.0,
	}
	th := NewThrottle(policy)

	d1 := th.Next()
	d2 := th.Next()
	d3 := th.Next()

	if d1 != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", d1)
	}
	if d2 != 200*time.Millisecond {
		t.Errorf("expected 200ms, got %v", d2)
	}
	if d3 != 400*time.Millisecond {
		t.Errorf("expected 400ms, got %v", d3)
	}
}

func TestThrottle_CapsAtMaxDelay(t *testing.T) {
	policy := ThrottlePolicy{
		MinDelay: 1 * time.Second,
		MaxDelay: 4 * time.Second,
		Factor:   2.0,
	}
	th := NewThrottle(policy)

	th.Next() // 1s -> next=2s
	th.Next() // 2s -> next=4s
	th.Next() // 4s -> next capped at 4s
	d := th.Next()

	if d != 4*time.Second {
		t.Errorf("expected max 4s, got %v", d)
	}
}

func TestThrottle_Reset(t *testing.T) {
	policy := DefaultThrottlePolicy()
	th := NewThrottle(policy)

	th.Next()
	th.Next()
	th.Reset()

	if th.Current() != policy.MinDelay {
		t.Errorf("after reset expected %v, got %v", policy.MinDelay, th.Current())
	}
	if th.Resets() != 1 {
		t.Errorf("expected 1 reset, got %d", th.Resets())
	}
}

func TestThrottle_DefaultPolicy(t *testing.T) {
	p := DefaultThrottlePolicy()
	if p.MinDelay <= 0 {
		t.Error("MinDelay should be positive")
	}
	if p.MaxDelay <= p.MinDelay {
		t.Error("MaxDelay should exceed MinDelay")
	}
	if p.Factor <= 1.0 {
		t.Error("Factor should be greater than 1.0")
	}
}
