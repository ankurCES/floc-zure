//go:build e2e

package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// UniqueResourceGroupName generates a unique resource group name for tests.
func UniqueResourceGroupName() string {
	return fmt.Sprintf("azfloci-e2e-%d-%04x", time.Now().Unix(), rand.Intn(0xFFFF))
}

// TestResourceGroup holds info about a test resource group and provides cleanup.
type TestResourceGroup struct {
	Name     string
	Location string
	t        *testing.T
}

// SetupResourceGroup creates a resource group via `az` CLI and registers
// a t.Cleanup that deletes it (best-effort) when the test finishes.
func SetupResourceGroup(t *testing.T, location string) *TestResourceGroup {
	t.Helper()
	RequireAzCLI(t)
	RequireAzAuth(t)

	name := UniqueResourceGroupName()
	if location == "" {
		location = "eastus"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := RunAzCLI(ctx, "group", "create",
		"--name", name,
		"--location", location,
		"--tags", "purpose=e2e-test", "created-by=azfloci-e2e",
		"--output", "json")
	if err != nil {
		t.Fatalf("failed to create resource group %s: %v\nstderr: %s", name, err, res.Stderr)
	}

	rg := &TestResourceGroup{Name: name, Location: location, t: t}

	// Register cleanup — delete the resource group when the test ends.
	t.Cleanup(func() {
		rg.Delete()
	})

	t.Logf("created test resource group: %s in %s", name, location)
	return rg
}

// Delete removes the test resource group. Called automatically via t.Cleanup.
func (rg *TestResourceGroup) Delete() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	res, err := RunAzCLI(ctx, "group", "delete",
		"--name", rg.Name,
		"--yes",
		"--no-wait")
	if err != nil {
		rg.t.Logf("WARNING: failed to delete resource group %s: %v\nstderr: %s", rg.Name, err, res.Stderr)
	} else {
		rg.t.Logf("deleted test resource group: %s", rg.Name)
	}
}

// RequireAzAuth skips the test if the user is not authenticated with Azure.
func RequireAzAuth(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := RunAzCLI(ctx, "account", "show", "--output", "json")
	if err != nil || res.ExitCode != 0 {
		t.Skip("skipping: not authenticated with Azure CLI (run 'az login')")
	}
}

// ResourceGroupExists checks via `az` CLI whether a resource group exists.
func ResourceGroupExists(t *testing.T, name string) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := RunAzCLI(ctx, "group", "exists", "--name", name)
	if err != nil {
		return false
	}
	return strings.TrimSpace(res.Stdout) == "true"
}

// GetResourceGroupTags returns the tags map for a resource group.
func GetResourceGroupTags(t *testing.T, name string) map[string]string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := RunAzCLI(ctx, "group", "show", "--name", name, "--output", "json")
	if err != nil {
		t.Fatalf("failed to get resource group %s: %v", name, err)
	}

	var rg struct {
		Tags map[string]string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &rg); err != nil {
		t.Fatalf("failed to parse resource group JSON: %v", err)
	}
	return rg.Tags
}

// ListResourcesInGroup returns resource names in a resource group.
func ListResourcesInGroup(t *testing.T, rgName string) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := RunAzCLI(ctx, "resource", "list",
		"--resource-group", rgName,
		"--output", "json")
	if err != nil {
		t.Fatalf("failed to list resources in %s: %v", rgName, err)
	}

	var resources []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &resources); err != nil {
		t.Fatalf("failed to parse resources JSON: %v", err)
	}

	names := make([]string, len(resources))
	for i, r := range resources {
		names[i] = r.Name
	}
	return names
}

// GetCurrentSubscription returns the active subscription ID.
func GetCurrentSubscription(t *testing.T) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := RunAzCLI(ctx, "account", "show", "--query", "id", "--output", "tsv")
	if err != nil {
		t.Fatalf("failed to get current subscription: %v", err)
	}
	return strings.TrimSpace(res.Stdout)
}
