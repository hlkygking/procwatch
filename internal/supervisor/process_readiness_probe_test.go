package supervisor

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProcessReadinessProbe_HTTPSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	probe := NewProcessReadinessProbe("web", ReadinessProbeConfig{
		Kind:   ProbeHTTP,
		Target: ts.URL,
	})
	res := probe.Probe(context.Background())
	if !res.Ready {
		t.Fatalf("expected ready, got err: %v", res.Err)
	}
	if res.Process != "web" {
		t.Errorf("expected process 'web', got %q", res.Process)
	}
}

func TestProcessReadinessProbe_HTTPFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	probe := NewProcessReadinessProbe("api", ReadinessProbeConfig{
		Kind:   ProbeHTTP,
		Target: ts.URL,
	})
	res := probe.Probe(context.Background())
	if res.Ready {
		t.Fatal("expected not ready for 503 response")
	}
	if res.Err == nil {
		t.Error("expected non-nil error for 503 response")
	}
}

func TestProcessReadinessProbe_TCPSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	probe := NewProcessReadinessProbe("db", ReadinessProbeConfig{
		Kind:   ProbeTCP,
		Target: ln.Addr().String(),
	})
	res := probe.Probe(context.Background())
	if !res.Ready {
		t.Fatalf("expected TCP probe ready, got: %v", res.Err)
	}
}

func TestProcessReadinessProbe_TCPFailure(t *testing.T) {
	probe := NewProcessReadinessProbe("db", ReadinessProbeConfig{
		Kind:    ProbeTCP,
		Target:  "127.0.0.1:19999",
		Timeout: 200 * time.Millisecond,
	})
	res := probe.Probe(context.Background())
	if res.Ready {
		t.Fatal("expected TCP probe to fail for closed port")
	}
}

func TestProcessReadinessProbe_DefaultsApplied(t *testing.T) {
	probe := NewProcessReadinessProbe("svc", ReadinessProbeConfig{})
	if probe.cfg.Timeout != 2*time.Second {
		t.Errorf("expected default timeout 2s, got %v", probe.cfg.Timeout)
	}
	if probe.cfg.Interval != 5*time.Second {
		t.Errorf("expected default interval 5s, got %v", probe.cfg.Interval)
	}
	if probe.cfg.SuccessThreshold != 1 {
		t.Errorf("expected success threshold 1, got %d", probe.cfg.SuccessThreshold)
	}
	if probe.cfg.FailureThreshold != 3 {
		t.Errorf("expected failure threshold 3, got %d", probe.cfg.FailureThreshold)
	}
}

func TestProcessReadinessProbe_ContextCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	probe := NewProcessReadinessProbe("slow", ReadinessProbeConfig{
		Kind:   ProbeHTTP,
		Target: ts.URL,
	})
	res := probe.Probe(ctx)
	if res.Ready {
		t.Fatal("expected probe to fail on context cancellation")
	}
}
