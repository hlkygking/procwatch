package supervisor

import (
	"encoding/json"
	"fmt"
	"os"
)

// ProcessConfig holds the configuration for a single supervised process.
type ProcessConfig struct {
	Name          string            `json:"name"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	RestartPolicy string            `json:"restart_policy"`
	MaxRestarts   int               `json:"max_restarts"`
	Tags          []string          `json:"tags"`
}

// Config is the top-level configuration structure loaded from JSON.
type Config struct {
	Processes []ProcessConfig `json:"processes"`
}

// LoadConfig reads and parses a JSON config file from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks the config for logical errors.
func (c *Config) Validate() error {
	seen := make(map[string]struct{}, len(c.Processes))
	for i, p := range c.Processes {
		if p.Name == "" {
			return fmt.Errorf("process[%d]: name is required", i)
		}
		if p.Command == "" {
			return fmt.Errorf("process %q: command is required", p.Name)
		}
		if _, dup := seen[p.Name]; dup {
			return fmt.Errorf("duplicate process name: %q", p.Name)
		}
		seen[p.Name] = struct{}{}
		if p.RestartPolicy != "" {
			if _, err := ParseRestartPolicy(p.RestartPolicy); err != nil {
				return fmt.Errorf("process %q: %w", p.Name, err)
			}
		}
	}
	return nil
}
