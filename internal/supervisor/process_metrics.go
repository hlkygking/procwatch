package supervisor

import (
	"sync"
	"time"
)

// ProcessMetrics holds runtime statistics for a single process.
type ProcessMetrics struct {
	mu           sync.RWMutex
	Name         string
	RestartCount int
	TotalUptime  time.Duration
	LastStarted  time.Time
	LastExitCode int
	Running      bool
}

// RecordStart marks the process as running.
func (m *ProcessMetrics) RecordStart() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastStarted = time.Now()
	m.Running = true
}

// RecordStop marks the process as stopped and accumulates uptime.
func (m *ProcessMetrics) RecordStop(exitCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Running && !m.LastStarted.IsZero() {
		m.TotalUptime += time.Since(m.LastStarted)
	}
	m.LastExitCode = exitCode
	m.Running = false
}

// RecordRestart increments the restart counter.
func (m *ProcessMetrics) RecordRestart() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RestartCount++
}

// Snapshot returns a point-in-time copy of the metrics.
func (m *ProcessMetrics) Snapshot() ProcessMetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	uptime := m.TotalUptime
	if m.Running && !m.LastStarted.IsZero() {
		uptime += time.Since(m.LastStarted)
	}
	return ProcessMetricsSnapshot{
		Name:         m.Name,
		RestartCount: m.RestartCount,
		TotalUptime:  uptime,
		LastStarted:  m.LastStarted,
		LastExitCode: m.LastExitCode,
		Running:      m.Running,
	}
}

// ProcessMetricsSnapshot is an immutable snapshot of ProcessMetrics.
type ProcessMetricsSnapshot struct {
	Name         string
	RestartCount int
	TotalUptime  time.Duration
	LastStarted  time.Time
	LastExitCode int
	Running      bool
}
