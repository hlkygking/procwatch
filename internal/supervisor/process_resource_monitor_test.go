package supervisor

import (
	"os"
	"testing"
	"time"
)

func TestProcessResourceMonitor_TrackAndSample(t *testing.T) {
	m := NewProcessResourceMonitor(50 * time.Millisecond)
	pid := os.Getpid()
	m.Track("self", pid)

	m.Start()
	time.Sleep(120 * time.Millisecond)
	m.Stop()

	s := m.Sample("self")
	if s == nil {
		// On non-Linux systems /proc is unavailable; just verify no panic.
		t.Log("no sample available (non-Linux environment)")
		return
	}
	if s.PID != pid {
		t.Errorf("expected PID %d, got %d", pid, s.PID)
	}
	if s.Name != "self" {
		t.Errorf("expected name 'self', got %q", s.Name)
	}
	if s.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestProcessResourceMonitor_SampleUnknown(t *testing.T) {
	m := NewProcessResourceMonitor(time.Second)
	s := m.Sample("nonexistent")
	if s != nil {
		t.Errorf("expected nil sample for unknown process, got %+v", s)
	}
}

func TestProcessResourceMonitor_Untrack(t *testing.T) {
	m := NewProcessResourceMonitor(50 * time.Millisecond)
	m.Track("svc", os.Getpid())
	m.Untrack("svc")

	m.Start()
	time.Sleep(120 * time.Millisecond)
	m.Stop()

	if s := m.Sample("svc"); s != nil {
		t.Error("expected no sample after untrack")
	}
}

func TestProcessResourceMonitor_DefaultInterval(t *testing.T) {
	m := NewProcessResourceMonitor(0)
	if m.interval != 5*time.Second {
		t.Errorf("expected default interval 5s, got %v", m.interval)
	}
}

func TestProcessResourceMonitor_SampleIsCopy(t *testing.T) {
	m := NewProcessResourceMonitor(50 * time.Millisecond)
	m.Track("self", os.Getpid())

	m.Start()
	time.Sleep(120 * time.Millisecond)
	m.Stop()

	s1 := m.Sample("self")
	if s1 == nil {
		t.Skip("no sample available")
	}
	s1.RSSBytes = -999

	s2 := m.Sample("self")
	if s2 == nil {
		t.Skip("no second sample available")
	}
	if s2.RSSBytes == -999 {
		t.Error("sample should be a copy, not a shared pointer")
	}
}

func TestItoa(t *testing.T) {
	cases := []struct {
		in  int
		out string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{1234, "1234"},
	}
	for _, c := range cases {
		if got := itoa(c.in); got != c.out {
			t.Errorf("itoa(%d) = %q, want %q", c.in, got, c.out)
		}
	}
}
