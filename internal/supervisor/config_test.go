package supervisor

import (
	"os"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "procwatch-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadConfig_Valid(t *testing.T) {
	content := `{"processes":[{"name":"app","command":"/bin/true","restart_policy":"on-failure","max_restarts":3}]}`
	path := writeTempConfig(t, content)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(cfg.Processes))
	}
	if cfg.Processes[0].Name != "app" {
		t.Errorf("expected name 'app', got %q", cfg.Processes[0].Name)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	path := writeTempConfig(t, `{not valid json}`)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestConfig_Validate_DuplicateName(t *testing.T) {
	content := `{"processes":[{"name":"app","command":"/bin/true"},{"name":"app","command":"/bin/false"}]}`
	path := writeTempConfig(t, content)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for duplicate process name")
	}
}

func TestConfig_Validate_UnknownPolicy(t *testing.T) {
	content := `{"processes":[{"name":"app","command":"/bin/true","restart_policy":"maybe"}]}`
	path := writeTempConfig(t, content)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for unknown restart policy")
	}
}

func TestConfig_Validate_NoProcesses(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty process list")
	}
}
