package supervisor

import (
	"fmt"
	"sync"
)

// Priority represents a process startup/shutdown priority level.
type Priority int

const (
	PriorityLow    Priority = 10
	PriorityNormal Priority = 50
	PriorityHigh   Priority = 90
)

// ProcessPriorityEntry holds the priority assigned to a named process.
type ProcessPriorityEntry struct {
	Process  string
	Priority Priority
}

// ProcessPriorityStore tracks priority assignments for processes.
type ProcessPriorityStore struct {
	mu       sync.RWMutex
	entries  map[string]Priority
}

// NewProcessPriorityStore creates an empty ProcessPriorityStore.
func NewProcessPriorityStore() *ProcessPriorityStore {
	return &ProcessPriorityStore{
		entries: make(map[string]Priority),
	}
}

// Set assigns a priority to a named process.
func (s *ProcessPriorityStore) Set(process string, p Priority) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[process] = p
}

// Get returns the priority for a named process and whether it was found.
func (s *ProcessPriorityStore) Get(process string) (Priority, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.entries[process]
	return p, ok
}

// All returns all priority entries sorted by priority descending.
func (s *ProcessPriorityStore) All() []ProcessPriorityEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessPriorityEntry, 0, len(s.entries))
	for name, p := range s.entries {
		out = append(out, ProcessPriorityEntry{Process: name, Priority: p})
	}
	sortPriorityEntries(out)
	return out
}

// ParsePriority parses a string into a Priority value.
func ParsePriority(s string) (Priority, error) {
	switch s {
	case "low":
		return PriorityLow, nil
	case "normal", "":
		return PriorityNormal, nil
	case "high":
		return PriorityHigh, nil
	default:
		return PriorityNormal, fmt.Errorf("unknown priority %q: use low, normal, or high", s)
	}
}

// String returns the string representation of a Priority.
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityHigh:
		return "high"
	default:
		return "normal"
	}
}

func sortPriorityEntries(entries []ProcessPriorityEntry) {
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].Priority > entries[j-1].Priority; j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
}
