package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessRunbook_AddAndAll(t *testing.T) {
	rb := NewProcessRunbook()
	rb.Add("web", "step-1", "check logs")
	rb.Add("worker", "step-1", "drain queue")

	all := rb.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestProcessRunbook_TimestampSet(t *testing.T) {
	rb := NewProcessRunbook()
	rb.Add("web", "step-1", "restart service")

	all := rb.All()
	if all[0].CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestProcessRunbook_ForProcess_Filtered(t *testing.T) {
	rb := NewProcessRunbook()
	rb.Add("web", "step-1", "note a")
	rb.Add("worker", "step-1", "note b")
	rb.Add("web", "step-2", "note c")

	results := rb.ForProcess("web")
	if len(results) != 2 {
		t.Fatalf("expected 2 entries for 'web', got %d", len(results))
	}
	for _, e := range results {
		if e.Process != "web" {
			t.Errorf("unexpected process %q in filtered results", e.Process)
		}
	}
}

func TestProcessRunbook_ForProcess_NoMatch(t *testing.T) {
	rb := NewProcessRunbook()
	rb.Add("web", "step-1", "note")

	results := rb.ForProcess("unknown")
	if len(results) != 0 {
		t.Errorf("expected 0 entries, got %d", len(results))
	}
}

func TestProcessRunbook_PrintTable_NoEntries(t *testing.T) {
	rb := NewProcessRunbook()
	var buf bytes.Buffer
	rb.PrintTable(&buf)
	if !strings.Contains(buf.String(), "no runbook entries") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestProcessRunbook_PrintTable_ContainsProcess(t *testing.T) {
	rb := NewProcessRunbook()
	rb.Add("web", "step-1", "check health endpoint")
	var buf bytes.Buffer
	rb.PrintTable(&buf)
	if !strings.Contains(buf.String(), "web") {
		t.Errorf("expected 'web' in table output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "check health endpoint") {
		t.Errorf("expected note in table output, got: %s", buf.String())
	}
}

func TestProcessRunbook_PrintJSON_Valid(t *testing.T) {
	rb := NewProcessRunbook()
	rb.Add("api", "step-1", "rotate credentials")
	var buf bytes.Buffer
	rb.PrintJSON(&buf)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSON line, got %d", len(lines))
	}
	var entry RunbookEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Process != "api" || entry.Step != "step-1" {
		t.Errorf("unexpected entry: %+v", entry)
	}
}

func TestProcessRunbook_NilWriter_UsesStdout(t *testing.T) {
	rb := NewProcessRunbook()
	// Should not panic with nil writer
	rb.PrintTable(nil)
	rb.PrintJSON(nil)
}
