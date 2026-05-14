package supervisor

import (
	"testing"
	"time"
)

func TestProcessRetryPolicy_ShouldRetry_Never(t *testing.T) {
	p := ProcessRetryPolicy{Behavior: RetryNever, MaxAttempts: 10}
	if p.ShouldRetry(1, 0) {
		t.Error("expected ShouldRetry=false for RetryNever")
	}
}

func TestProcessRetryPolicy_ShouldRetry_Always(t *testing.T) {
	p := ProcessRetryPolicy{Behavior: RetryAlways, MaxAttempts: 0}
	if !p.ShouldRetry(0, 99) {
		t.Error("expected ShouldRetry=true for RetryAlways with unlimited attempts")
	}
}

func TestProcessRetryPolicy_ShouldRetry_OnError_ZeroExit(t *testing.T) {
	p := ProcessRetryPolicy{Behavior: RetryOnError, MaxAttempts: 5}
	if p.ShouldRetry(0, 1) {
		t.Error("expected ShouldRetry=false for zero exit with RetryOnError")
	}
}

func TestProcessRetryPolicy_ShouldRetry_OnError_NonZeroExit(t *testing.T) {
	p := ProcessRetryPolicy{Behavior: RetryOnError, MaxAttempts: 5}
	if !p.ShouldRetry(1, 1) {
		t.Error("expected ShouldRetry=true for non-zero exit with RetryOnError")
	}
}

func TestProcessRetryPolicy_MaxAttemptsEnforced(t *testing.T) {
	p := ProcessRetryPolicy{Behavior: RetryAlways, MaxAttempts: 3}
	if p.ShouldRetry(0, 3) {
		t.Error("expected ShouldRetry=false when attempts >= MaxAttempts")
	}
	if !p.ShouldRetry(0, 2) {
		t.Error("expected ShouldRetry=true when attempts < MaxAttempts")
	}
}

func TestProcessRetryPolicy_NextDelay_BaseDelay(t *testing.T) {
	p := ProcessRetryPolicy{
		Delay:    100 * time.Millisecond,
		MaxDelay: 10 * time.Second,
	}
	if got := p.NextDelay(0); got != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %s", got)
	}
}

func TestProcessRetryPolicy_NextDelay_Exponential(t *testing.T) {
	p := ProcessRetryPolicy{
		Delay:    100 * time.Millisecond,
		MaxDelay: 10 * time.Second,
	}
	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, want := range expected {
		if got := p.NextDelay(i); got != want {
			t.Errorf("attempt %d: expected %s, got %s", i, want, got)
		}
	}
}

func TestProcessRetryPolicy_NextDelay_CapsAtMaxDelay(t *testing.T) {
	p := ProcessRetryPolicy{
		Delay:    1 * time.Second,
		MaxDelay: 4 * time.Second,
	}
	if got := p.NextDelay(10); got != 4*time.Second {
		t.Errorf("expected delay capped at 4s, got %s", got)
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	p := DefaultRetryPolicy()
	if p.Behavior != RetryOnError {
		t.Errorf("expected RetryOnError, got %v", p.Behavior)
	}
	if p.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", p.MaxAttempts)
	}
}

func TestProcessRetryPolicy_String(t *testing.T) {
	p := DefaultRetryPolicy()
	s := p.String()
	if s == "" {
		t.Error("expected non-empty String()")
	}
	if !contains(s, "on-error") {
		t.Errorf("expected 'on-error' in String(), got: %s", s)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
