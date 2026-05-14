package supervisor

import (
	"os"
	"testing"
)

func TestProcessSignalPolicyStore_DefaultPolicy(t *testing.T) {
	store := NewProcessSignalPolicyStore()
	p := store.Get("missing")
	if p.ShutdownSignal != os.Interrupt {
		t.Errorf("expected os.Interrupt for shutdown, got %v", p.ShutdownSignal)
	}
	if p.RestartSignal != os.Interrupt {
		t.Errorf("expected os.Interrupt for restart, got %v", p.RestartSignal)
	}
}

func TestProcessSignalPolicyStore_SetAndGet(t *testing.T) {
	store := NewProcessSignalPolicyStore()
	policy := SignalPolicy{
		ShutdownSignal: os.Interrupt,
		RestartSignal:  os.Interrupt,
	}
	store.Set("worker", policy)
	got := store.Get("worker")
	if got.ShutdownSignal != policy.ShutdownSignal {
		t.Errorf("shutdown signal mismatch: got %v, want %v", got.ShutdownSignal, policy.ShutdownSignal)
	}
}

func TestProcessSignalPolicyStore_All(t *testing.T) {
	store := NewProcessSignalPolicyStore()
	store.Set("alpha", DefaultSignalPolicy())
	store.Set("beta", DefaultSignalPolicy())
	all := store.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
	if _, ok := all["alpha"]; !ok {
		t.Error("expected 'alpha' in All()")
	}
	if _, ok := all["beta"]; !ok {
		t.Error("expected 'beta' in All()")
	}
}

func TestParseSignalName_Valid(t *testing.T) {
	cases := []struct {
		input string
	}{
		{"SIGINT"},
		{"sigint"},
		{"INT"},
		{"SIGTERM"},
		{"sigterm"},
		{"TERM"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			sig, err := ParseSignalName(tc.input)
			if err != nil {
				t.Errorf("unexpected error for %q: %v", tc.input, err)
			}
			if sig == nil {
				t.Errorf("expected non-nil signal for %q", tc.input)
			}
		})
	}
}

func TestParseSignalName_Invalid(t *testing.T) {
	_, err := ParseSignalName("SIGUSR1")
	if err == nil {
		t.Error("expected error for unsupported signal SIGUSR1")
	}
}

func TestDefaultSignalPolicy(t *testing.T) {
	p := DefaultSignalPolicy()
	if p.ShutdownSignal == nil {
		t.Error("expected non-nil ShutdownSignal")
	}
	if p.RestartSignal == nil {
		t.Error("expected non-nil RestartSignal")
	}
}
