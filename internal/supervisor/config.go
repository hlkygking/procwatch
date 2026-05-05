package supervisor

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ProcessConfig holds the configuration for a single supervised process.
type ProcessConfig struct {
	Name          string        `json:"name"`
	Command       string        `json:"command"`
	Args          []string      `json:"args"`
	RestartPolicy string        `json:"restart_policy"`
	MaxRestarts   int           `json:"max_restarts"`
	RestartDelay  time.Duration `json:"restart_delay"`
	Env           []string      `json:"env"`
}

// Config is the top-level configuration for procwatch.
type Config struct {
	Processes []ProcessConfig `json:"processes"`
}

// LoadConfig reads and parses a JSON config file from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the config is semantically valid.
func (c *Config) Validate() error {
	if len(c.Processes) == 0 {
		return fmt.Errorf("no processes defined")
	}
	seen := make(map[string]struct{})
	for i, p := range c.Processes {
		if p.Name == "" {
			return fmt.Errorf("process[%d]: name is required", i)
		}
		if p.Command == "" {
			return fmt.Errorf("process %q: command is required", p.Name)
		}
		if _, exists := seen[p.Name]; exists {
			return fmt.Errorf("duplicate process name: %q", p.Name)
		}
		seen[p.Name] = struct{}{}
		if p.RestartPolicy != "" {
			if _, ok := ParseRestartPolicy(p.RestartPolicy); !ok {
				return fmt.Errorf("process %q: unknown restart_policy %q", p.Name, p.RestartPolicy)
			}
		}
		if p.MaxRestarts < 0 {
			return fmt.Errorf("process %q: max_restarts must be >= 0", p.Name)
		}
	}
	return nil
}
