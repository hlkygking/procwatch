package supervisor

import (
	"testing"
	"time"
)

func TestProcessTracingStore_RecordAndAll(t *testing.T) {
	s := NewProcessTracingStore()
	s.Record(TraceEvent{Process: "web", Kind: "start", StartedAt: time.Now()})
	s.Record(TraceEvent{Process: "worker", Kind: "stop", StartedAt: time.Now()})
	if got := len(s.All()); got != 2 {
		t.Fatalf("expected 2 events, got %d", got)
	}
}

func TestProcessTracingStore_DurationComputed(t *testing.T) {
	s := NewProcessTracingStore()
	start := time.Now().Add(-100 * time.Millisecond)
	s.Record(TraceEvent{Process: "web", Kind: "run", StartedAt: start})
	events := s.All()
	if events[0].Duration < 50*time.Millisecond {
		t.Fatalf("expected duration >= 50ms, got %s", events[0].Duration)
	}
}

func TestProcessTracingStore_ForProcess_Filtered(t *testing.T) {
	s := NewProcessTracingStore()
	s.Record(TraceEvent{Process: "web", Kind: "start", StartedAt: time.Now()})
	s.Record(TraceEvent{Process: "db", Kind: "start", StartedAt: time.Now()})
	s.Record(TraceEvent{Process: "web", Kind: "stop", StartedAt: time.Now()})
	got := s.ForProcess("web")
	if len(got) != 2 {
		t.Fatalf("expected 2 events for web, got %d", len(got))
	}
}

func TestProcessTracingStore_ForProcess_NoMatch(t *testing.T) {
	s := NewProcessTracingStore()
	s.Record(TraceEvent{Process: "web", Kind: "start", StartedAt: time.Now()})
	got := s.ForProcess("missing")
	if len(got) != 0 {
		t.Fatalf("expected 0 events, got %d", len(got))
	}
}

func TestProcessTracingStore_Clear(t *testing.T) {
	s := NewProcessTracingStore()
	s.Record(TraceEvent{Process: "web", Kind: "start", StartedAt: time.Now()})
	s.Clear()
	if got := len(s.All()); got != 0 {
		t.Fatalf("expected 0 after clear, got %d", got)
	}
}

func TestProcessTracingStore_MetaPreserved(t *testing.T) {
	s := NewProcessTracingStore()
	s.Record(TraceEvent{
		Process:   "api",
		Kind:      "restart",
		StartedAt: time.Now(),
		Meta:      map[string]string{"reason": "crash"},
	})
	events := s.All()
	if events[0].Meta["reason"] != "crash" {
		t.Fatalf("expected meta reason=crash, got %v", events[0].Meta)
	}
}
