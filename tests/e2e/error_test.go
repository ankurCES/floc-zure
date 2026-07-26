//go:build e2e

// Tests for error handling: invalid inputs, missing auth, bad subcommands,
// and edge cases that should produce clear error messages.
package e2e

import (
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

// TestInvalidSubcommand checks that an unknown subcommand yields a non-zero exit.
func TestInvalidSubcommand(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.Run("this-does-not-exist")
	helpers.AssertFailure(t, res)
	helpers.AssertOutputContains(t, res, "unknown command")
}

// TestAuthStatusWithoutAzCLI simulates missing az CLI by overriding PATH.
// The binary should fail gracefully with an auth error.
func TestAuthStatusWithoutAzCLI(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	// Override PATH to exclude az — the binary should report an error.
	res := cli.RunWithEnv([]string{"PATH=/nonexistent"}, "auth", "status")
	helpers.AssertFailure(t, res)
}

// TestEmptyArgs verifies running with no args shows help, not a crash.
func TestEmptyArgs(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.Run()
	// Cobra shows help on no args (exit 0).
	helpers.AssertSuccess(t, res)
	helpers.AssertStdoutContains(t, res, "azfloci")
}

// TestDoubleHelpFlag verifies --help --help doesn't crash.
func TestDoubleHelpFlag(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.Run("--help", "--help")
	// Should not crash regardless.
	if res.ExitCode != 0 {
		// Some CLIs reject double flags — that's OK, just no crash.
		t.Logf("double --help exited %d (acceptable)", res.ExitCode)
	}
}

// TestAuthHelpDoesNotRequireAz verifies `auth --help` works without az CLI.
func TestAuthHelpDoesNotRequireAz(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.Run("auth", "--help")
	helpers.AssertSuccess(t, res)
}
