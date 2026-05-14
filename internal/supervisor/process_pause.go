package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// PauseState represents whether a process is paused or active.
type PauseState int

const (
	PauseStateActive PauseState = iota
	PauseStatePaused
)

func (p PauseState) String() string {
	switch p {
	case PauseStateActive:
		return "active"
	case PauseStatePaused:
		return "paused"
	default:
		return "unknown"
	}
}

// PauseEntry records a pause or resume event for a process.
type PauseEntry struct {
	Process   string
	State     PauseState
	Reason    string
	Timestamp time.Time
}

// ProcessPauseStore tracks pause/resume state per process.
type ProcessPauseStore struct {
	mu      sync.RWMutex
	states  map[string]PauseState
	history []PauseEntry
}

// NewProcessPauseStore creates a new ProcessPauseStore.
func NewProcessPauseStore() *ProcessPauseStore {
	return &ProcessPauseStore{
		states: make(map[string]PauseState),
	}
}

// Pause marks a process as paused, recording the reason.
func (s *ProcessPauseStore) Pause(process, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[process] = PauseStatePaused
	s.history = append(s.history, PauseEntry{
		Process:   process,
		State:     PauseStatePaused,
		Reason:    reason,
		Timestamp: time.Now(),
	})
}

// Resume marks a process as active.
func (s *ProcessPauseStore) Resume(process, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[process] = PauseStateActive
	s.history = append(s.history, PauseEntry{
		Process:   process,
		State:     PauseStateActive,
		Reason:    reason,
		Timestamp: time.Now(),
	})
}

// IsPaused returns true if the process is currently paused.
func (s *ProcessPauseStore) IsPaused(process string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[process] == PauseStatePaused
}

// State returns the current PauseState for a process.
func (s *ProcessPauseStore) State(process string) PauseState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[process]
}

// History returns all recorded pause/resume entries.
func (s *ProcessPauseStore) History() []PauseEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PauseEntry, len(s.history))
	copy(out, s.history)
	return out
}

// ForProcess returns pause/resume history for a specific process.
func (s *ProcessPauseStore) ForProcess(process string) []PauseEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []PauseEntry
	for _, e := range s.history {
		if e.Process == process {
			out = append(out, e)
		}
	}
	return out
}

// String returns a human-readable summary for a process pause entry.
func (e PauseEntry) String() string {
	return fmt.Sprintf("%s process=%s reason=%q", e.State, e.Process, e.Reason)
}
