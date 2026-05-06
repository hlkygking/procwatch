package supervisor

import "time"

// LifecycleEvent represents a named transition in a process's lifecycle.
type LifecycleEvent string

const (
	LifecycleStarting  LifecycleEvent = "starting"
	LifecycleStarted   LifecycleEvent = "started"
	LifecycleRestarting LifecycleEvent = "restarting"
	LifecycleStopping  LifecycleEvent = "stopping"
	LifecycleStopped   LifecycleEvent = "stopped"
	LifecycleFailed    LifecycleEvent = "failed"
)

// LifecycleRecord captures a single lifecycle transition with metadata.
type LifecycleRecord struct {
	Process   string
	Event     LifecycleEvent
	Timestamp time.Time
	ExitCode  *int
	Message   string
}

// ProcessLifecycle tracks lifecycle transitions for a single process.
type ProcessLifecycle struct {
	process string
	records []LifecycleRecord
}

// NewProcessLifecycle creates a new lifecycle tracker for the given process name.
func NewProcessLifecycle(process string) *ProcessLifecycle {
	return &ProcessLifecycle{process: process}
}

// Record appends a lifecycle event with the current timestamp.
func (l *ProcessLifecycle) Record(event LifecycleEvent, message string) {
	l.records = append(l.records, LifecycleRecord{
		Process:   l.process,
		Event:     event,
		Timestamp: time.Now(),
		Message:   message,
	})
}

// RecordExit appends a lifecycle event that includes an exit code.
func (l *ProcessLifecycle) RecordExit(event LifecycleEvent, code int, message string) {
	l.records = append(l.records, LifecycleRecord{
		Process:   l.process,
		Event:     event,
		Timestamp: time.Now(),
		ExitCode:  &code,
		Message:   message,
	})
}

// All returns a copy of all recorded lifecycle events.
func (l *ProcessLifecycle) All() []LifecycleRecord {
	out := make([]LifecycleRecord, len(l.records))
	copy(out, l.records)
	return out
}

// Last returns the most recent lifecycle record, or nil if none exist.
func (l *ProcessLifecycle) Last() *LifecycleRecord {
	if len(l.records) == 0 {
		return nil
	}
	r := l.records[len(l.records)-1]
	return &r
}

// CountOf returns how many times the given event has been recorded.
func (l *ProcessLifecycle) CountOf(event LifecycleEvent) int {
	n := 0
	for _, r := range l.records {
		if r.Event == event {
			n++
		}
	}
	return n
}
