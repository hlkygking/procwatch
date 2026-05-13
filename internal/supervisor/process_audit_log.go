package supervisor

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// AuditEventKind classifies the type of audit event.
type AuditEventKind string

const (
	AuditEventStart   AuditEventKind = "start"
	AuditEventStop    AuditEventKind = "stop"
	AuditEventRestart AuditEventKind = "restart"
	AuditEventKill    AuditEventKind = "kill"
	AuditEventSkip    AuditEventKind = "skip"
)

// AuditEntry represents a single structured audit log entry.
type AuditEntry struct {
	Timestamp   time.Time      `json:"timestamp"`
	Process     string         `json:"process"`
	Kind        AuditEventKind `json:"kind"`
	ExitCode    *int           `json:"exit_code,omitempty"`
	Message     string         `json:"message,omitempty"`
	RestartNum  int            `json:"restart_num,omitempty"`
}

// ProcessAuditLog records lifecycle audit events for all supervised processes.
type ProcessAuditLog struct {
	mu      sync.Mutex
	entries []AuditEntry
	w       io.Writer
}

// NewProcessAuditLog creates a new audit log. If w is nil, os.Stdout is used.
func NewProcessAuditLog(w io.Writer) *ProcessAuditLog {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessAuditLog{w: w}
}

// Record appends an audit entry and writes it as a JSON line to the writer.
func (a *ProcessAuditLog) Record(entry AuditEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, entry)
	data, _ := json.Marshal(entry)
	_, _ = a.w.Write(append(data, '\n'))
}

// All returns a copy of all recorded audit entries.
func (a *ProcessAuditLog) All() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AuditEntry, len(a.entries))
	copy(out, a.entries)
	return out
}

// FilterByProcess returns entries for a specific process name.
func (a *ProcessAuditLog) FilterByProcess(name string) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	var out []AuditEntry
	for _, e := range a.entries {
		if e.Process == name {
			out = append(out, e)
		}
	}
	return out
}

// FilterByKind returns entries matching a specific event kind.
func (a *ProcessAuditLog) FilterByKind(kind AuditEventKind) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	var out []AuditEntry
	for _, e := range a.entries {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}
