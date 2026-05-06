package supervisor

import "sync"

// ProcessStateStore holds ProcessState instances keyed by process name.
type ProcessStateStore struct {
	mu     sync.RWMutex
	states map[string]*ProcessState
}

// NewProcessStateStore creates an empty ProcessStateStore.
func NewProcessStateStore() *ProcessStateStore {
	return &ProcessStateStore{
		states: make(map[string]*ProcessState),
	}
}

// GetOrCreate returns the existing ProcessState for name, or creates one.
func (s *ProcessStateStore) GetOrCreate(name string) *ProcessState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ps, ok := s.states[name]; ok {
		return ps
	}
	ps := NewProcessState(name)
	s.states[name] = ps
	return ps
}

// Get returns the ProcessState for name and whether it was found.
func (s *ProcessStateStore) Get(name string) (*ProcessState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ps, ok := s.states[name]
	return ps, ok
}

// Snapshots returns snapshots of all tracked processes.
func (s *ProcessStateStore) Snapshots() []ProcessStateSnapshot {
	s.mu.RLock()
	names := make([]string, 0, len(s.states))
	for name := range s.states {
		names = append(names, name)
	}
	s.mu.RUnlock()

	snaps := make([]ProcessStateSnapshot, 0, len(names))
	for _, name := range names {
		if ps, ok := s.Get(name); ok {
			snaps = append(snaps, ps.Snapshot())
		}
	}
	return snaps
}

// Len returns the number of tracked processes.
func (s *ProcessStateStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.states)
}
