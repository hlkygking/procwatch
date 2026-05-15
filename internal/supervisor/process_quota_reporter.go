package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// ProcessQuotaReporter renders quota and violation data.
type ProcessQuotaReporter struct {
	store  *ProcessQuotaStore
	writer io.Writer
}

func NewProcessQuotaReporter(store *ProcessQuotaStore, w io.Writer) *ProcessQuotaReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessQuotaReporter{store: store, writer: w}
}

func (r *ProcessQuotaReporter) PrintTable() {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROCESS\tREASON\tAT")
	for _, v := range r.store.Violations() {
		fmt.Fprintf(w, "%s\t%s\t%s\n", v.Process, v.Reason, v.At.Format(time.RFC3339))
	}
	w.Flush()
}

func (r *ProcessQuotaReporter) PrintJSON() error {
	violations := r.store.Violations()
	type entry struct {
		Process string    `json:"process"`
		Reason  string    `json:"reason"`
		At      time.Time `json:"at"`
	}
	out := make([]entry, len(violations))
	for i, v := range violations {
		out[i] = entry{Process: v.Process, Reason: v.Reason, At: v.At}
	}
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
