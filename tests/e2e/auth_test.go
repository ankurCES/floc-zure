//go:build e2e

// Tests for authentication flow: verifies azfloci can detect Azure auth state.
package e2e

import (
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

// TestAuthStatusAuthenticated checks `azfloci auth status` when logged in.
// Skips if az CLI unavailable or user not authenticated.
func TestAuthStatusAuthenticated(t *testing.T) {
	helpers.RequireAzCLI(t)
	helpers.RequireAzAuth(t)

	cli := helpers.NewCLIRunner(t)
	res := cli.Run("auth", "status")

	helpers.AssertSuccess(t, res)
	helpers.AssertStdoutContains(t, res, "Authenticated")
	helpers.AssertStdoutContains(t, res, "Subscription")
	helpers.AssertStdoutContains(t, res, "Tenant")
}

// TestAuthStatusShowsUser verifies user info is printed.
func TestAuthStatusShowsUser(t *testing.T) {
	helpers.RequireAzCLI(t)
	helpers.RequireAzAuth(t)

	cli := helpers.NewCLIRunner(t)
	res := cli.MustRun(t, "auth", "status")
	helpers.AssertStdoutContains(t, res, "User")
}

// TestAuthSubcommandHelp verifies `azfloci auth --help` works.
func TestAuthSubcommandHelp(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	res := cli.MustRun(t, "auth", "--help")
	helpers.AssertStdoutContains(t, res, "authentication")
}
