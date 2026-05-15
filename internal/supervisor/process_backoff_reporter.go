package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// ProcessBackoffReporter prints backoff state for all tracked processes.
type ProcessBackoffReporter struct {
	store  *ProcessBackoffStore
	writer io.Writer
}

// NewProcessBackoffReporter creates a reporter backed by the given store.
func NewProcessBackoffReporter(store *ProcessBackoffStore, w io.Writer) *ProcessBackoffReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessBackoffReporter{store: store, writer: w}
}

type backoffEntry struct {
	Process  string `json:"process"`
	Attempts int    `json:"attempts"`
	Strategy string `json:"strategy"`
	BaseDelay string `json:"base_delay"`
	MaxDelay  string `json:"max_delay"`
}

func (r *ProcessBackoffReporter) entries(processes []string) []backoffEntry {
	out := make([]backoffEntry, 0, len(processes))
	for _, p := range processes {
		pol := r.store.GetPolicy(p)
		out = append(out, backoffEntry{
			Process:   p,
			Attempts:  r.store.Attempts(p),
			Strategy:  string(pol.Strategy),
			BaseDelay: pol.BaseDelay.String(),
			MaxDelay:  pol.MaxDelay.String(),
		})
	}
	return out
}

// PrintTable writes a human-readable table of backoff state.
func (r *ProcessBackoffReporter) PrintTable(processes []string) {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROCESS\tATTEMPTS\tSTRATEGY\tBASE\tMAX")
	for _, e := range r.entries(processes) {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n", e.Process, e.Attempts, e.Strategy, e.BaseDelay, e.MaxDelay)
	}
	w.Flush()
}

// PrintJSON writes backoff state as a JSON array.
func (r *ProcessBackoffReporter) PrintJSON(processes []string) error {
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(r.entries(processes))
}
