// Package resource manages Azure resource lifecycle operations.
package resource

import (
	"context"

	"github.com/ankurCES/floc-zure/pkg/models"
)

// Manager provides CRUD operations for Azure resources via the CLI.
type Manager interface {
	// --- Resource Groups ---

	// CreateResourceGroup creates a resource group.
	CreateResourceGroup(ctx context.Context, name, location string, tags map[string]string) (*models.ResourceGroup, error)

	// GetResourceGroup returns a resource group by name.
	GetResourceGroup(ctx context.Context, name string) (*models.ResourceGroup, error)

	// ListResourceGroups lists all resource groups in the active subscription.
	ListResourceGroups(ctx context.Context) ([]models.ResourceGroup, error)

	// DeleteResourceGroup deletes a resource group (and all its resources).
	DeleteResourceGroup(ctx context.Context, name string) error

	// --- Generic Resources ---

	// ListResources lists resources, optionally filtered by resource group.
	ListResources(ctx context.Context, resourceGroup string) ([]models.Resource, error)

	// GetResource returns a resource by its full ARM resource ID.
	GetResource(ctx context.Context, resourceID string) (*models.Resource, error)

	// DeleteResource deletes a resource by its full ARM resource ID.
	DeleteResource(ctx context.Context, resourceID string) error

	// TagResource adds/updates tags on a resource.
	TagResource(ctx context.Context, resourceID string, tags map[string]string) error
}
