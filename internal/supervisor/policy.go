package supervisor

import "strings"

// RestartPolicy defines how a process should be restarted.
type RestartPolicy int

const (
	// RestartNever means the process will not be restarted.
	RestartNever RestartPolicy = iota
	// RestartOnFailure means the process will be restarted only if it exits with a non-zero code.
	RestartOnFailure
	// RestartAlways means the process will always be restarted regardless of exit code.
	RestartAlways
)

// ParseRestartPolicy parses a string into a RestartPolicy.
func ParseRestartPolicy(s string) (RestartPolicy, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "never", "no":
		return RestartNever, true
	case "on-failure", "onfailure", "failure":
		return RestartOnFailure, true
	case "always":
		return RestartAlways, true
	default:
		return RestartNever, false
	}
}

// String returns the string representation of a RestartPolicy.
func (p RestartPolicy) String() string {
	switch p {
	case RestartNever:
		return "never"
	case RestartOnFailure:
		return "on-failure"
	case RestartAlways:
		return "always"
	default:
		return "unknown"
	}
}

// ShouldRestart determines whether a process should be restarted based on the policy and exit error.
func (p RestartPolicy) ShouldRestart(exitErr error) bool {
	switch p {
	case RestartAlways:
		return true
	case RestartOnFailure:
		return exitErr != nil
	default:
		return false
	}
}
