package supervisor

import (
	"bytes"
	"context"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSignalHandler_CancelsContextOnSIGINT(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	ctx, cancel := context.WithCancel(context.Background())

	sh := NewSignalHandler(logger, cancel)
	sh.Start()
	defer sh.Stop()

	sh.signals <- syscall.SIGINT

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled after SIGINT")
	}
}

func TestSignalHandler_CancelsContextOnSIGTERM(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	ctx, cancel := context.WithCancel(context.Background())

	sh := NewSignalHandler(logger, cancel)
	sh.Start()
	defer sh.Stop()

	sh.signals <- syscall.SIGTERM

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled after SIGTERM")
	}
}

func TestSignalHandler_LogsSignalName(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	_, cancel := context.WithCancel(context.Background())

	sh := NewSignalHandler(logger, cancel)
	sh.Start()

	sh.signals <- syscall.SIGINT
	time.Sleep(50 * time.Millisecond)
	sh.Stop()

	output := buf.String()
	if !strings.Contains(output, "received signal") {
		t.Errorf("expected log to contain 'received signal', got: %s", output)
	}
	if !strings.Contains(output, "interrupt") {
		t.Errorf("expected log to contain signal name 'interrupt', got: %s", output)
	}
}

func TestSignalHandler_StopWithoutSignal(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	_, cancel := context.WithCancel(context.Background())

	sh := NewSignalHandler(logger, cancel)
	sh.Start()

	done := make(chan struct{})
	go func() {
		sh.Stop()
		close(done)
	}()

	select {
	case <-done:
		// expected — Stop should return cleanly
	case <-time.After(time.Second):
		t.Fatal("Stop() blocked unexpectedly")
	}
}
