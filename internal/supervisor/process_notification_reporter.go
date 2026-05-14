package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// ProcessNotificationReporter prints notification history.
type ProcessNotificationReporter struct {
	bus    *ProcessNotificationBus
	writer io.Writer
}

// NewProcessNotificationReporter creates a reporter backed by the given bus.
func NewProcessNotificationReporter(bus *ProcessNotificationBus, w io.Writer) *ProcessNotificationReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessNotificationReporter{bus: bus, writer: w}
}

// PrintTable writes a human-readable table of all notifications.
func (r *ProcessNotificationReporter) PrintTable() {
	entries := r.bus.All()
	tw := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROCESS\tKIND\tMESSAGE\tTIMESTAMP")
	for _, n := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			n.Process,
			string(n.Kind),
			n.Message,
			n.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		)
	}
	tw.Flush()
}

// PrintJSON writes all notifications as a JSON array.
func (r *ProcessNotificationReporter) PrintJSON() error {
	entries := r.bus.All()
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
