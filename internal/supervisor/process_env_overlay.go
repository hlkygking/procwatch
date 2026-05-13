package supervisor

import (
	"fmt"
	"os"
	"strings"
)

// EnvOverlay merges environment variable layers for a process.
// Priority (highest to lowest): overrides > process config > inherited OS env.
type EnvOverlay struct {
	inheritOS bool
	base      map[string]string
	overrides map[string]string
}

// NewEnvOverlay creates an EnvOverlay. If inheritOS is true, the current
// process environment is included as the lowest-priority layer.
func NewEnvOverlay(inheritOS bool) *EnvOverlay {
	return &EnvOverlay{
		inheritOS: inheritOS,
		base:      make(map[string]string),
		overrides: make(map[string]string),
	}
}

// SetBase sets the base environment from a process config (e.g. procwatch.json env block).
func (e *EnvOverlay) SetBase(env map[string]string) {
	for k, v := range env {
		e.base[k] = v
	}
}

// SetOverride adds a highest-priority key/value pair (e.g. from CLI flags).
func (e *EnvOverlay) SetOverride(key, value string) {
	e.overrides[key] = value
}

// ParseAndSetOverride parses a "KEY=VALUE" string and sets it as an override.
func (e *EnvOverlay) ParseAndSetOverride(pair string) error {
	parts := strings.SplitN(pair, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid env pair %q: expected KEY=VALUE", pair)
	}
	e.overrides[parts[0]] = parts[1]
	return nil
}

// Build returns the final merged environment as a []string slice
// suitable for use with exec.Cmd.Env.
func (e *EnvOverlay) Build() []string {
	merged := make(map[string]string)

	if e.inheritOS {
		for _, entry := range os.Environ() {
			parts := strings.SplitN(entry, "=", 2)
			if len(parts) == 2 {
				merged[parts[0]] = parts[1]
			}
		}
	}

	for k, v := range e.base {
		merged[k] = v
	}

	for k, v := range e.overrides {
		merged[k] = v
	}

	result := make([]string, 0, len(merged))
	for k, v := range merged {
		result = append(result, k+"="+v)
	}
	return result
}

// Len returns the number of unique keys in the final merged environment.
func (e *EnvOverlay) Len() int {
	return len(e.Build())
}
