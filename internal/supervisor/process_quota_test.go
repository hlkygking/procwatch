package supervisor

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestProcessQuotaStore_SetAndGet(t *testing.T) {
	s := NewProcessQuotaStore()
	q := ProcessQuota{MaxRestarts: 5, MaxUptime: time.Minute}
	s.Set("web", q)
	got, ok := s.Get("web")
	if !ok {
		t.Fatal("expected quota to be found")
	}
	if got.MaxRestarts != 5 {
		t.Errorf("expected MaxRestarts=5, got %d", got.MaxRestarts)
	}
}

func TestProcessQuotaStore_GetMissing(t *testing.T) {
	s := NewProcessQuotaStore()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected no quota for unknown process")
	}
}

func TestProcessQuotaStore_CheckRestarts_NoViolation(t *testing.T) {
	s := NewProcessQuotaStore()
	s.Set("web", ProcessQuota{MaxRestarts: 10})
	v := s.CheckRestarts("web", 3)
	if v != nil {
		t.Errorf("expected no violation, got: %v", v)
	}
}

func TestProcessQuotaStore_CheckRestarts_Violation(t *testing.T) {
	s := NewProcessQuotaStore()
	s.Set("web", ProcessQuota{MaxRestarts: 3})
	v := s.CheckRestarts("web", 5)
	if v == nil {
		t.Fatal("expected a violation")
	}
	if v.Process != "web" {
		t.Errorf("expected process=web, got %s", v.Process)
	}
	if !strings.Contains(v.Reason, "5") {
		t.Errorf("expected reason to mention count, got: %s", v.Reason)
	}
}

func TestProcessQuotaStore_CheckUptime_Violation(t *testing.T) {
	s := NewProcessQuotaStore()
	s.Set("worker", ProcessQuota{MaxUptime: 10 * time.Second})
	v := s.CheckUptime("worker", 30*time.Second)
	if v == nil {
		t.Fatal("expected uptime violation")
	}
	if !strings.Contains(v.Reason, "exceeds") {
		t.Errorf("unexpected reason: %s", v.Reason)
	}
}

func TestProcessQuotaStore_CheckUptime_NoQuota(t *testing.T) {
	s := NewProcessQuotaStore()
	v := s.CheckUptime("unknown", time.Hour)
	if v != nil {
		t.Errorf("expected nil for unknown process, got %v", v)
	}
}

func TestProcessQuotaStore_ViolationsFor(t *testing.T) {
	s := NewProcessQuotaStore()
	s.Set("a", ProcessQuota{MaxRestarts: 1})
	s.Set("b", ProcessQuota{MaxRestarts: 1})
	s.CheckRestarts("a", 5)
	s.CheckRestarts("b", 5)
	s.CheckRestarts("a", 10)
	vs := s.ViolationsFor("a")
	if len(vs) != 2 {
		t.Errorf("expected 2 violations for 'a', got %d", len(vs))
	}
	for _, v := range vs {
		if v.Process != "a" {
			t.Errorf("unexpected process in filter: %s", v.Process)
		}
	}
}

func TestProcessQuotaReporter_PrintTable(t *testing.T) {
	s := NewProcessQuotaStore()
	s.Set("svc", ProcessQuota{MaxRestarts: 2})
	s.CheckRestarts("svc", 4)
	var buf bytes.Buffer
	r := NewProcessQuotaReporter(s, &buf)
	r.PrintTable()
	out := buf.String()
	if !strings.Contains(out, "svc") {
		t.Errorf("expected 'svc' in output, got: %s", out)
	}
}

func TestProcessQuotaReporter_PrintJSON(t *testing.T) {
	s := NewProcessQuotaStore()
	s.Set("api", ProcessQuota{MaxRestarts: 1})
	s.CheckRestarts("api", 3)
	var buf bytes.Buffer
	r := NewProcessQuotaReporter(s, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "api") {
		t.Errorf("expected 'api' in JSON output")
	}
}

func TestProcessQuotaReporter_NilWriter_UsesStdout(t *testing.T) {
	s := NewProcessQuotaStore()
	r := NewProcessQuotaReporter(s, nil)
	if r.writer == nil {
		t.Error("expected non-nil writer fallback")
	}
}

func TestProcessQuotaViolation_String(t *testing.T) {
	v := ProcessQuotaViolation{Process: "p", Reason: "too many restarts", At: time.Now()}
	s := v.String()
	if !strings.Contains(s, "p") || !strings.Contains(s, "too many restarts") {
		t.Errorf("unexpected String() output: %s", s)
	}
}
