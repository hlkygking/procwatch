package supervisor

import (
	"context"
	"testing"
	"time"
)

// mockLogger captures log calls for assertions.
type mockLogger struct {
	entries []map[string]interface{}
}

func (m *mockLogger) Info(msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["msg"] = msg
	m.entries = append(m.entries, fields)
}

func TestProcess_RunExitsCleanly(t *testing.T) {
	logger := &mockLogger{}
	cfg := ProcessConfig{
		Name:          "echo-test",
		Command:       "echo",
		Args:          []string{"hello"},
		RestartPolicy: RestartNever,
	}
	p := NewProcess(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.Run(ctx)
	if err != nil {
		t.Fatalf("expected nil error for clean exit, got: %v", err)
	}
}

func TestProcess_RestartsOnFailure(t *testing.T) {
	logger := &mockLogger{}
	cfg := ProcessConfig{
		Name:          "fail-test",
		Command:       "false",
		RestartPolicy: RestartOnFailure,
		MaxRestarts:   2,
		BackoffDelay:  10 * time.Millisecond,
	}
	p := NewProcess(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p.Run(ctx) //nolint:errcheck

	if p.restarts != 2 {
		t.Errorf("expected 2 restarts, got %d", p.restarts)
	}
}

func TestProcess_NoRestartOnSuccess(t *testing.T) {
	logger := &mockLogger{}
	cfg := ProcessConfig{
		Name:          "success-no-restart",
		Command:       "true",
		RestartPolicy: RestartOnFailure,
		MaxRestarts:   3,
		BackoffDelay:  10 * time.Millisecond,
	}
	p := NewProcess(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.restarts != 0 {
		t.Errorf("expected 0 restarts for successful process, got %d", p.restarts)
	}
}

func TestProcess_ContextCancellation(t *testing.T) {
	logger := &mockLogger{}
	cfg := ProcessConfig{
		Name:          "long-running",
		Command:       "sleep",
		Args:          []string{"60"},
		RestartPolicy: RestartAlways,
		BackoffDelay:  10 * time.Millisecond,
	}
	p := NewProcess(cfg, logger)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("process did not stop after context cancellation")
	}
}
