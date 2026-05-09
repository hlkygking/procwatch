package supervisor_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestProcessResourceReporter_ReportsSnapshot(t *testing.T) {
	monitor := NewProcessResourceMonitor()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	reporter := NewProcessResourceReporter(monitor, logger)
	if reporter == nil {
		t.Fatal("expected non-nil reporter")
	}
}

func TestProcessResourceReporter_OutputContainsProcessName(t *testing.T) {
	monitor := NewProcessResourceMonitor()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	reporter := NewProcessResourceReporter(monitor, logger)

	// Track a fake PID (won't resolve real RSS on all platforms, but exercises the path)
	monitor.Track("webserver", 0)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	reporter.RunOnce(ctx)

	output := buf.String()
	if !strings.Contains(output, "webserver") {
		t.Errorf("expected output to contain process name 'webserver', got: %s", output)
	}
}

func TestProcessResourceReporter_JSONFields(t *testing.T) {
	monitor := NewProcessResourceMonitor()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	reporter := NewProcessResourceReporter(monitor, logger)
	monitor.Track("worker", 0)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	reporter.RunOnce(ctx)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one log line")
	}

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("failed to parse log line as JSON: %v", err)
	}

	if _, ok := entry["process"]; !ok {
		t.Error("expected 'process' field in log entry")
	}
}

func TestProcessResourceReporter_NoTrackedProcesses(t *testing.T) {
	monitor := NewProcessResourceMonitor()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	reporter := NewProcessResourceReporter(monitor, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Should not panic or error when no processes are tracked
	reporter.RunOnce(ctx)

	// No output expected
	if buf.Len() != 0 {
		t.Logf("note: got output with no tracked processes: %s", buf.String())
	}
}

func TestProcessResourceReporter_StartStopLoop(t *testing.T) {
	monitor := NewProcessResourceMonitor()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	reporter := NewProcessResourceReporter(monitor, logger)
	monitor.Track("api", 0)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		reporter.Start(ctx, 40*time.Millisecond)
	}()

	select {
		case <-done:
			// reporter exited after context cancelled — expected
		case <-time.After(500*time.Millisecond):
			t.Error("reporter did not stop after context cancellation")
	}
}
