//go:build e2e

// Package helpers provides test utilities for azfloci end-to-end tests.
// It wraps the compiled CLI binary via os/exec, captures stdout/stderr/exit
// codes, and offers Azure-specific helpers for resource setup and teardown.
package helpers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// CLIResult captures the output of a CLI invocation.
type CLIResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Success returns true when the process exited 0.
func (r *CLIResult) Success() bool { return r.ExitCode == 0 }

// Combined returns stdout + stderr concatenated (useful for substring checks).
func (r *CLIResult) Combined() string { return r.Stdout + r.Stderr }

// CLIRunner manages execution of the azfloci binary under test.
type CLIRunner struct {
	// BinaryPath is the absolute path to the compiled azfloci binary.
	BinaryPath string
	// WorkDir is the working directory for CLI invocations.
	WorkDir string
	// Env holds extra environment variables injected into every run.
	Env []string
	// Timeout is the default per-command timeout (overridable per call).
	Timeout time.Duration
}

// NewCLIRunner builds the azfloci binary into a temp directory and returns
// a runner pointing at it. Call t.Cleanup to remove the temp dir.
func NewCLIRunner(t *testing.T) *CLIRunner {
	t.Helper()

	// Locate the repo root (walk up from this file's dir until we find go.mod).
	repoRoot := findRepoRoot(t)

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "azfloci")

	// Build the binary with race detector.
	cmd := exec.Command("go", "build", "-race", "-o", binPath, "./cmd/azfloci")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build azfloci binary: %v\n%s", err, string(out))
	}

	return &CLIRunner{
		BinaryPath: binPath,
		WorkDir:    tmpDir,
		Timeout:    30 * time.Second,
	}
}

// Run executes the azfloci binary with the given arguments.
func (r *CLIRunner) Run(args ...string) *CLIResult {
	return r.RunWithContext(context.Background(), args...)
}

// RunWithContext executes the binary with a caller-supplied context.
func (r *CLIRunner) RunWithContext(ctx context.Context, args ...string) *CLIResult {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, r.BinaryPath, args...)
	cmd.Dir = r.WorkDir
	if len(r.Env) > 0 {
		cmd.Env = append(os.Environ(), r.Env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	result := &CLIResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: elapsed,
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil && result.ExitCode == 0 {
		// Context cancelled / signal — mark as non-zero.
		result.ExitCode = -1
	}
	return result
}

// RunWithEnv executes the binary with additional env vars for this call only.
func (r *CLIRunner) RunWithEnv(env []string, args ...string) *CLIResult {
	old := r.Env
	r.Env = append(r.Env, env...)
	defer func() { r.Env = old }()
	return r.Run(args...)
}

// MustRun calls Run and fails the test if the exit code is non-zero.
func (r *CLIRunner) MustRun(t *testing.T, args ...string) *CLIResult {
	t.Helper()
	res := r.Run(args...)
	if !res.Success() {
		t.Fatalf("azfloci %s failed (exit %d):\nstdout: %s\nstderr: %s",
			strings.Join(args, " "), res.ExitCode, res.Stdout, res.Stderr)
	}
	return res
}

// findRepoRoot walks up from the current working directory to find go.mod.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find repo root (no go.mod found in parent dirs)")
		}
		dir = parent
	}
}

// AzCLIAvailable returns true if the `az` CLI is on PATH.
func AzCLIAvailable() bool {
	_, err := exec.LookPath("az")
	return err == nil
}

// RequireAzCLI skips the test if `az` is not available.
func RequireAzCLI(t *testing.T) {
	t.Helper()
	if !AzCLIAvailable() {
		t.Skip("skipping: az CLI not found on PATH")
	}
}

// RunAzCLI executes a raw `az` command and returns the result.
// Useful for validating Azure state outside the azfloci binary.
func RunAzCLI(ctx context.Context, args ...string) (*CLIResult, error) {
	azPath, err := exec.LookPath("az")
	if err != nil {
		return nil, fmt.Errorf("az CLI not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, azPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)

	result := &CLIResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: elapsed,
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	return result, err
}
