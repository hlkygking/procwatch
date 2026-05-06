package supervisor

import (
	"sync"
	"time"
)

// HealthStatus represents the current health state of a supervised process.
type HealthStatus int

const (
	StatusUnknown HealthStatus = iota
	StatusHealthy
	StatusUnhealthy
	StatusStarting
)

func (h HealthStatus) String() string {
	switch h {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	case StatusStarting:
		return "starting"
	default:
		return "unknown"
	}
}

// HealthRecord holds a snapshot of a process's health at a point in time.
type HealthRecord struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Restarts  int          `json:"restarts"`
	LastStart time.Time    `json:"last_start"`
	LastExit  *time.Time   `json:"last_exit,omitempty"`
	ExitCode  *int         `json:"exit_code,omitempty"`
}

// HealthTracker tracks runtime health state for a set of processes.
type HealthTracker struct {
	mu      sync.RWMutex
	records map[string]*HealthRecord
}

// NewHealthTracker creates an empty HealthTracker.
func NewHealthTracker() *HealthTracker {
	return &HealthTracker{
		records: make(map[string]*HealthRecord),
	}
}

// SetStatus updates the health status for a named process.
func (h *HealthTracker) SetStatus(name string, status HealthStatus) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rec := h.getOrCreate(name)
	rec.Status = status
}

// RecordStart marks a process as started.
func (h *HealthTracker) RecordStart(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rec := h.getOrCreate(name)
	now := time.Now()
	rec.LastStart = now
	rec.Status = StatusStarting
}

// RecordExit records the exit of a process along with its exit code.
func (h *HealthTracker) RecordExit(name string, code int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rec := h.getOrCreate(name)
	now := time.Now()
	rec.LastExit = &now
	rec.ExitCode = &code
	rec.Restarts++
	rec.Status = StatusUnhealthy
}

// Get returns a copy of the HealthRecord for the named process.
func (h *HealthTracker) Get(name string) (HealthRecord, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rec, ok := h.records[name]
	if !ok {
		return HealthRecord{}, false
	}
	return *rec, true
}

// All returns a snapshot of all tracked health records.
func (h *HealthTracker) All() []HealthRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]HealthRecord, 0, len(h.records))
	for _, rec := range h.records {
		out = append(out, *rec)
	}
	return out
}

func (h *HealthTracker) getOrCreate(name string) *HealthRecord {
	if _, ok := h.records[name]; !ok {
		h.records[name] = &HealthRecord{Name: name, Status: StatusUnknown}
	}
	return h.records[name]
}
