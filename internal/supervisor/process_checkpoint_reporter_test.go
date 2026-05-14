package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessCheckpointReporter_PrintTable_NoEntries(t *testing.T) {
	log := NewProcessCheckpointLog()
	var buf bytes.Buffer
	r := NewProcessCheckpointReporter(log, &buf)
	r.PrintTable()

	if !strings.Contains(buf.String(), "PROCESS") {
		t.Errorf("expected header in output, got: %s", buf.String())
	}
}

func TestProcessCheckpointReporter_PrintTable_ContainsProcess(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("api", CheckpointStarted, nil)
	log.Record("worker", CheckpointReady, nil)

	var buf bytes.Buffer
	r := NewProcessCheckpointReporter(log, &buf)
	r.PrintTable()

	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Errorf("expected 'api' in output: %s", out)
	}
	if !strings.Contains(out, "worker") {
		t.Errorf("expected 'worker' in output: %s", out)
	}
	if !strings.Contains(out, "started") {
		t.Errorf("expected 'started' kind in output: %s", out)
	}
}

func TestProcessCheckpointReporter_PrintJSON_Valid(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("svc", CheckpointFailed, map[string]string{"reason": "oom"})

	var buf bytes.Buffer
	r := NewProcessCheckpointReporter(log, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []ProcessCheckpoint
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].Process != "svc" {
		t.Errorf("expected process 'svc', got %s", result[0].Process)
	}
}

func TestProcessCheckpointReporter_NilWriter_UsesStdout(t *testing.T) {
	log := NewProcessCheckpointLog()
	r := NewProcessCheckpointReporter(log, nil)
	if r.writer == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
