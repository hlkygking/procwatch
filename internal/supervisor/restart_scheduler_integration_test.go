package supervisor

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRestartScheduler_LogsDelayAndProcess(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf)
	limiter := NewRestartLimiter(5, time.Minute)
	policy := ThrottlePolicy{
		MinDelay: 10 * time.Millisecond,
		MaxDelay: 200 * time.Millisecond,
		Factor:   2.0,
	}
	rs := NewRestartScheduler("my-service", limiter, policy, logger)

	ctx := context.Background()
	rs.Schedule(ctx, 2)

	output := buf.String()
	if !strings.Contains(output, "my-service") {
		t.Errorf("expected log to contain process name, got: %s", output)
	}
	if !strings.Contains(output, "scheduling restart") {
		t.Errorf("expected log to contain scheduling restart, got: %s", output)
	}
}

func TestRestartScheduler_BackoffGrowsAcrossRestarts(t *testing.T) {
	rs := makeScheduler(10)
	ctx := context.Background()

	delays := make([]time.Duration, 3)
	for i := range delays {
		before := rs.throttle.Current()
		rs.Schedule(ctx, 1)
		delays[i] = before
	}

	if delays[1] <= delays[0] {
		t.Errorf("expected delay[1] > delay[0], got %v <= %v", delays[1], delays[0])
	}
	if delays[2] <= delays[1] {
		t.Errorf("expected delay[2] > delay[1], got %v <= %v", delays[2], delays[1])
	}
}

func TestRestartScheduler_MultipleResetsAreTracked(t *testing.T) {
	rs := makeScheduler(10)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		rs.Schedule(ctx, 0)
		rs.RecordSuccess()
	}

	if rs.throttle.Resets() != 3 {
		t.Errorf("expected 3 resets, got %d", rs.throttle.Resets())
	}
}
