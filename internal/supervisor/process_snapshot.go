package supervisor

import (
	"sync"
	"time"
)

// ProcessSnapshot captures a point-in-time view of a process's state.
type ProcessSnapshot struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	PID       int               `json:"pid,omitempty"`
	ExitCode  int               `json:"exit_code,omitempty"`
	Restarts  int               `json:"restarts"`
	Uptime    float64           `json:"uptime_seconds"`
	Labels    map[string]string `json:"labels,omitempty"`
	TakenAt   time.Time         `json:"taken_at"`
}

// ProcessSnapshotStore records and retrieves snapshots for processes.
type ProcessSnapshotStore struct {
	mu        sync.RWMutex
	snapshots map[string][]ProcessSnapshot
}

// NewProcessSnapshotStore creates an empty snapshot store.
func NewProcessSnapshotStore() *ProcessSnapshotStore {
	return &ProcessSnapshotStore{
		snapshots: make(map[string][]ProcessSnapshot),
	}
}

// Record saves a snapshot for the given process name.
func (s *ProcessSnapshotStore) Record(snap ProcessSnapshot) {
	if snap.TakenAt.IsZero() {
		snap.TakenAt = time.Now()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snap.Name] = append(s.snapshots[snap.Name], snap)
}

// All returns every snapshot recorded across all processes.
func (s *ProcessSnapshotStore) All() []ProcessSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ProcessSnapshot
	for _, snaps := range s.snapshots {
		out = append(out, snaps...)
	}
	return out
}

// ForProcess returns all snapshots recorded for the named process.
func (s *ProcessSnapshotStore) ForProcess(name string) []ProcessSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make([]ProcessSnapshot, len(s.snapshots[name]))
	copy(copied, s.snapshots[name])
	return copied
}

// Latest returns the most recent snapshot for the named process, and whether one exists.
func (s *ProcessSnapshotStore) Latest(name string) (ProcessSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snaps := s.snapshots[name]
	if len(snaps) == 0 {
		return ProcessSnapshot{}, false
	}
	return snaps[len(snaps)-1], true
}
