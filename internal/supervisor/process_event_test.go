package supervisor

import (
	"testing"
	"time"
)

func TestProcessEventBus_PublishAndAll(t *testing.T) {
	bus := NewProcessEventBus()
	bus.Publish(ProcessEvent{ProcessName: "web", Type: EventStarted})
	bus.Publish(ProcessEvent{ProcessName: "worker", Type: EventStopped, ExitCode: 1})

	events := bus.All()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].ProcessName != "web" || events[0].Type != EventStarted {
		t.Errorf("unexpected first event: %+v", events[0])
	}
}

func TestProcessEventBus_TimestampSetAutomatically(t *testing.T) {
	bus := NewProcessEventBus()
	before := time.Now()
	bus.Publish(ProcessEvent{ProcessName: "web", Type: EventStarted})
	after := time.Now()

	events := bus.All()
	ts := events[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}

func TestProcessEventBus_FilterByProcess(t *testing.T) {
	bus := NewProcessEventBus()
	bus.Publish(ProcessEvent{ProcessName: "web", Type: EventStarted})
	bus.Publish(ProcessEvent{ProcessName: "worker", Type: EventStarted})
	bus.Publish(ProcessEvent{ProcessName: "web", Type: EventStopped})

	webEvents := bus.FilterByProcess("web")
	if len(webEvents) != 2 {
		t.Fatalf("expected 2 events for 'web', got %d", len(webEvents))
	}
	for _, e := range webEvents {
		if e.ProcessName != "web" {
			t.Errorf("unexpected process name: %s", e.ProcessName)
		}
	}
}

func TestProcessEventBus_FilterByProcess_NoMatch(t *testing.T) {
	bus := NewProcessEventBus()
	bus.Publish(ProcessEvent{ProcessName: "web", Type: EventStarted})

	result := bus.FilterByProcess("missing")
	if result != nil {
		t.Errorf("expected nil for no-match, got %v", result)
	}
}

func TestProcessEventBus_Subscribe_ReceivesEvent(t *testing.T) {
	bus := NewProcessEventBus()
	ch := bus.Subscribe(4)

	bus.Publish(ProcessEvent{ProcessName: "web", Type: EventRestart})

	select {
	case e := <-ch:
		if e.Type != EventRestart {
			t.Errorf("expected EventRestart, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for event on subscriber channel")
	}
}

func TestProcessEventBus_Subscribe_DefaultBufferSize(t *testing.T) {
	bus := NewProcessEventBus()
	ch := bus.Subscribe(0) // should default to 16
	if cap(ch) != 16 {
		t.Errorf("expected buffer size 16, got %d", cap(ch))
	}
}

func TestProcessEventBus_AllReturnsCopy(t *testing.T) {
	bus := NewProcessEventBus()
	bus.Publish(ProcessEvent{ProcessName: "web", Type: EventStarted})

	a := bus.All()
	a[0].ProcessName = "mutated"

	b := bus.All()
	if b[0].ProcessName == "mutated" {
		t.Error("All() should return a copy, not a reference to internal slice")
	}
}
