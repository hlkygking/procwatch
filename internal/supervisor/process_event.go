package supervisor

import "time"

// EventType represents the kind of lifecycle event that occurred for a process.
type EventType string

const (
	EventStarted  EventType = "started"
	EventStopped  EventType = "stopped"
	EventRestart  EventType = "restarting"
	EventFailed   EventType = "failed"
	EventThrottle EventType = "throttled"
)

// ProcessEvent captures a single lifecycle event for a named process.
type ProcessEvent struct {
	ProcessName string    `json:"process"`
	Type        EventType `json:"event"`
	Message     string    `json:"message,omitempty"`
	ExitCode    int       `json:"exit_code,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// ProcessEventBus collects and distributes process lifecycle events.
type ProcessEventBus struct {
	events []ProcessEvent
	subs   []chan ProcessEvent
}

// NewProcessEventBus creates an empty ProcessEventBus.
func NewProcessEventBus() *ProcessEventBus {
	return &ProcessEventBus{}
}

// Publish records an event and fans it out to all subscribers.
func (b *ProcessEventBus) Publish(e ProcessEvent) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	b.events = append(b.events, e)
	for _, ch := range b.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

// Subscribe returns a channel that receives future events.
// The channel is buffered to avoid blocking the publisher.
func (b *ProcessEventBus) Subscribe(bufSize int) <-chan ProcessEvent {
	if bufSize <= 0 {
		bufSize = 16
	}
	ch := make(chan ProcessEvent, bufSize)
	b.subs = append(b.subs, ch)
	return ch
}

// All returns a snapshot of all events recorded so far.
func (b *ProcessEventBus) All() []ProcessEvent {
	out := make([]ProcessEvent, len(b.events))
	copy(out, b.events)
	return out
}

// FilterByProcess returns events whose ProcessName matches name.
func (b *ProcessEventBus) FilterByProcess(name string) []ProcessEvent {
	var out []ProcessEvent
	for _, e := range b.events {
		if e.ProcessName == name {
			out = append(out, e)
		}
	}
	return out
}
