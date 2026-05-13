package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestProcessAuditLog_RecordAndAll(t *testing.T) {
	var buf bytes.Buffer
	al := NewProcessAuditLog(&buf)

	al.Record(AuditEntry{Process: "web", Kind: AuditEventStart, Message: "started"})
	al.Record(AuditEntry{Process: "worker", Kind: AuditEventStop})

	entries := al.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Process != "web" || entries[0].Kind != AuditEventStart {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
}

func TestProcessAuditLog_TimestampAutoSet(t *testing.T) {
	var buf bytes.Buffer
	al := NewProcessAuditLog(&buf)
	before := time.Now().UTC()
	al.Record(AuditEntry{Process: "svc", Kind: AuditEventRestart})
	after := time.Now().UTC()

	e := al.All()[0]
	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", e.Timestamp, before, after)
	}
}

func TestProcessAuditLog_WritesJSONLines(t *testing.T) {
	var buf bytes.Buffer
	al := NewProcessAuditLog(&buf)
	al.Record(AuditEntry{Process: "api", Kind: AuditEventKill, Message: "oom"})

	line := strings.TrimSpace(buf.String())
	var parsed AuditEntry
	if err := json.Unmarshal([]byte(line), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v — raw: %s", err, line)
	}
	if parsed.Process != "api" || parsed.Kind != AuditEventKill {
		t.Errorf("unexpected parsed entry: %+v", parsed)
	}
}

func TestProcessAuditLog_FilterByProcess(t *testing.T) {
	var buf bytes.Buffer
	al := NewProcessAuditLog(&buf)
	al.Record(AuditEntry{Process: "web", Kind: AuditEventStart})
	al.Record(AuditEntry{Process: "db", Kind: AuditEventStart})
	al.Record(AuditEntry{Process: "web", Kind: AuditEventStop})

	results := al.FilterByProcess("web")
	if len(results) != 2 {
		t.Errorf("expected 2 entries for 'web', got %d", len(results))
	}
}

func TestProcessAuditLog_FilterByKind(t *testing.T) {
	var buf bytes.Buffer
	al := NewProcessAuditLog(&buf)
	al.Record(AuditEntry{Process: "web", Kind: AuditEventStart})
	al.Record(AuditEntry{Process: "db", Kind: AuditEventRestart, RestartNum: 1})
	al.Record(AuditEntry{Process: "cache", Kind: AuditEventRestart, RestartNum: 2})

	results := al.FilterByKind(AuditEventRestart)
	if len(results) != 2 {
		t.Errorf("expected 2 restart entries, got %d", len(results))
	}
}

func TestProcessAuditLog_NilWriter_UsesStdout(t *testing.T) {
	al := NewProcessAuditLog(nil)
	if al.w == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}

func TestProcessAuditLog_ExitCodeOmittedWhenNil(t *testing.T) {
	var buf bytes.Buffer
	al := NewProcessAuditLog(&buf)
	al.Record(AuditEntry{Process: "svc", Kind: AuditEventStart})

	line := strings.TrimSpace(buf.String())
	if strings.Contains(line, "exit_code") {
		t.Errorf("exit_code should be omitted when nil, got: %s", line)
	}
}
