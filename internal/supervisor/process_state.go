package supervisor

import (
	"sync"
	"time"
)

// ProcessState tracks the runtime state of a supervised process.
type ProcessState struct {
	mu         sync.RWMutex
	name       string
	pid        int
	startedAt  time.Time
	stoppedAt  time.Time
	restarts   int
	exitCode   int
	running    bool
}

// NewProcessState creates a new ProcessState for the named process.
func NewProcessState(name string) *ProcessState {
	return &ProcessState{name: name}
}

// RecordStart marks the process as running with the given PID.
func (s *ProcessState) RecordStart(pid int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pid = pid
	s.startedAt = time.Now()
	s.running = true
	s.exitCode = 0
}

// RecordStop marks the process as stopped with the given exit code.
func (s *ProcessState) RecordStop(exitCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.stoppedAt = time.Now()
	s.exitCode = exitCode
	if exitCode != 0 {
		s.restarts++
	}
}

// Snapshot returns an immutable copy of the current state.
func (s *ProcessState) Snapshot() ProcessStateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return ProcessStateSnapshot{
		Name:      s.name,
		PID:       s.pid,
		StartedAt: s.startedAt,
		StoppedAt: s.stoppedAt,
		Restarts:  s.restarts,
		ExitCode:  s.exitCode,
		Running:   s.running,
	}
}

// ProcessStateSnapshot is a point-in-time copy of ProcessState.
type ProcessStateSnapshot struct {
	Name      string
	PID       int
	StartedAt time.Time
	StoppedAt time.Time
	Restarts  int
	ExitCode  int
	Running   bool
}

// Uptime returns the duration the process has been running, or the
// duration it ran before stopping.
func (snap ProcessStateSnapshot) Uptime() time.Duration {
	if snap.Running {
		return time.Since(snap.StartedAt)
	}
	if snap.StoppedAt.IsZero() {
		return 0
	}
	return snap.StoppedAt.Sub(snap.StartedAt)
}
