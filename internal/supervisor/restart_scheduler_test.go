package supervisor

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func makeScheduler(maxRestarts int) *RestartScheduler {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf)
	limiter := NewRestartLimiter(maxRestarts, time.Minute)
	policy := ThrottlePolicy{
		MinDelay: 10 * time.Millisecond,
		MaxDelay: 100 * time.Millisecond,
		Factor:   2.0,
	}
	return NewRestartScheduler("test-proc", limiter, policy, logger)
}

func TestRestartScheduler_AllowsRestart(t *testing.T) {
	rs := makeScheduler(5)
	ctx := context.Background()
	if !rs.Schedule(ctx, 1) {
		t.Error("expected restart to be allowed")
	}
}

func TestRestartScheduler_DeniesWhenLimitReached(t *testing.T) {
	rs := makeScheduler(1)
	ctx := context.Background()
	rs.Schedule(ctx, 1) // consume the one allowed restart
	if rs.Schedule(ctx, 1) {
		t.Error("expected restart to be denied after limit")
	}
}

func TestRestartScheduler_ContextCancellation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf)
	limiter := NewRestartLimiter(-1, time.Minute)
	policy := ThrottlePolicy{
		MinDelay: 5 * time.Second, // long delay to ensure cancellation wins
		MaxDelay: 10 * time.Second,
		Factor:   2.0,
	}
	rs := NewRestartScheduler("slow-proc", limiter, policy, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	result := rs.Schedule(ctx, 0)
	elapsed := time.Since(start)

	if result {
		t.Error("expected false when context is cancelled")
	}
	if elapsed > time.Second {
		t.Errorf("expected fast cancellation, took %v", elapsed)
	}
}

func TestRestartScheduler_RecordSuccessResetsState(t *testing.T) {
	rs := makeScheduler(2)
	ctx := context.Background()

	rs.Schedule(ctx, 1)
	rs.RecordSuccess()

	// After reset, throttle should be back to min delay
	if rs.throttle.Current() != rs.throttle.policy.MinDelay {
		t.Errorf("expected throttle reset to min delay, got %v", rs.throttle.Current())
	}
}
