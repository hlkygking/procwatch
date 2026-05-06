package supervisor

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func makeGroupRunner(t *testing.T, configs []ProcessConfig) (*GroupRunner, *HealthTracker) {
	t.Helper()
	logger := NewLogger(&bytes.Buffer{})
	health := NewHealthTracker()
	gr, err := NewGroupRunner(configs, logger, health)
	if err != nil {
		t.Fatalf("NewGroupRunner: %v", err)
	}
	return gr, health
}

func TestGroupRunner_Len(t *testing.T) {
	configs := []ProcessConfig{
		{Name: "p1", Command: "echo", Args: []string{"hello"}, Restart: "never"},
		{Name: "p2", Command: "echo", Args: []string{"world"}, Restart: "never"},
	}
	gr, _ := makeGroupRunner(t, configs)
	if gr.Len() != 2 {
		t.Fatalf("expected 2 processes, got %d", gr.Len())
	}
}

func TestGroupRunner_RunAllExit(t *testing.T) {
	configs := []ProcessConfig{
		{Name: "a", Command: "echo", Args: []string{"a"}, Restart: "never"},
		{Name: "b", Command: "echo", Args: []string{"b"}, Restart: "never"},
	}
	gr, _ := makeGroupRunner(t, configs)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		gr.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("GroupRunner.Run did not return in time")
	}
}

func TestGroupRunner_ContextCancellationStopsAll(t *testing.T) {
	configs := []ProcessConfig{
		{Name: "s1", Command: "sleep", Args: []string{"60"}, Restart: "never"},
		{Name: "s2", Command: "sleep", Args: []string{"60"}, Restart: "never"},
	}
	gr, _ := makeGroupRunner(t, configs)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		gr.Run(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("GroupRunner did not stop after context cancellation")
	}
}

func TestGroupRunner_InvalidCommand(t *testing.T) {
	configs := []ProcessConfig{
		{Name: "bad", Command: "", Restart: "never"},
	}
	logger := NewLogger(&bytes.Buffer{})
	health := NewHealthTracker()
	_, err := NewGroupRunner(configs, logger, health)
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
}
