package supervisor

import "sync"

// ProcessMetricsStore manages metrics for all tracked processes.
type ProcessMetricsStore struct {
	mu      sync.RWMutex
	entries map[string]*ProcessMetrics
}

// NewProcessMetricsStore creates an empty metrics store.
func NewProcessMetricsStore() *ProcessMetricsStore {
	return &ProcessMetricsStore{
		entries: make(map[string]*ProcessMetrics),
	}
}

// GetOrCreate returns existing metrics for the process or creates new ones.
func (s *ProcessMetricsStore) GetOrCreate(name string) *ProcessMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.entries[name]; ok {
		return m
	}
	m := &ProcessMetrics{Name: name}
	s.entries[name] = m
	return m
}

// Get returns the metrics for a process, or nil if not found.
func (s *ProcessMetricsStore) Get(name string) (*ProcessMetrics, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.entries[name]
	return m, ok
}

// Snapshots returns a snapshot of all tracked process metrics.
func (s *ProcessMetricsStore) Snapshots() []ProcessMetricsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessMetricsSnapshot, 0, len(s.entries))
	for _, m := range s.entries {
		out = append(out, m.Snapshot())
	}
	return out
}
