package supervisor

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestProcessHeartbeatStore_BeatAndStatus(t *testing.T) {
	s := NewProcessHeartbeatStore(5 * time.Second)
	s.Beat("alpha")
	st, ok := s.Status("alpha")
	if !ok {
		t.Fatal("expected status to exist after Beat")
	}
	if !st.Alive {
		t.Error("expected process to be alive after Beat")
	}
	if st.Missed != 0 {
		t.Errorf("expected 0 missed, got %d", st.Missed)
	}
}

func TestProcessHeartbeatStore_Status_Missing(t *testing.T) {
	s := NewProcessHeartbeatStore(5 * time.Second)
	_, ok := s.Status("ghost")
	if ok {
		t.Error("expected false for unseen process")
	}
}

func TestProcessHeartbeatStore_CheckMarksDead(t *testing.T) {
	s := NewProcessHeartbeatStore(1 * time.Millisecond)
	s.Beat("beta")
	time.Sleep(10 * time.Millisecond)
	s.Check()
	st, _ := s.Status("beta")
	if st.Alive {
		t.Error("expected process to be dead after timeout")
	}
	if st.Missed != 1 {
		t.Errorf("expected missed=1, got %d", st.Missed)
	}
}

func TestProcessHeartbeatStore_BeatResetsState(t *testing.T) {
	s := NewProcessHeartbeatStore(1 * time.Millisecond)
	s.Beat("gamma")
	time.Sleep(10 * time.Millisecond)
	s.Check()
	s.Beat("gamma")
	st, _ := s.Status("gamma")
	if !st.Alive {
		t.Error("expected alive after re-beat")
	}
	if st.Missed != 0 {
		t.Errorf("expected missed reset to 0, got %d", st.Missed)
	}
}

func TestProcessHeartbeatStore_Remove(t *testing.T) {
	s := NewProcessHeartbeatStore(5 * time.Second)
	s.Beat("delta")
	s.Remove("delta")
	_, ok := s.Status("delta")
	if ok {
		t.Error("expected process to be removed")
	}
}

func TestProcessHeartbeatStore_DefaultTimeout(t *testing.T) {
	s := NewProcessHeartbeatStore(0)
	if s.timeout != 30*time.Second {
		t.Errorf("expected default 30s, got %v", s.timeout)
	}
}

func TestProcessHeartbeatReporter_PrintTable(t *testing.T) {
	s := NewProcessHeartbeatStore(5 * time.Second)
	s.Beat("web")
	var buf bytes.Buffer
	r := NewProcessHeartbeatReporter(s, &buf)
	r.PrintTable()
	out := buf.String()
	if !strings.Contains(out, "web") {
		t.Errorf("expected 'web' in table output, got: %s", out)
	}
	if !strings.Contains(out, "yes") {
		t.Errorf("expected 'yes' alive status, got: %s", out)
	}
}

func TestProcessHeartbeatReporter_PrintJSON(t *testing.T) {
	s := NewProcessHeartbeatStore(5 * time.Second)
	s.Beat("api")
	var buf bytes.Buffer
	r := NewProcessHeartbeatReporter(s, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "api") {
		t.Errorf("expected 'api' in JSON output")
	}
}

func TestProcessHeartbeatReporter_NilWriter_UsesStdout(t *testing.T) {
	s := NewProcessHeartbeatStore(5 * time.Second)
	r := NewProcessHeartbeatReporter(s, nil)
	if r.writer == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
