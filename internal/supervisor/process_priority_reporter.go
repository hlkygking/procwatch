package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ProcessPriorityReporter renders process priority assignments.
type ProcessPriorityReporter struct {
	store  *ProcessPriorityStore
	writer io.Writer
}

// NewProcessPriorityReporter creates a reporter backed by the given store.
func NewProcessPriorityReporter(store *ProcessPriorityStore, w io.Writer) *ProcessPriorityReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessPriorityReporter{store: store, writer: w}
}

// PrintTable writes a human-readable table of process priorities.
func (r *ProcessPriorityReporter) PrintTable() {
	entries := r.store.All()
	if len(entries) == 0 {
		fmt.Fprintln(r.writer, "no process priorities configured")
		return
	}
	fmt.Fprintf(r.writer, "%-24s %s\n", "PROCESS", "PRIORITY")
	for _, e := range entries {
		fmt.Fprintf(r.writer, "%-24s %s\n", e.Process, e.Priority.String())
	}
}

// PrintJSON writes process priorities as a JSON array.
func (r *ProcessPriorityReporter) PrintJSON() error {
	entries := r.store.All()
	type row struct {
		Process  string `json:"process"`
		Priority string `json:"priority"`
	}
	rows := make([]row, len(entries))
	for i, e := range entries {
		rows[i] = row{Process: e.Process, Priority: e.Priority.String()}
	}
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
