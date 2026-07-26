//go:build e2e

package helpers

import (
	"strings"
	"testing"
)

// AssertExitCode fails if the result's exit code doesn't match expected.
func AssertExitCode(t *testing.T, res *CLIResult, expected int) {
	t.Helper()
	if res.ExitCode != expected {
		t.Errorf("expected exit code %d, got %d\nstdout: %s\nstderr: %s",
			expected, res.ExitCode, res.Stdout, res.Stderr)
	}
}

// AssertStdoutContains fails if stdout doesn't contain substr.
func AssertStdoutContains(t *testing.T, res *CLIResult, substr string) {
	t.Helper()
	if !strings.Contains(res.Stdout, substr) {
		t.Errorf("expected stdout to contain %q, got:\n%s", substr, res.Stdout)
	}
}

// AssertStderrContains fails if stderr doesn't contain substr.
func AssertStderrContains(t *testing.T, res *CLIResult, substr string) {
	t.Helper()
	if !strings.Contains(res.Stderr, substr) {
		t.Errorf("expected stderr to contain %q, got:\n%s", substr, res.Stderr)
	}
}

// AssertOutputContains checks combined stdout+stderr for substr.
func AssertOutputContains(t *testing.T, res *CLIResult, substr string) {
	t.Helper()
	if !strings.Contains(res.Combined(), substr) {
		t.Errorf("expected output to contain %q, got:\nstdout: %s\nstderr: %s",
			substr, res.Stdout, res.Stderr)
	}
}

// AssertStdoutNotContains fails if stdout contains substr.
func AssertStdoutNotContains(t *testing.T, res *CLIResult, substr string) {
	t.Helper()
	if strings.Contains(res.Stdout, substr) {
		t.Errorf("expected stdout NOT to contain %q, got:\n%s", substr, res.Stdout)
	}
}

// AssertSuccess fails if exit code is not 0.
func AssertSuccess(t *testing.T, res *CLIResult) {
	t.Helper()
	AssertExitCode(t, res, 0)
}

// AssertFailure fails if exit code is 0.
func AssertFailure(t *testing.T, res *CLIResult) {
	t.Helper()
	if res.ExitCode == 0 {
		t.Errorf("expected non-zero exit code, got 0\nstdout: %s", res.Stdout)
	}
}
