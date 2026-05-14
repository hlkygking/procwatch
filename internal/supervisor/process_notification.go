package supervisor

import (
	"sync"
	"time"
)

// NotificationKind represents the type of process notification.
type NotificationKind string

const (
	NotifyStarted  NotificationKind = "started"
	NotifyStopped  NotificationKind = "stopped"
	NotifyRestart  NotificationKind = "restart"
	NotifyDegraded NotificationKind = "degraded"
	NotifyRecovered NotificationKind = "recovered"
)

// ProcessNotification represents a single notification event for a process.
type ProcessNotification struct {
	Process   string           `json:"process"`
	Kind      NotificationKind `json:"kind"`
	Message   string           `json:"message"`
	Timestamp time.Time        `json:"timestamp"`
}

// ProcessNotificationBus collects and dispatches process notifications.
type ProcessNotificationBus struct {
	mu            sync.Mutex
	notifications []ProcessNotification
	subscribers   []chan ProcessNotification
}

// NewProcessNotificationBus creates a new ProcessNotificationBus.
func NewProcessNotificationBus() *ProcessNotificationBus {
	return &ProcessNotificationBus{}
}

// Publish records and dispatches a notification to all subscribers.
func (b *ProcessNotificationBus) Publish(process string, kind NotificationKind, message string) {
	n := ProcessNotification{
		Process:   process,
		Kind:      kind,
		Message:   message,
		Timestamp: time.Now(),
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.notifications = append(b.notifications, n)
	for _, ch := range b.subscribers {
		select {
		case ch <- n:
		default:
		}
	}
}

// All returns a copy of all recorded notifications.
func (b *ProcessNotificationBus) All() []ProcessNotification {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]ProcessNotification, len(b.notifications))
	copy(out, b.notifications)
	return out
}

// ForProcess returns notifications filtered by process name.
func (b *ProcessNotificationBus) ForProcess(name string) []ProcessNotification {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []ProcessNotification
	for _, n := range b.notifications {
		if n.Process == name {
			out = append(out, n)
		}
	}
	return out
}

// Subscribe returns a channel that receives future notifications.
func (b *ProcessNotificationBus) Subscribe() <-chan ProcessNotification {
	ch := make(chan ProcessNotification, 16)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = append(b.subscribers, ch)
	return ch
}
