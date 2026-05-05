package supervisor

import (
	"errors"
	"testing"
)

func TestParseRestartPolicy(t *testing.T) {
	tests := []struct {
		input    string
		want     RestartPolicy
		wantOK   bool
	}{
		{"never", RestartNever, true},
		{"no", RestartNever, true},
		{"on-failure", RestartOnFailure, true},
		{"onfailure", RestartOnFailure, true},
		{"failure", RestartOnFailure, true},
		{"always", RestartAlways, true},
		{"ALWAYS", RestartAlways, true},
		{"invalid", RestartNever, false},
		{"", RestartNever, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseRestartPolicy(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseRestartPolicy(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("ParseRestartPolicy(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRestartPolicy_ShouldRestart(t *testing.T) {
	errFake := errors.New("exit status 1")

	tests := []struct {
		policy   RestartPolicy
		exitErr  error
		want     bool
	}{
		{RestartNever, nil, false},
		{RestartNever, errFake, false},
		{RestartOnFailure, nil, false},
		{RestartOnFailure, errFake, true},
		{RestartAlways, nil, true},
		{RestartAlways, errFake, true},
	}

	for _, tt := range tests {
		t.Run(tt.policy.String(), func(t *testing.T) {
			got := tt.policy.ShouldRestart(tt.exitErr)
			if got != tt.want {
				t.Errorf("%v.ShouldRestart(%v) = %v, want %v", tt.policy, tt.exitErr, got, tt.want)
			}
		})
	}
}

func TestRestartPolicy_String(t *testing.T) {
	if s := RestartNever.String(); s != "never" {
		t.Errorf("expected 'never', got %q", s)
	}
	if s := RestartOnFailure.String(); s != "on-failure" {
		t.Errorf("expected 'on-failure', got %q", s)
	}
	if s := RestartAlways.String(); s != "always" {
		t.Errorf("expected 'always', got %q", s)
	}
}
