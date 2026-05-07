package supervisor

import (
	"testing"
	"time"
)

func TestProcessMetrics_RecordStartAndStop(t *testing.T) {
	m := &ProcessMetrics{Name: "web"}
	m.RecordStart()
	if !m.Running {
		t.Fatal("expected Running to be true after RecordStart")
	}
	time.Sleep(10 * time.Millisecond)
	m.RecordStop(0)
	if m.Running {
		t.Fatal("expected Running to be false after RecordStop")
	}
	if m.TotalUptime < 10*time.Millisecond {
		t.Errorf("expected TotalUptime >= 10ms, got %v", m.TotalUptime)
	}
	if m.LastExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", m.LastExitCode)
	}
}

func TestProcessMetrics_RecordRestart(t *testing.T) {
	m := &ProcessMetrics{Name: "worker"}
	m.RecordRestart()
	m.RecordRestart()
	snap := m.Snapshot()
	if snap.RestartCount != 2 {
		t.Errorf("expected RestartCount 2, got %d", snap.RestartCount)
	}
}

func TestProcessMetrics_Snapshot_UptimeWhileRunning(t *testing.T) {
	m := &ProcessMetrics{Name: "api"}
	m.RecordStart()
	time.Sleep(15 * time.Millisecond)
	snap := m.Snapshot()
	if snap.TotalUptime < 15*time.Millisecond {
		t.Errorf("expected live uptime >= 15ms, got %v", snap.TotalUptime)
	}
	if !snap.Running {
		t.Error("expected snapshot to show Running=true")
	}
}

func TestProcessMetrics_NonZeroExitCode(t *testing.T) {
	m := &ProcessMetrics{Name: "job"}
	m.RecordStart()
	m.RecordStop(2)
	if m.LastExitCode != 2 {
		t.Errorf("expected exit code 2, got %d", m.LastExitCode)
	}
}

func TestProcessMetricsStore_GetOrCreate_Idempotent(t *testing.T) {
	store := NewProcessMetricsStore()
	a := store.GetOrCreate("svc")
	b := store.GetOrCreate("svc")
	if a != b {
		t.Error("expected same pointer for same process name")
	}
}

func TestProcessMetricsStore_Snapshots(t *testing.T) {
	store := NewProcessMetricsStore()
	store.GetOrCreate("alpha").RecordStart()
	store.GetOrCreate("beta").RecordStart()
	snaps := store.Snapshots()
	if len(snaps) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestProcessMetricsStore_Get_Missing(t *testing.T) {
	store := NewProcessMetricsStore()
	_, ok := store.Get("ghost")
	if ok {
		t.Error("expected Get to return false for unknown process")
	}
}
