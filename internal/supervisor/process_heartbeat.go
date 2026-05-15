package supervisor

import (
	"sync"
	"time"
)

// HeartbeatStatus represents the last known heartbeat state of a process.
type HeartbeatStatus struct {
	Process   string
	LastBeat  time.Time
	Missed    int
	Alive     bool
}

// ProcessHeartbeatStore tracks heartbeat signals emitted by managed processes.
type ProcessHeartbeatStore struct {
	mu      sync.RWMutex
	entries map[string]*HeartbeatStatus
	timeout time.Duration
}

// NewProcessHeartbeatStore creates a store with the given stale timeout.
// A process is considered dead if no heartbeat is received within timeout.
func NewProcessHeartbeatStore(timeout time.Duration) *ProcessHeartbeatStore {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ProcessHeartbeatStore{
		entries: make(map[string]*HeartbeatStatus),
		timeout: timeout,
	}
}

// Beat records a heartbeat for the named process.
func (s *ProcessHeartbeatStore) Beat(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[process]
	if !ok {
		entry = &HeartbeatStatus{Process: process}
		s.entries[process] = entry
	}
	entry.LastBeat = time.Now()
	entry.Missed = 0
	entry.Alive = true
}

// Check evaluates all tracked processes and marks those that have exceeded
// the stale timeout as dead, incrementing their missed count.
func (s *ProcessHeartbeatStore) Check() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range s.entries {
		if entry.Alive && now.Sub(entry.LastBeat) > s.timeout {
			entry.Missed++
			entry.Alive = false
		}
	}
}

// Status returns the current heartbeat status for a process.
// Returns false if the process has never been seen.
func (s *ProcessHeartbeatStore) Status(process string) (HeartbeatStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[process]
	if !ok {
		return HeartbeatStatus{}, false
	}
	return *entry, true
}

// All returns a snapshot of all tracked heartbeat statuses.
func (s *ProcessHeartbeatStore) All() []HeartbeatStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]HeartbeatStatus, 0, len(s.entries))
	for _, entry := range s.entries {
		out = append(out, *entry)
	}
	return out
}

// Remove stops tracking heartbeats for the named process.
func (s *ProcessHeartbeatStore) Remove(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, process)
}
