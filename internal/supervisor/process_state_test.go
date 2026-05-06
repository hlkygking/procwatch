package supervisor

import (
	"testing"
	"time"
)

func TestProcessState_InitialState(t *testing.T) {
	ps := NewProcessState("worker")
	snap := ps.Snapshot()
	if snap.Name != "worker" {
		t.Errorf("expected name 'worker', got %q", snap.Name)
	}
	if snap.Running {
		t.Error("expected not running initially")
	}
	if snap.Restarts != 0 {
		t.Errorf("expected 0 restarts, got %d", snap.Restarts)
	}
}

func TestProcessState_RecordStart(t *testing.T) {
	ps := NewProcessState("worker")
	before := time.Now()
	ps.RecordStart(1234)
	snap := ps.Snapshot()
	if !snap.Running {
		t.Error("expected running after RecordStart")
	}
	if snap.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", snap.PID)
	}
	if snap.StartedAt.Before(before) {
		t.Error("StartedAt should be after test start")
	}
}

func TestProcessState_RecordStop_NonZeroExit(t *testing.T) {
	ps := NewProcessState("worker")
	ps.RecordStart(42)
	ps.RecordStop(1)
	snap := ps.Snapshot()
	if snap.Running {
		t.Error("expected not running after RecordStop")
	}
	if snap.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", snap.ExitCode)
	}
	if snap.Restarts != 1 {
		t.Errorf("expected 1 restart, got %d", snap.Restarts)
	}
}

func TestProcessState_RecordStop_ZeroExit_NoRestart(t *testing.T) {
	ps := NewProcessState("worker")
	ps.RecordStart(99)
	ps.RecordStop(0)
	snap := ps.Snapshot()
	if snap.Restarts != 0 {
		t.Errorf("expected 0 restarts on clean exit, got %d", snap.Restarts)
	}
}

func TestProcessStateSnapshot_Uptime_Running(t *testing.T) {
	ps := NewProcessState("svc")
	ps.RecordStart(5)
	time.Sleep(10 * time.Millisecond)
	snap := ps.Snapshot()
	if snap.Uptime() < 10*time.Millisecond {
		t.Error("expected uptime >= 10ms for running process")
	}
}

func TestProcessStateSnapshot_Uptime_Stopped(t *testing.T) {
	ps := NewProcessState("svc")
	ps.RecordStart(5)
	time.Sleep(10 * time.Millisecond)
	ps.RecordStop(0)
	snap := ps.Snapshot()
	uptime := snap.Uptime()
	if uptime < 10*time.Millisecond {
		t.Errorf("expected uptime >= 10ms for stopped process, got %v", uptime)
	}
}

func TestProcessStateSnapshot_Uptime_NeverStarted(t *testing.T) {
	ps := NewProcessState("svc")
	snap := ps.Snapshot()
	if snap.Uptime() != 0 {
		t.Error("expected zero uptime for never-started process")
	}
}
