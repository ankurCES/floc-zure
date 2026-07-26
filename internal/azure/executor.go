// Package azure provides the Azure CLI integration layer.
package azure

import (
	"context"
	"encoding/json"

	"github.com/ankurCES/floc-zure/pkg/models"
)

// CLIResult holds the raw output of an Azure CLI command.
type CLIResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// CLIExecutor wraps Azure CLI invocations. Implementations must be safe
// for concurrent use. The interface is the seam for testing — swap in a
// mock executor to test business logic without shelling out to `az`.
type CLIExecutor interface {
	// Run executes an arbitrary `az` command with args and returns raw output.
	Run(ctx context.Context, args ...string) (*CLIResult, error)

	// RunJSON executes an `az` command, parses JSON stdout into dest.
	RunJSON(ctx context.Context, dest interface{}, args ...string) error

	// GetAccount returns the current Azure account (az account show).
	GetAccount(ctx context.Context) (*models.AzureAccount, error)

	// IsAuthenticated returns true if `az account show` succeeds.
	IsAuthenticated(ctx context.Context) (bool, error)

	// SetSubscription switches the active subscription.
	SetSubscription(ctx context.Context, subscriptionID string) error
}

// ParseJSON is a helper to unmarshal a CLIResult's stdout into a typed value.
func ParseJSON[T any](result *CLIResult) (*T, error) {
	var v T
	if err := json.Unmarshal([]byte(result.Stdout), &v); err != nil {
		return nil, err
	}
	return &v, nil
}
