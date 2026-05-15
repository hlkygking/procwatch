package supervisor

import (
	"strings"
	"testing"
	"time"
)

func TestProcessBackoffPolicy_FixedDelay(t *testing.T) {
	p := ProcessBackoffPolicy{Strategy: BackoffFixed, BaseDelay: 5 * time.Second, MaxDelay: 30 * time.Second}
	for _, attempt := range []int{1, 2, 5, 10} {
		if got := p.Delay(attempt); got != 5*time.Second {
			t.Errorf("attempt %d: expected 5s, got %s", attempt, got)
		}
	}
}

func TestProcessBackoffPolicy_LinearDelay(t *testing.T) {
	p := ProcessBackoffPolicy{Strategy: BackoffLinear, BaseDelay: 2 * time.Second, MaxDelay: 20 * time.Second}
	if got := p.Delay(1); got != 2*time.Second {
		t.Errorf("attempt 1: expected 2s, got %s", got)
	}
	if got := p.Delay(3); got != 6*time.Second {
		t.Errorf("attempt 3: expected 6s, got %s", got)
	}
	if got := p.Delay(20); got != 20*time.Second {
		t.Errorf("attempt 20: expected cap at 20s, got %s", got)
	}
}

func TestProcessBackoffPolicy_ExponentialDelay(t *testing.T) {
	p := ProcessBackoffPolicy{Strategy: BackoffExponential, BaseDelay: 1 * time.Second, MaxDelay: 16 * time.Second, Multiplier: 2.0}
	expected := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 16 * time.Second}
	for i, want := range expected {
		if got := p.Delay(i + 1); got != want {
			t.Errorf("attempt %d: expected %s, got %s", i+1, want, got)
		}
	}
}

func TestProcessBackoffPolicy_DefaultPolicy(t *testing.T) {
	p := DefaultBackoffPolicy()
	if p.Strategy != BackoffExponential {
		t.Errorf("expected exponential, got %s", p.Strategy)
	}
	if p.BaseDelay != time.Second {
		t.Errorf("expected 1s base, got %s", p.BaseDelay)
	}
}

func TestProcessBackoffPolicy_String(t *testing.T) {
	p := DefaultBackoffPolicy()
	s := p.String()
	if !strings.Contains(s, "exponential") {
		t.Errorf("expected string to contain 'exponential', got: %s", s)
	}
}

func TestProcessBackoffStore_RecordAndReset(t *testing.T) {
	store := NewProcessBackoffStore()
	store.SetPolicy("worker", ProcessBackoffPolicy{
		Strategy: BackoffFixed, BaseDelay: 3 * time.Second, MaxDelay: 10 * time.Second,
	})
	attempt, _ := store.RecordAttempt("worker")
	if attempt != 1 {
		t.Errorf("expected attempt 1, got %d", attempt)
	}
	store.RecordAttempt("worker")
	if store.Attempts("worker") != 2 {
		t.Errorf("expected 2 attempts, got %d", store.Attempts("worker"))
	}
	store.ResetAttempts("worker")
	if store.Attempts("worker") != 0 {
		t.Errorf("expected 0 after reset, got %d", store.Attempts("worker"))
	}
}

func TestProcessBackoffStore_DefaultPolicyWhenUnset(t *testing.T) {
	store := NewProcessBackoffStore()
	p := store.GetPolicy("unknown")
	if p.Strategy != BackoffExponential {
		t.Errorf("expected default exponential policy, got %s", p.Strategy)
	}
}

func TestProcessBackoffReporter_PrintTable(t *testing.T) {
	store := NewProcessBackoffStore()
	store.SetPolicy("api", DefaultBackoffPolicy())
	store.RecordAttempt("api")
	var buf strings.Builder
	r := NewProcessBackoffReporter(store, &buf)
	r.PrintTable([]string{"api"})
	if !strings.Contains(buf.String(), "api") {
		t.Error("expected output to contain process name 'api'")
	}
	if !strings.Contains(buf.String(), "exponential") {
		t.Error("expected output to contain strategy 'exponential'")
	}
}

func TestProcessBackoffReporter_PrintJSON(t *testing.T) {
	store := NewProcessBackoffStore()
	store.SetPolicy("worker", DefaultBackoffPolicy())
	var buf strings.Builder
	r := NewProcessBackoffReporter(store, &buf)
	if err := r.PrintJSON([]string{"worker"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"process"`) {
		t.Error("expected JSON output to contain 'process' field")
	}
}

func TestProcessBackoffReporter_NilWriter_UsesStdout(t *testing.T) {
	store := NewProcessBackoffStore()
	r := NewProcessBackoffReporter(store, nil)
	if r.writer == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
