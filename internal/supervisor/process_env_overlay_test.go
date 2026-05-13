package supervisor

import (
	"os"
	"strings"
	"testing"
)

func TestEnvOverlay_BaseOnly(t *testing.T) {
	ov := NewEnvOverlay(false)
	ov.SetBase(map[string]string{"FOO": "bar", "BAZ": "qux"})

	env := ov.Build()
	if !containsEnv(env, "FOO=bar") {
		t.Errorf("expected FOO=bar in env, got %v", env)
	}
	if !containsEnv(env, "BAZ=qux") {
		t.Errorf("expected BAZ=qux in env, got %v", env)
	}
}

func TestEnvOverlay_OverrideWinsOverBase(t *testing.T) {
	ov := NewEnvOverlay(false)
	ov.SetBase(map[string]string{"PORT": "8080"})
	ov.SetOverride("PORT", "9090")

	env := ov.Build()
	if !containsEnv(env, "PORT=9090") {
		t.Errorf("expected PORT=9090 (override), got %v", env)
	}
	if containsEnv(env, "PORT=8080") {
		t.Errorf("base PORT=8080 should be overridden")
	}
}

func TestEnvOverlay_InheritOS(t *testing.T) {
	os.Setenv("_PROCWATCH_TEST_VAR", "inherited")
	defer os.Unsetenv("_PROCWATCH_TEST_VAR")

	ov := NewEnvOverlay(true)
	env := ov.Build()

	if !containsEnv(env, "_PROCWATCH_TEST_VAR=inherited") {
		t.Errorf("expected inherited OS env var, got %v", env)
	}
}

func TestEnvOverlay_BaseOverridesOS(t *testing.T) {
	os.Setenv("_PROCWATCH_LAYER", "os")
	defer os.Unsetenv("_PROCWATCH_LAYER")

	ov := NewEnvOverlay(true)
	ov.SetBase(map[string]string{"_PROCWATCH_LAYER": "base"})

	env := ov.Build()
	if !containsEnv(env, "_PROCWATCH_LAYER=base") {
		t.Errorf("expected base to override OS, got %v", env)
	}
}

func TestEnvOverlay_ParseAndSetOverride_Valid(t *testing.T) {
	ov := NewEnvOverlay(false)
	if err := ov.ParseAndSetOverride("KEY=value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	env := ov.Build()
	if !containsEnv(env, "KEY=value") {
		t.Errorf("expected KEY=value, got %v", env)
	}
}

func TestEnvOverlay_ParseAndSetOverride_Invalid(t *testing.T) {
	ov := NewEnvOverlay(false)
	if err := ov.ParseAndSetOverride("NOEQUALSIGN"); err == nil {
		t.Error("expected error for invalid pair, got nil")
	}
}

func TestEnvOverlay_ParseAndSetOverride_ValueWithEquals(t *testing.T) {
	ov := NewEnvOverlay(false)
	if err := ov.ParseAndSetOverride("URL=http://host?a=1&b=2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	env := ov.Build()
	if !containsEnv(env, "URL=http://host?a=1&b=2") {
		t.Errorf("expected full URL value, got %v", env)
	}
}

func TestEnvOverlay_Len(t *testing.T) {
	ov := NewEnvOverlay(false)
	ov.SetBase(map[string]string{"A": "1", "B": "2"})
	ov.SetOverride("C", "3")

	if ov.Len() != 3 {
		t.Errorf("expected Len()=3, got %d", ov.Len())
	}
}

// containsEnv checks whether a "KEY=VALUE" pair exists in an env slice.
func containsEnv(env []string, pair string) bool {
	for _, e := range env {
		if strings.EqualFold(e, pair) || e == pair {
			return true
		}
	}
	return false
}
