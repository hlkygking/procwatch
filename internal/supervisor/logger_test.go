package supervisor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, "myproc")
	l.Info("started", nil)

	var entry LogEntry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}
	if entry.Level != LogLevelInfo {
		t.Errorf("expected level info, got %s", entry.Level)
	}
	if entry.Message != "started" {
		t.Errorf("expected message 'started', got %s", entry.Message)
	}
	if entry.Process != "myproc" {
		t.Errorf("expected process 'myproc', got %s", entry.Process)
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, "svc")
	l.Warn("restarting", map[string]any{"attempt": 3})

	var entry LogEntry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Fields["attempt"] == nil {
		t.Error("expected 'attempt' field")
	}
}

func TestLogger_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, "p")
	l.Info("one", nil)
	l.Error("two", nil)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}
}

func TestLogger_NilWriter_UsesStdout(t *testing.T) {
	// Should not panic when writer is nil (falls back to os.Stdout).
	l := NewLogger(nil, "x")
	if l.writer == nil {
		t.Error("expected non-nil writer")
	}
}

func TestLogger_TimestampPresent(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, "ts")
	l.Debug("check", nil)

	var entry LogEntry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}
