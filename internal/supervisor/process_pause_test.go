package supervisor

import (
	"testing"
	"time"
)

func TestProcessPauseStore_InitiallyActive(t *testing.T) {
	s := NewProcessPauseStore()
	if s.IsPaused("web") {
		t.Error("expected process to be active initially")
	}
	if s.State("web") != PauseStateActive {
		t.Errorf("expected active, got %s", s.State("web"))
	}
}

func TestProcessPauseStore_PauseAndResume(t *testing.T) {
	s := NewProcessPauseStore()
	s.Pause("web", "maintenance")
	if !s.IsPaused("web") {
		t.Error("expected process to be paused")
	}
	s.Resume("web", "done")
	if s.IsPaused("web") {
		t.Error("expected process to be active after resume")
	}
}

func TestProcessPauseStore_HistoryRecorded(t *testing.T) {
	s := NewProcessPauseStore()
	s.Pause("api", "deploy")
	s.Resume("api", "ready")
	h := s.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(h))
	}
	if h[0].State != PauseStatePaused {
		t.Errorf("expected first entry paused, got %s", h[0].State)
	}
	if h[1].State != PauseStateActive {
		t.Errorf("expected second entry active, got %s", h[1].State)
	}
}

func TestProcessPauseStore_ForProcess_Filtered(t *testing.T) {
	s := NewProcessPauseStore()
	s.Pause("web", "r1")
	s.Pause("db", "r2")
	s.Resume("web", "r3")

	entries := s.ForProcess("web")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for web, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Process != "web" {
			t.Errorf("unexpected process in filtered result: %s", e.Process)
		}
	}
}

func TestProcessPauseStore_ForProcess_NoMatch(t *testing.T) {
	s := NewProcessPauseStore()
	s.Pause("db", "backup")
	if len(s.ForProcess("web")) != 0 {
		t.Error("expected no entries for untracked process")
	}
}

func TestProcessPauseStore_TimestampIsSet(t *testing.T) {
	before := time.Now()
	s := NewProcessPauseStore()
	s.Pause("svc", "test")
	after := time.Now()

	h := s.History()
	if len(h) == 0 {
		t.Fatal("expected history entry")
	}
	ts := h[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v out of expected range [%v, %v]", ts, before, after)
	}
}

func TestPauseState_String(t *testing.T) {
	if PauseStateActive.String() != "active" {
		t.Errorf("unexpected string for active: %s", PauseStateActive.String())
	}
	if PauseStatePaused.String() != "paused" {
		t.Errorf("unexpected string for paused: %s", PauseStatePaused.String())
	}
}
