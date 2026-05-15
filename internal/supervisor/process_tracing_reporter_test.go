package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestProcessTracingReporter_PrintTable_NoEntries(t *testing.T) {
	s := NewProcessTracingStore()
	var buf bytes.Buffer
	r := NewProcessTracingReporter(s, &buf)
	r.PrintTable()
	if !strings.Contains(buf.String(), "PROCESS") {
		t.Fatal("expected header in table output")
	}
}

func TestProcessTracingReporter_PrintTable_ContainsProcess(t *testing.T) {
	s := NewProcessTracingStore()
	s.Record(TraceEvent{
		Process:   "api",
		Kind:      "start",
		StartedAt: time.Now().Add(-200 * time.Millisecond),
	})
	var buf bytes.Buffer
	r := NewProcessTracingReporter(s, &buf)
	r.PrintTable()
	if !strings.Contains(buf.String(), "api") {
		t.Fatalf("expected 'api' in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "start") {
		t.Fatalf("expected 'start' kind in output, got: %s", buf.String())
	}
}

func TestProcessTracingReporter_PrintJSON_Valid(t *testing.T) {
	s := NewProcessTracingStore()
	s.Record(TraceEvent{
		Process:   "worker",
		Kind:      "stop",
		StartedAt: time.Now().Add(-50 * time.Millisecond),
	})
	var buf bytes.Buffer
	r := NewProcessTracingReporter(s, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var events []TraceEvent
	if err := json.Unmarshal(buf.Bytes(), &events); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(events) != 1 || events[0].Process != "worker" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestProcessTracingReporter_NilWriter_UsesStdout(t *testing.T) {
	s := NewProcessTracingStore()
	r := NewProcessTracingReporter(s, nil)
	if r.writer == nil {
		t.Fatal("expected non-nil writer when nil passed")
	}
}
