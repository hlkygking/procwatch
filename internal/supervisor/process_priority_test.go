package supervisor

import (
	"testing"
)

func TestParsePriority_Valid(t *testing.T) {
	cases := []struct {
		input    string
		want     Priority
	}{
		{"low", PriorityLow},
		{"normal", PriorityNormal},
		{"", PriorityNormal},
		{"high", PriorityHigh},
	}
	for _, tc := range cases {
		p, err := ParsePriority(tc.input)
		if err != nil {
			t.Fatalf("ParsePriority(%q) unexpected error: %v", tc.input, err)
		}
		if p != tc.want {
			t.Errorf("ParsePriority(%q) = %v, want %v", tc.input, p, tc.want)
		}
	}
}

func TestParsePriority_Invalid(t *testing.T) {
	_, err := ParsePriority("critical")
	if err == nil {
		t.Fatal("expected error for unknown priority")
	}
}

func TestPriority_String(t *testing.T) {
	if PriorityLow.String() != "low" {
		t.Errorf("expected 'low'")
	}
	if PriorityNormal.String() != "normal" {
		t.Errorf("expected 'normal'")
	}
	if PriorityHigh.String() != "high" {
		t.Errorf("expected 'high'")
	}
}

func TestProcessPriorityStore_SetAndGet(t *testing.T) {
	s := NewProcessPriorityStore()
	s.Set("web", PriorityHigh)
	p, ok := s.Get("web")
	if !ok {
		t.Fatal("expected to find 'web'")
	}
	if p != PriorityHigh {
		t.Errorf("got %v, want high", p)
	}
}

func TestProcessPriorityStore_GetMissing(t *testing.T) {
	s := NewProcessPriorityStore()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected missing process to return false")
	}
}

func TestProcessPriorityStore_All_SortedDescending(t *testing.T) {
	s := NewProcessPriorityStore()
	s.Set("worker", PriorityLow)
	s.Set("api", PriorityHigh)
	s.Set("cache", PriorityNormal)

	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i].Priority > all[i-1].Priority {
			t.Errorf("entries not sorted descending at index %d", i)
		}
	}
}

func TestProcessPriorityStore_OverwriteExisting(t *testing.T) {
	s := NewProcessPriorityStore()
	s.Set("db", PriorityLow)
	s.Set("db", PriorityHigh)
	p, _ := s.Get("db")
	if p != PriorityHigh {
		t.Errorf("expected high after overwrite, got %v", p)
	}
}
