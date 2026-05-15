package supervisor

import (
	"sync"
	"time"
)

// TraceEvent represents a single trace span for a process lifecycle event.
type TraceEvent struct {
	Process   string
	Kind      string
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
	Meta      map[string]string
}

// ProcessTracingStore records trace events per process.
type ProcessTracingStore struct {
	mu     sync.Mutex
	events []TraceEvent
}

// NewProcessTracingStore creates an empty tracing store.
func NewProcessTracingStore() *ProcessTracingStore {
	return &ProcessTracingStore{}
}

// Record appends a completed trace event.
func (s *ProcessTracingStore) Record(e TraceEvent) {
	if e.EndedAt.IsZero() {
		e.EndedAt = time.Now()
	}
	e.Duration = e.EndedAt.Sub(e.StartedAt)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}

// All returns a copy of all recorded trace events.
func (s *ProcessTracingStore) All() []TraceEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]TraceEvent, len(s.events))
	copy(out, s.events)
	return out
}

// ForProcess returns trace events for a specific process.
func (s *ProcessTracingStore) ForProcess(name string) []TraceEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []TraceEvent
	for _, e := range s.events {
		if e.Process == name {
			out = append(out, e)
		}
	}
	return out
}

// Clear removes all stored trace events.
func (s *ProcessTracingStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = nil
}
