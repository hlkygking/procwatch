package supervisor

import "sync"

// ProcessLifecycleStore manages per-process lifecycle trackers.
type ProcessLifecycleStore struct {
	mu      sync.Mutex
	entries map[string]*ProcessLifecycle
}

// NewProcessLifecycleStore returns an initialised store.
func NewProcessLifecycleStore() *ProcessLifecycleStore {
	return &ProcessLifecycleStore{
		entries: make(map[string]*ProcessLifecycle),
	}
}

// GetOrCreate returns the lifecycle tracker for the given process,
// creating one if it does not yet exist.
func (s *ProcessLifecycleStore) GetOrCreate(process string) *ProcessLifecycle {
	s.mu.Lock()
	defer s.mu.Unlock()
	if lc, ok := s.entries[process]; ok {
		return lc
	}
	lc := NewProcessLifecycle(process)
	s.entries[process] = lc
	return lc
}

// Get returns the lifecycle tracker for the given process, or nil if not found.
func (s *ProcessLifecycleStore) Get(process string) *ProcessLifecycle {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries[process]
}

// AllRecords returns a flat slice of every lifecycle record across all processes.
func (s *ProcessLifecycleStore) AllRecords() []LifecycleRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []LifecycleRecord
	for _, lc := range s.entries {
		out = append(out, lc.All()...)
	}
	return out
}

// ProcessNames returns the names of all tracked processes.
func (s *ProcessLifecycleStore) ProcessNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.entries))
	for name := range s.entries {
		names = append(names, name)
	}
	return names
}
