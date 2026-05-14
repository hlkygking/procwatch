package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessNotificationReporter_PrintTable_NoEntries(t *testing.T) {
	bus := NewProcessNotificationBus()
	var buf bytes.Buffer
	r := NewProcessNotificationReporter(bus, &buf)
	r.PrintTable()

	if !strings.Contains(buf.String(), "PROCESS") {
		t.Error("expected header row in output")
	}
}

func TestProcessNotificationReporter_PrintTable_ContainsProcess(t *testing.T) {
	bus := NewProcessNotificationBus()
	bus.Publish("api", NotifyStarted, "launched")

	var buf bytes.Buffer
	r := NewProcessNotificationReporter(bus, &buf)
	r.PrintTable()

	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Error("expected process name 'api' in table output")
	}
	if !strings.Contains(out, "started") {
		t.Error("expected kind 'started' in table output")
	}
	if !strings.Contains(out, "launched") {
		t.Error("expected message 'launched' in table output")
	}
}

func TestProcessNotificationReporter_PrintJSON_Valid(t *testing.T) {
	bus := NewProcessNotificationBus()
	bus.Publish("worker", NotifyRestart, "retry")

	var buf bytes.Buffer
	r := NewProcessNotificationReporter(bus, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	var entries []ProcessNotification
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Process != "worker" {
		t.Errorf("unexpected process: %s", entries[0].Process)
	}
}

func TestProcessNotificationReporter_NilWriter_UsesStdout(t *testing.T) {
	bus := NewProcessNotificationBus()
	r := NewProcessNotificationReporter(bus, nil)
	if r.writer == nil {
		t.Error("expected non-nil writer when nil is passed")
	}
}
