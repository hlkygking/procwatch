package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessPauseReporter_PrintTable_NoEntries(t *testing.T) {
	s := NewProcessPauseStore()
	var buf bytes.Buffer
	r := NewProcessPauseReporter(s, &buf)
	r.PrintTable()
	if !strings.Contains(buf.String(), "PROCESS") {
		t.Error("expected header row in table output")
	}
}

func TestProcessPauseReporter_PrintTable_ContainsProcess(t *testing.T) {
	s := NewProcessPauseStore()
	s.Pause("worker", "overload")
	s.Resume("worker", "recovered")

	var buf bytes.Buffer
	r := NewProcessPauseReporter(s, &buf)
	r.PrintTable()

	out := buf.String()
	if !strings.Contains(out, "worker") {
		t.Error("expected process name in table")
	}
	if !strings.Contains(out, "paused") {
		t.Error("expected 'paused' state in table")
	}
	if !strings.Contains(out, "active") {
		t.Error("expected 'active' state in table")
	}
	if !strings.Contains(out, "overload") {
		t.Error("expected reason in table")
	}
}

func TestProcessPauseReporter_PrintJSON_Valid(t *testing.T) {
	s := NewProcessPauseStore()
	s.Pause("api", "deploy")

	var buf bytes.Buffer
	r := NewProcessPauseReporter(s, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0]["process"] != "api" {
		t.Errorf("expected process=api, got %v", result[0]["process"])
	}
	if result[0]["state"] != "paused" {
		t.Errorf("expected state=paused, got %v", result[0]["state"])
	}
}

func TestProcessPauseReporter_NilWriter_UsesStdout(t *testing.T) {
	s := NewProcessPauseStore()
	r := NewProcessPauseReporter(s, nil)
	if r.writer == nil {
		t.Error("expected writer to fall back to stdout")
	}
}
