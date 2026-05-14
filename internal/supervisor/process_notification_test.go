package supervisor

import (
	"testing"
	"time"
)

func TestProcessNotificationBus_PublishAndAll(t *testing.T) {
	bus := NewProcessNotificationBus()
	bus.Publish("web", NotifyStarted, "process started")
	bus.Publish("worker", NotifyStopped, "process stopped")

	all := bus.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(all))
	}
	if all[0].Process != "web" || all[0].Kind != NotifyStarted {
		t.Errorf("unexpected first notification: %+v", all[0])
	}
	if all[1].Process != "worker" || all[1].Kind != NotifyStopped {
		t.Errorf("unexpected second notification: %+v", all[1])
	}
}

func TestProcessNotificationBus_TimestampSet(t *testing.T) {
	bus := NewProcessNotificationBus()
	before := time.Now()
	bus.Publish("svc", NotifyRestart, "restarting")
	after := time.Now()

	all := bus.All()
	if len(all) == 0 {
		t.Fatal("expected at least one notification")
	}
	ts := all[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}

func TestProcessNotificationBus_ForProcess(t *testing.T) {
	bus := NewProcessNotificationBus()
	bus.Publish("alpha", NotifyStarted, "")
	bus.Publish("beta", NotifyStarted, "")
	bus.Publish("alpha", NotifyDegraded, "high restarts")

	results := bus.ForProcess("alpha")
	if len(results) != 2 {
		t.Fatalf("expected 2 notifications for alpha, got %d", len(results))
	}
	for _, n := range results {
		if n.Process != "alpha" {
			t.Errorf("unexpected process in filtered results: %s", n.Process)
		}
	}
}

func TestProcessNotificationBus_ForProcess_NoMatch(t *testing.T) {
	bus := NewProcessNotificationBus()
	bus.Publish("alpha", NotifyStarted, "")

	results := bus.ForProcess("ghost")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestProcessNotificationBus_Subscribe_ReceivesEvent(t *testing.T) {
	bus := NewProcessNotificationBus()
	ch := bus.Subscribe()

	bus.Publish("svc", NotifyRecovered, "back online")

	select {
	case n := <-ch:
		if n.Kind != NotifyRecovered {
			t.Errorf("expected NotifyRecovered, got %s", n.Kind)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for notification")
	}
}

func TestProcessNotificationBus_AllReturnsCopy(t *testing.T) {
	bus := NewProcessNotificationBus()
	bus.Publish("svc", NotifyStarted, "")

	a := bus.All()
	a[0].Message = "mutated"

	b := bus.All()
	if b[0].Message == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}
