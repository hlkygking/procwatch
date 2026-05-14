package supervisor

import "sync"

// ProcessLabelStore manages LabelSets for all tracked processes.
type ProcessLabelStore struct {
	mu     sync.RWMutex
	labels map[string]*LabelSet
}

// NewProcessLabelStore creates an empty ProcessLabelStore.
func NewProcessLabelStore() *ProcessLabelStore {
	return &ProcessLabelStore{
		labels: make(map[string]*LabelSet),
	}
}

// GetOrCreate returns the existing LabelSet for a process or creates a new one.
func (s *ProcessLabelStore) GetOrCreate(process string) *LabelSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ls, ok := s.labels[process]; ok {
		return ls
	}
	ls := NewLabelSet(process)
	s.labels[process] = ls
	return ls
}

// Get returns the LabelSet for a process and whether it exists.
func (s *ProcessLabelStore) Get(process string) (*LabelSet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ls, ok := s.labels[process]
	return ls, ok
}

// Delete removes the LabelSet for a process from the store.
// It returns true if the process was found and removed, false otherwise.
func (s *ProcessLabelStore) Delete(process string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.labels[process]; !ok {
		return false
	}
	delete(s.labels, process)
	return true
}

// All returns clones of all LabelSets in the store.
func (s *ProcessLabelStore) All() []*LabelSet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*LabelSet, 0, len(s.labels))
	for _, ls := range s.labels {
		out = append(out, ls.Clone())
	}
	return out
}

// SelectByLabels returns clones of LabelSets whose labels match all selector pairs.
func (s *ProcessLabelStore) SelectByLabels(selector map[string]string) []*LabelSet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*LabelSet
	for _, ls := range s.labels {
		if ls.Matches(selector) {
			out = append(out, ls.Clone())
		}
	}
	return out
}
