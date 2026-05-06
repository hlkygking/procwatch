package supervisor

import (
	"testing"
	"time"
)

func TestHealthStatus_String(t *testing.T) {
	cases := []struct {
		status HealthStatus
		want   string
	}{
		{StatusHealthy, "healthy"},
		{StatusUnhealthy, "unhealthy"},
		{StatusStarting, "starting"},
		{StatusUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("HealthStatus(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestHealthTracker_RecordStart(t *testing.T) {
	ht := NewHealthTracker()
	before := time.Now()
	ht.RecordStart("svc-a")

	rec, ok := ht.Get("svc-a")
	if !ok {
		t.Fatal("expected record for svc-a")
	}
	if rec.Status != StatusStarting {
		t.Errorf("expected StatusStarting, got %s", rec.Status)
	}
	if rec.LastStart.Before(before) {
		t.Error("LastStart should be >= time before RecordStart call")
	}
}

func TestHealthTracker_RecordExit(t *testing.T) {
	ht := NewHealthTracker()
	ht.RecordStart("svc-b")
	ht.RecordExit("svc-b", 1)

	rec, ok := ht.Get("svc-b")
	if !ok {
		t.Fatal("expected record for svc-b")
	}
	if rec.Status != StatusUnhealthy {
		t.Errorf("expected StatusUnhealthy, got %s", rec.Status)
	}
	if rec.Restarts != 1 {
		t.Errorf("expected 1 restart, got %d", rec.Restarts)
	}
	if rec.ExitCode == nil || *rec.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %v", rec.ExitCode)
	}
	if rec.LastExit == nil {
		t.Error("expected LastExit to be set")
	}
}

func TestHealthTracker_SetStatus(t *testing.T) {
	ht := NewHealthTracker()
	ht.SetStatus("svc-c", StatusHealthy)

	rec, ok := ht.Get("svc-c")
	if !ok {
		t.Fatal("expected record for svc-c")
	}
	if rec.Status != StatusHealthy {
		t.Errorf("expected StatusHealthy, got %s", rec.Status)
	}
}

func TestHealthTracker_All(t *testing.T) {
	ht := NewHealthTracker()
	ht.RecordStart("svc-x")
	ht.RecordStart("svc-y")

	all := ht.All()
	if len(all) != 2 {
		t.Errorf("expected 2 records, got %d", len(all))
	}
}

func TestHealthTracker_GetMissing(t *testing.T) {
	ht := NewHealthTracker()
	_, ok := ht.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown process")
	}
}

func TestHealthTracker_MultipleRestarts(t *testing.T) {
	ht := NewHealthTracker()
	for i := 0; i < 5; i++ {
		ht.RecordStart("svc-d")
		ht.RecordExit("svc-d", 1)
	}
	rec, _ := ht.Get("svc-d")
	if rec.Restarts != 5 {
		t.Errorf("expected 5 restarts, got %d", rec.Restarts)
	}
}
