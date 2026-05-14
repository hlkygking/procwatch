package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessLabelReporter_PrintTable_NoLabels(t *testing.T) {
	store := NewProcessLabelStore()
	var buf bytes.Buffer
	r := NewProcessLabelReporter(store, &buf)
	r.PrintTable()
	if !strings.Contains(buf.String(), "no labels defined") {
		t.Fatalf("expected empty message, got: %q", buf.String())
	}
}

func TestProcessLabelReporter_PrintTable_ContainsProcess(t *testing.T) {
	store := NewProcessLabelStore()
	ls := store.GetOrCreate("web")
	ls.Set("env", "prod")
	var buf bytes.Buffer
	r := NewProcessLabelReporter(store, &buf)
	r.PrintTable()
	out := buf.String()
	if !strings.Contains(out, "web") {
		t.Fatalf("expected process name in output: %q", out)
	}
	if !strings.Contains(out, "env=prod") {
		t.Fatalf("expected label pair in output: %q", out)
	}
}

func TestProcessLabelReporter_PrintJSON_Valid(t *testing.T) {
	store := NewProcessLabelStore()
	ls := store.GetOrCreate("worker")
	ls.Set("tier", "backend")
	var buf bytes.Buffer
	r := NewProcessLabelReporter(store, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}
	var out []struct {
		Process string            `json:"process"`
		Labels  map[string]string `json:"labels"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 1 || out[0].Process != "worker" {
		t.Fatalf("unexpected JSON output: %+v", out)
	}
	if out[0].Labels["tier"] != "backend" {
		t.Fatalf("expected tier=backend, got %v", out[0].Labels)
	}
}

func TestProcessLabelReporter_NilWriter_UsesStdout(t *testing.T) {
	store := NewProcessLabelStore()
	r := NewProcessLabelReporter(store, nil)
	if r.writer == nil {
		t.Fatal("expected non-nil writer when nil passed")
	}
}
