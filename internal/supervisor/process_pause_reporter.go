package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// ProcessPauseReporter renders pause/resume state for all tracked processes.
type ProcessPauseReporter struct {
	store  *ProcessPauseStore
	writer io.Writer
}

// NewProcessPauseReporter creates a reporter backed by the given store.
func NewProcessPauseReporter(store *ProcessPauseStore, w io.Writer) *ProcessPauseReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessPauseReporter{store: store, writer: w}
}

// PrintTable writes a tabular summary of all pause history entries.
func (r *ProcessPauseReporter) PrintTable() {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROCESS\tSTATE\tREASON\tTIMESTAMP")
	for _, e := range r.store.History() {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			e.Process,
			e.State.String(),
			e.Reason,
			e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		)
	}
	w.Flush()
}

// PrintJSON writes all pause history as a JSON array.
func (r *ProcessPauseReporter) PrintJSON() error {
	entries := r.store.History()
	type jsonEntry struct {
		Process   string `json:"process"`
		State     string `json:"state"`
		Reason    string `json:"reason"`
		Timestamp string `json:"timestamp"`
	}
	out := make([]jsonEntry, len(entries))
	for i, e := range entries {
		out[i] = jsonEntry{
			Process:   e.Process,
			State:     e.State.String(),
			Reason:    e.Reason,
			Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
