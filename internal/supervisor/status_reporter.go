package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ProcessStatus represents a snapshot of a process's current state.
type ProcessStatus struct {
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	PID         int           `json:"pid,omitempty"`
	Restarts    int           `json:"restarts"`
	LastExitCode int          `json:"last_exit_code,omitempty"`
	Uptime      string        `json:"uptime,omitempty"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
}

// StatusReporter collects and formats process status snapshots.
type StatusReporter struct {
	health  *HealthTracker
	writer  io.Writer
}

// NewStatusReporter creates a StatusReporter using the given HealthTracker.
// If writer is nil, os.Stdout is used.
func NewStatusReporter(health *HealthTracker, writer io.Writer) *StatusReporter {
	if writer == nil {
		writer = os.Stdout
	}
	return &StatusReporter{
		health: health,
		writer: writer,
	}
}

// Snapshot returns a slice of ProcessStatus for all tracked processes.
func (sr *StatusReporter) Snapshot() []ProcessStatus {
	records := sr.health.All()
	statuses := make([]ProcessStatus, 0, len(records))
	for _, r := range records {
		ps := ProcessStatus{
			Name:         r.Name,
			Status:       r.Status.String(),
			PID:          r.PID,
			Restarts:     r.Restarts,
			LastExitCode: r.LastExitCode,
		}
		if r.StartedAt != nil {
			t := *r.StartedAt
			ps.StartedAt = &t
			if r.Status == StatusRunning {
				ps.Uptime = time.Since(t).Truncate(time.Second).String()
			}
		}
		statuses = append(statuses, ps)
	}
	return statuses
}

// PrintJSON writes a JSON-encoded status snapshot to the reporter's writer.
func (sr *StatusReporter) PrintJSON() error {
	snap := sr.Snapshot()
	enc := json.NewEncoder(sr.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(snap)
}

// PrintTable writes a human-readable table of process statuses to the writer.
func (sr *StatusReporter) PrintTable() {
	snap := sr.Snapshot()
	fmt.Fprintf(sr.writer, "%-20s %-12s %-8s %-10s %s\n", "NAME", "STATUS", "PID", "RESTARTS", "UPTIME")
	fmt.Fprintf(sr.writer, "%s\n", "------------------------------------------------------------")
	for _, ps := range snap {
		uptime := ps.Uptime
		if uptime == "" {
			uptime = "-"
		}
		pid := "-"
		if ps.PID != 0 {
			pid = fmt.Sprintf("%d", ps.PID)
		}
		fmt.Fprintf(sr.writer, "%-20s %-12s %-8s %-10d %s\n", ps.Name, ps.Status, pid, ps.Restarts, uptime)
	}
}
