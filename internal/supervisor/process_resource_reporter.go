package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// ResourceSnapshot is a JSON-serialisable view of a single resource sample.
type ResourceSnapshot struct {
	Name      string    `json:"name"`
	PID       int       `json:"pid"`
	RSSBytes  int64     `json:"rss_bytes"`
	RSSMB     float64   `json:"rss_mb"`
	Timestamp time.Time `json:"sampled_at"`
}

// ProcessResourceReporter formats and prints resource samples from a monitor.
type ProcessResourceReporter struct {
	monitor *ProcessResourceMonitor
	names   []string
	out     io.Writer
}

// NewProcessResourceReporter creates a reporter for the given monitor and
// process names. If out is nil, os.Stdout is used.
func NewProcessResourceReporter(monitor *ProcessResourceMonitor, names []string, out io.Writer) *ProcessResourceReporter {
	if out == nil {
		out = os.Stdout
	}
	return &ProcessResourceReporter{
		monitor: monitor,
		names:   names,
		out:     out,
	}
}

// Snapshots returns the current resource snapshots for all tracked names.
func (r *ProcessResourceReporter) Snapshots() []ResourceSnapshot {
	var out []ResourceSnapshot
	for _, name := range r.names {
		s := r.monitor.Sample(name)
		if s == nil {
			continue
		}
		out = append(out, ResourceSnapshot{
			Name:      s.Name,
			PID:       s.PID,
			RSSBytes:  s.RSSBytes,
			RSSMB:     float64(s.RSSBytes) / (1024 * 1024),
			Timestamp: s.Timestamp,
		})
	}
	return out
}

// PrintJSON writes resource snapshots as a JSON array.
func (r *ProcessResourceReporter) PrintJSON() error {
	snaps := r.Snapshots()
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(snaps)
}

// PrintTable writes resource snapshots as a human-readable table.
func (r *ProcessResourceReporter) PrintTable() {
	snaps := r.Snapshots()
	w := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPID\tRSS (MB)\tSAMPLED AT")
	for _, s := range snaps {
		fmt.Fprintf(w, "%s\t%d\t%.2f\t%s\n",
			s.Name, s.PID, s.RSSMB, s.Timestamp.Format(time.RFC3339))
	}
	w.Flush()
}
