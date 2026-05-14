package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessPriorityReporter_PrintTable_NoEntries(t *testing.T) {
	s := NewProcessPriorityStore()
	var buf bytes.Buffer
	r := NewProcessPriorityReporter(s, &buf)
	r.PrintTable()
	if !strings.Contains(buf.String(), "no process priorities") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestProcessPriorityReporter_PrintTable_ContainsProcess(t *testing.T) {
	s := NewProcessPriorityStore()
	s.Set("api", PriorityHigh)
	s.Set("worker", PriorityLow)
	var buf bytes.Buffer
	r := NewProcessPriorityReporter(s, &buf)
	r.PrintTable()
	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Errorf("expected 'api' in output: %s", out)
	}
	if !strings.Contains(out, "high") {
		t.Errorf("expected 'high' in output: %s", out)
	}
	if !strings.Contains(out, "worker") {
		t.Errorf("expected 'worker' in output: %s", out)
	}
}

func TestProcessPriorityReporter_PrintJSON_Valid(t *testing.T) {
	s := NewProcessPriorityStore()
	s.Set("db", PriorityNormal)
	var buf bytes.Buffer
	r := NewProcessPriorityReporter(s, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}
	var rows []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["process"] != "db" {
		t.Errorf("expected process 'db', got %s", rows[0]["process"])
	}
	if rows[0]["priority"] != "normal" {
		t.Errorf("expected priority 'normal', got %s", rows[0]["priority"])
	}
}

func TestProcessPriorityReporter_NilWriter_UsesStdout(t *testing.T) {
	s := NewProcessPriorityStore()
	r := NewProcessPriorityReporter(s, nil)
	if r.writer == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
