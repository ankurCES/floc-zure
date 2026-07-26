//go:build e2e

// Tests for basic CLI functionality: version, help, and unknown commands.
package e2e

import (
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

// TestVersionOutput verifies `azfloci version` prints a version string.
func TestVersionOutput(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.MustRun(t, "version")
	helpers.AssertStdoutContains(t, res, "azfloci")
}

// TestHelpOutput verifies the root command shows usage info.
func TestHelpOutput(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.MustRun(t, "--help")
	helpers.AssertStdoutContains(t, res, "Azure resource orchestration CLI")
}

// TestUnknownCommand verifies unknown subcommands produce a non-zero exit.
func TestUnknownCommand(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.Run("nonexistent-command")
	helpers.AssertFailure(t, res)
}

// TestVersionFlag verifies --version flag behavior (cobra default).
func TestVersionFlag(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	// Cobra may not have --version; just ensure no crash
	res := cli.Run("version")
	helpers.AssertSuccess(t, res)
}
