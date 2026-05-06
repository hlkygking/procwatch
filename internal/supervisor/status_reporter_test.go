package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func makeTrackerWithProcesses(t *testing.T) *HealthTracker {
	t.Helper()
	tracker := NewHealthTracker()

	now := time.Now()
	tracker.RecordStart("web", 1001)
	tracker.SetStatus("web", StatusRunning)
	// Manually adjust started_at for uptime test stability
	tracker.mu.Lock()
	if r, ok := tracker.records["web"]; ok {
		r.StartedAt = &now
		tracker.records["web"] = r
	}
	tracker.mu.Unlock()

	tracker.RecordStart("worker", 1002)
	tracker.RecordExit("worker", 1)
	return tracker
}

func TestStatusReporter_Snapshot(t *testing.T) {
	tracker := makeTrackerWithProcesses(t)
	reporter := NewStatusReporter(tracker, nil)

	snap := reporter.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(snap))
	}

	byName := map[string]ProcessStatus{}
	for _, s := range snap {
		byName[s.Name] = s
	}

	web := byName["web"]
	if web.Status != StatusRunning.String() {
		t.Errorf("expected web status running, got %s", web.Status)
	}
	if web.PID != 1001 {
		t.Errorf("expected PID 1001, got %d", web.PID)
	}
	if web.Uptime == "" {
		t.Error("expected non-empty uptime for running process")
	}

	worker := byName["worker"]
	if worker.LastExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", worker.LastExitCode)
	}
	if worker.Uptime != "" {
		t.Errorf("expected empty uptime for non-running process, got %s", worker.Uptime)
	}
}

func TestStatusReporter_PrintJSON(t *testing.T) {
	tracker := makeTrackerWithProcesses(t)
	var buf bytes.Buffer
	reporter := NewStatusReporter(tracker, &buf)

	if err := reporter.PrintJSON(); err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}

	var result []ProcessStatus
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries in JSON, got %d", len(result))
	}
}

func TestStatusReporter_PrintTable(t *testing.T) {
	tracker := makeTrackerWithProcesses(t)
	var buf bytes.Buffer
	reporter := NewStatusReporter(tracker, &buf)

	reporter.PrintTable()

	output := buf.String()
	if !strings.Contains(output, "NAME") {
		t.Error("expected table header NAME")
	}
	if !strings.Contains(output, "web") {
		t.Error("expected process name 'web' in table output")
	}
	if !strings.Contains(output, "worker") {
		t.Error("expected process name 'worker' in table output")
	}
}

func TestStatusReporter_NilWriter_UsesStdout(t *testing.T) {
	tracker := NewHealthTracker()
	reporter := NewStatusReporter(tracker, nil)
	if reporter.writer == nil {
		t.Error("expected non-nil writer when nil is passed")
	}
}
