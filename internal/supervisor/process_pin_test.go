package supervisor

import (
	"testing"
)

func TestProcessPinStore_SetAndGet(t *testing.T) {
	s := NewProcessPinStore()
	policy := ProcessPinPolicy{Mode: PinModeCPU, Target: 2}
	s.Set("worker", policy)

	got := s.Get("worker")
	if got.Mode != PinModeCPU || got.Target != 2 {
		t.Fatalf("expected cpu:2, got %s", got)
	}
}

func TestProcessPinStore_GetMissing(t *testing.T) {
	s := NewProcessPinStore()
	got := s.Get("nonexistent")
	if got.Mode != PinModeNone {
		t.Fatalf("expected default none policy, got %s", got)
	}
}

func TestProcessPinStore_Remove(t *testing.T) {
	s := NewProcessPinStore()
	s.Set("worker", ProcessPinPolicy{Mode: PinModeNUMA, Target: 1})
	s.Remove("worker")

	got := s.Get("worker")
	if got.Mode != PinModeNone {
		t.Fatalf("expected default after remove, got %s", got)
	}
}

func TestProcessPinStore_All(t *testing.T) {
	s := NewProcessPinStore()
	s.Set("a", ProcessPinPolicy{Mode: PinModeCPU, Target: 0})
	s.Set("b", ProcessPinPolicy{Mode: PinModeNUMA, Target: 1})

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["a"].Mode != PinModeCPU {
		t.Errorf("expected cpu for a, got %s", all["a"].Mode)
	}
	if all["b"].Target != 1 {
		t.Errorf("expected target 1 for b, got %d", all["b"].Target)
	}
}

func TestPinPolicy_String_None(t *testing.T) {
	p := DefaultPinPolicy()
	if p.String() != "none" {
		t.Fatalf("expected 'none', got %q", p.String())
	}
}

func TestPinPolicy_String_CPU(t *testing.T) {
	p := ProcessPinPolicy{Mode: PinModeCPU, Target: 3}
	if p.String() != "cpu:3" {
		t.Fatalf("expected 'cpu:3', got %q", p.String())
	}
}

func TestParsePinMode_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected PinMode
	}{
		{"none", PinModeNone},
		{"cpu", PinModeCPU},
		{"numa", PinModeNUMA},
	}
	for _, c := range cases {
		got, err := ParsePinMode(c.input)
		if err != nil {
			t.Errorf("unexpected error for %q: %v", c.input, err)
		}
		if got != c.expected {
			t.Errorf("expected %q, got %q", c.expected, got)
		}
	}
}

func TestParsePinMode_Invalid(t *testing.T) {
	_, err := ParsePinMode("socket")
	if err == nil {
		t.Fatal("expected error for unknown pin mode")
	}
}
