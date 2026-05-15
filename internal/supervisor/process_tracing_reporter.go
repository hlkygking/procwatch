package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// ProcessTracingReporter prints trace events from a ProcessTracingStore.
type ProcessTracingReporter struct {
	store  *ProcessTracingStore
	writer io.Writer
}

// NewProcessTracingReporter creates a reporter backed by the given store.
func NewProcessTracingReporter(store *ProcessTracingStore, w io.Writer) *ProcessTracingReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessTracingReporter{store: store, writer: w}
}

// PrintTable writes a human-readable table of trace events.
func (r *ProcessTracingReporter) PrintTable() {
	events := r.store.All()
	tw := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROCESS\tKIND\tSTARTED\tDURATION")
	for _, e := range events {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			e.Process,
			e.Kind,
			e.StartedAt.Format("15:04:05"),
			e.Duration.Round(1000000),
		)
	}
	tw.Flush()
}

// PrintJSON writes all trace events as a JSON array.
func (r *ProcessTracingReporter) PrintJSON() error {
	events := r.store.All()
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(events)
}
