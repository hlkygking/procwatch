package supervisor

import (
	"sync"
	"time"
)

// CheckpointKind describes the type of checkpoint event.
type CheckpointKind string

const (
	CheckpointStarted  CheckpointKind = "started"
	CheckpointReady    CheckpointKind = "ready"
	CheckpointStopped  CheckpointKind = "stopped"
	CheckpointFailed   CheckpointKind = "failed"
	CheckpointRestored CheckpointKind = "restored"
)

// ProcessCheckpoint records a named milestone for a process at a point in time.
type ProcessCheckpoint struct {
	Process   string
	Kind      CheckpointKind
	Timestamp time.Time
	Meta      map[string]string
}

// ProcessCheckpointLog stores checkpoints per process.
type ProcessCheckpointLog struct {
	mu      sync.RWMutex
	entries []ProcessCheckpoint
}

// NewProcessCheckpointLog returns an empty checkpoint log.
func NewProcessCheckpointLog() *ProcessCheckpointLog {
	return &ProcessCheckpointLog{}
}

// Record appends a checkpoint for the given process.
func (l *ProcessCheckpointLog) Record(process string, kind CheckpointKind, meta map[string]string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cp := ProcessCheckpoint{
		Process:   process,
		Kind:      kind,
		Timestamp: time.Now(),
		Meta:      meta,
	}
	l.entries = append(l.entries, cp)
}

// All returns a copy of all recorded checkpoints.
func (l *ProcessCheckpointLog) All() []ProcessCheckpoint {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]ProcessCheckpoint, len(l.entries))
	copy(out, l.entries)
	return out
}

// ForProcess returns checkpoints for a specific process name.
func (l *ProcessCheckpointLog) ForProcess(name string) []ProcessCheckpoint {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []ProcessCheckpoint
	for _, e := range l.entries {
		if e.Process == name {
			out = append(out, e)
		}
	}
	return out
}

// LastOf returns the most recent checkpoint of a given kind for a process, or nil.
func (l *ProcessCheckpointLog) LastOf(process string, kind CheckpointKind) *ProcessCheckpoint {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i := len(l.entries) - 1; i >= 0; i-- {
		e := l.entries[i]
		if e.Process == process && e.Kind == kind {
			cp := e
			return &cp
		}
	}
	return nil
}
