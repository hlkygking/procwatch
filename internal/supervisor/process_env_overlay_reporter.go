package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ProcessEnvOverlayReporter prints the resolved environment for a named process.
type ProcessEnvOverlayReporter struct {
	overlay *EnvOverlay
	name    string
	out     io.Writer
}

// NewProcessEnvOverlayReporter creates a reporter for the given overlay and process name.
// If out is nil, os.Stdout is used.
func NewProcessEnvOverlayReporter(name string, overlay *EnvOverlay, out io.Writer) *ProcessEnvOverlayReporter {
	if out == nil {
		out = os.Stdout
	}
	return &ProcessEnvOverlayReporter{
		overlay: overlay,
		name:    name,
		out:     out,
	}
}

// PrintTable writes the resolved environment as a human-readable table.
func (r *ProcessEnvOverlayReporter) PrintTable() {
	env := r.overlay.Build()
	sort.Strings(env)

	fmt.Fprintf(r.out, "Environment for process %q (%d vars):\n", r.name, len(env))
	for _, pair := range env {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			fmt.Fprintf(r.out, "  %-30s = %s\n", parts[0], parts[1])
		}
	}
}

// PrintJSON writes the resolved environment as a JSON object.
func (r *ProcessEnvOverlayReporter) PrintJSON() error {
	env := r.overlay.Build()
	m := make(map[string]string, len(env))
	for _, pair := range env {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}

	payload := map[string]interface{}{
		"process": r.name,
		"env":     m,
	}

	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}
