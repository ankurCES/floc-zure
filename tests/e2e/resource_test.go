//go:build e2e

// Tests for resource CRUD lifecycle: create, list, get, tag, and delete
// Azure resource groups. These tests create real Azure resources and clean up
// via t.Cleanup / defer.
package e2e

import (
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

// TestResourceGroupLifecycle tests the full create → verify → delete cycle
// for a resource group using the az CLI directly (validates our helper infra).
func TestResourceGroupLifecycle(t *testing.T) {
	helpers.RequireAzCLI(t)
	helpers.RequireAzAuth(t)

	// SetupResourceGroup creates the RG and registers t.Cleanup to delete it.
	rg := helpers.SetupResourceGroup(t, "eastus")

	// Verify it exists via az CLI.
	if !helpers.ResourceGroupExists(t, rg.Name) {
		t.Fatalf("resource group %s should exist after creation", rg.Name)
	}

	// Verify tags were set.
	tags := helpers.GetResourceGroupTags(t, rg.Name)
	if tags["purpose"] != "e2e-test" {
		t.Errorf("expected tag purpose=e2e-test, got %v", tags)
	}

	// Verify it's empty initially.
	resources := helpers.ListResourcesInGroup(t, rg.Name)
	if len(resources) != 0 {
		t.Errorf("expected 0 resources in new group, got %d", len(resources))
	}
}

// TestResourceGroupExistsCheck validates the existence check helper.
func TestResourceGroupExistsCheck(t *testing.T) {
	helpers.RequireAzCLI(t)
	helpers.RequireAzAuth(t)

	// A random name should not exist.
	if helpers.ResourceGroupExists(t, "azfloci-nonexistent-rg-9999999") {
		t.Error("nonexistent resource group should return false")
	}
}

// TestGetCurrentSubscription verifies we can read the active subscription.
func TestGetCurrentSubscription(t *testing.T) {
	helpers.RequireAzCLI(t)
	helpers.RequireAzAuth(t)

	sub := helpers.GetCurrentSubscription(t)
	if sub == "" {
		t.Fatal("expected non-empty subscription ID")
	}
	t.Logf("active subscription: %s", sub)
}
