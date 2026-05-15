package supervisor

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestProcessSnapshotStore_RecordAndAll(t *testing.T) {
	store := NewProcessSnapshotStore()
	store.Record(ProcessSnapshot{Name: "web", Status: "running", Restarts: 1})
	store.Record(ProcessSnapshot{Name: "worker", Status: "stopped", Restarts: 0})

	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(all))
	}
}

func TestProcessSnapshotStore_TimestampAutoSet(t *testing.T) {
	store := NewProcessSnapshotStore()
	before := time.Now()
	store.Record(ProcessSnapshot{Name: "svc", Status: "running"})
	after := time.Now()

	snap, ok := store.Latest("svc")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if snap.TakenAt.Before(before) || snap.TakenAt.After(after) {
		t.Errorf("timestamp out of range: %v", snap.TakenAt)
	}
}

func TestProcessSnapshotStore_ForProcess_Filtered(t *testing.T) {
	store := NewProcessSnapshotStore()
	store.Record(ProcessSnapshot{Name: "web", Status: "running"})
	store.Record(ProcessSnapshot{Name: "web", Status: "restarting"})
	store.Record(ProcessSnapshot{Name: "db", Status: "running"})

	snaps := store.ForProcess("web")
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots for web, got %d", len(snaps))
	}
}

func TestProcessSnapshotStore_ForProcess_NoMatch(t *testing.T) {
	store := NewProcessSnapshotStore()
	snaps := store.ForProcess("missing")
	if len(snaps) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(snaps))
	}
}

func TestProcessSnapshotStore_Latest_ReturnsNewest(t *testing.T) {
	store := NewProcessSnapshotStore()
	store.Record(ProcessSnapshot{Name: "api", Status: "running", Restarts: 0})
	store.Record(ProcessSnapshot{Name: "api", Status: "restarting", Restarts: 1})

	snap, ok := store.Latest("api")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if snap.Status != "restarting" {
		t.Errorf("expected latest status 'restarting', got %q", snap.Status)
	}
}

func TestProcessSnapshotReporter_PrintTable_ContainsProcess(t *testing.T) {
	store := NewProcessSnapshotStore()
	store.Record(ProcessSnapshot{Name: "web", Status: "running", Restarts: 3, Uptime: 120.5})

	var buf bytes.Buffer
	reporter := NewProcessSnapshotReporter(store, &buf)
	reporter.PrintTable([]string{"web"})

	out := buf.String()
	if !strings.Contains(out, "web") {
		t.Errorf("expected process name in output, got: %s", out)
	}
	if !strings.Contains(out, "running") {
		t.Errorf("expected status in output, got: %s", out)
	}
}

func TestProcessSnapshotReporter_PrintJSON_Valid(t *testing.T) {
	store := NewProcessSnapshotStore()
	store.Record(ProcessSnapshot{Name: "worker", Status: "stopped", Restarts: 2})

	var buf bytes.Buffer
	reporter := NewProcessSnapshotReporter(store, &buf)
	if err := reporter.PrintJSON([]string{"worker"}); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "worker") {
		t.Errorf("expected process name in JSON output")
	}
}

func TestProcessSnapshotReporter_NilWriter_UsesStdout(t *testing.T) {
	store := NewProcessSnapshotStore()
	reporter := NewProcessSnapshotReporter(store, nil)
	if reporter.writer == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
