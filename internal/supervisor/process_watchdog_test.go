package supervisor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestProcessWatchdog_CallsOnDeadAfterMaxMissed(t *testing.T) {
	var deadCalled atomic.Int32
	var missedCount atomic.Int32

	wdg := NewProcessWatchdog(WatchdogConfig{
		ProcessName: "svc",
		Interval:    10 * time.Millisecond,
		MaxMissed:   2,
		PingFn:      func(_ context.Context) bool { return false },
		OnDead: func(name string, missed int) {
			deadCalled.Add(1)
			missedCount.Store(int32(missed))
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go wdg.Start(ctx)
	<-ctx.Done()

	if deadCalled.Load() == 0 {
		t.Fatal("expected OnDead to be called")
	}
	if missedCount.Load() < 2 {
		t.Errorf("expected at least 2 missed pings, got %d", missedCount.Load())
	}
}

func TestProcessWatchdog_NoCallWhenAlive(t *testing.T) {
	var deadCalled atomic.Int32

	wdg := NewProcessWatchdog(WatchdogConfig{
		ProcessName: "svc",
		Interval:    10 * time.Millisecond,
		MaxMissed:   2,
		PingFn:      func(_ context.Context) bool { return true },
		OnDead:      func(_ string, _ int) { deadCalled.Add(1) },
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	go wdg.Start(ctx)
	<-ctx.Done()

	if deadCalled.Load() != 0 {
		t.Errorf("expected OnDead not to be called, was called %d times", deadCalled.Load())
	}
}

func TestProcessWatchdog_ResetClearsMissedCount(t *testing.T) {
	wdg := NewProcessWatchdog(WatchdogConfig{
		ProcessName: "svc",
		Interval:    time.Hour,
		MaxMissed:   5,
		PingFn:      func(_ context.Context) bool { return false },
	})

	wdg.check(context.Background())
	wdg.check(context.Background())
	if wdg.MissedCount() != 2 {
		t.Fatalf("expected 2 missed, got %d", wdg.MissedCount())
	}
	wdg.Reset()
	if wdg.MissedCount() != 0 {
		t.Errorf("expected 0 after reset, got %d", wdg.MissedCount())
	}
}

func TestProcessWatchdog_DefaultIntervalAndMaxMissed(t *testing.T) {
	wdg := NewProcessWatchdog(WatchdogConfig{
		ProcessName: "svc",
		PingFn:      func(_ context.Context) bool { return true },
	})
	if wdg.cfg.Interval != 5*time.Second {
		t.Errorf("expected default interval 5s, got %v", wdg.cfg.Interval)
	}
	if wdg.cfg.MaxMissed != 3 {
		t.Errorf("expected default MaxMissed 3, got %d", wdg.cfg.MaxMissed)
	}
}

func TestProcessWatchdog_MissedCountAccumulates(t *testing.T) {
	wdg := NewProcessWatchdog(WatchdogConfig{
		ProcessName: "svc",
		Interval:    time.Hour,
		MaxMissed:   100,
		PingFn:      func(_ context.Context) bool { return false },
	})
	for i := 0; i < 5; i++ {
		wdg.check(context.Background())
	}
	if wdg.MissedCount() != 5 {
		t.Errorf("expected 5 missed, got %d", wdg.MissedCount())
	}
}
