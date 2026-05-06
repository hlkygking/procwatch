package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// HealthReporter formats and writes health summaries from a HealthTracker.
type HealthReporter struct {
	tracker *HealthTracker
	out     io.Writer
}

// NewHealthReporter creates a HealthReporter writing to out.
// If out is nil, os.Stdout is used.
func NewHealthReporter(tracker *HealthTracker, out io.Writer) *HealthReporter {
	if out == nil {
		out = os.Stdout
	}
	return &HealthReporter{tracker: tracker, out: out}
}

// PrintTable writes a human-readable table of all process health records.
func (r *HealthReporter) PrintTable() {
	w := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tRESTARTS\tLAST START\tEXIT CODE")
	for _, rec := range r.tracker.All() {
		exitCode := "-"
		if rec.ExitCode != nil {
			exitCode = fmt.Sprintf("%d", *rec.ExitCode)
		}
		lastStart := "-"
		if !rec.LastStart.IsZero() {
			lastStart = rec.LastStart.Format("15:04:05")
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			rec.Name,
			rec.Status.String(),
			rec.Restarts,
			lastStart,
			exitCode,
		)
	}
	w.Flush()
}

// PrintJSON writes all health records as a JSON array.
func (r *HealthReporter) PrintJSON() error {
	records := r.tracker.All()
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

// Summary returns a brief string summarising overall health.
func (r *HealthReporter) Summary() string {
	records := r.tracker.All()
	healthy, unhealthy, starting := 0, 0, 0
	for _, rec := range records {
		switch rec.Status {
		case StatusHealthy:
			healthy++
		case StatusUnhealthy:
			unhealthy++
		case StatusStarting:
			starting++
		}
	}
	return fmt.Sprintf("total=%d healthy=%d unhealthy=%d starting=%d",
		len(records), healthy, unhealthy, starting)
}
