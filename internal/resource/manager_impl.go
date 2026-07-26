// Package resource implements Azure resource lifecycle operations via the CLI.
package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/pkg/models"
	"github.com/ankurCES/floc-zure/pkg/utils"
)

// ManagerImpl delegates resource operations to the Azure CLI executor.
type ManagerImpl struct {
	exec azure.CLIExecutor
}

// NewManager creates a resource manager backed by the given CLI executor.
func NewManager(exec azure.CLIExecutor) *ManagerImpl {
	return &ManagerImpl{exec: exec}
}

// --- Resource Groups ---

func (m *ManagerImpl) CreateResourceGroup(ctx context.Context, name, location string, tags map[string]string) (*models.ResourceGroup, error) {
	if name == "" {
		return nil, &utils.ValidationError{Field: "name", Message: "resource group name is required"}
	}
	if location == "" {
		return nil, &utils.ValidationError{Field: "location", Message: "location is required"}
	}

	args := []string{"group", "create", "--name", name, "--location", location}
	if len(tags) > 0 {
		args = append(args, "--tags")
		for k, v := range tags {
			args = append(args, fmt.Sprintf("%s=%s", k, v))
		}
	}

	var rg models.ResourceGroup
	if err := m.exec.RunJSON(ctx, &rg, args...); err != nil {
		return nil, fmt.Errorf("create resource group %q: %w", name, err)
	}
	return &rg, nil
}

func (m *ManagerImpl) GetResourceGroup(ctx context.Context, name string) (*models.ResourceGroup, error) {
	if name == "" {
		return nil, &utils.ValidationError{Field: "name", Message: "resource group name is required"}
	}
	var rg models.ResourceGroup
	if err := m.exec.RunJSON(ctx, &rg, "group", "show", "--name", name); err != nil {
		return nil, fmt.Errorf("get resource group %q: %w", name, err)
	}
	return &rg, nil
}

func (m *ManagerImpl) ListResourceGroups(ctx context.Context) ([]models.ResourceGroup, error) {
	var groups []models.ResourceGroup
	if err := m.exec.RunJSON(ctx, &groups, "group", "list"); err != nil {
		return nil, fmt.Errorf("list resource groups: %w", err)
	}
	return groups, nil
}

func (m *ManagerImpl) DeleteResourceGroup(ctx context.Context, name string) error {
	if name == "" {
		return &utils.ValidationError{Field: "name", Message: "resource group name is required"}
	}
	_, err := m.exec.Run(ctx, "group", "delete", "--name", name, "--yes", "--no-wait")
	if err != nil {
		return fmt.Errorf("delete resource group %q: %w", name, err)
	}
	return nil
}

// --- Generic Resources ---

func (m *ManagerImpl) ListResources(ctx context.Context, resourceGroup string) ([]models.Resource, error) {
	args := []string{"resource", "list"}
	if resourceGroup != "" {
		args = append(args, "--resource-group", resourceGroup)
	}
	var resources []models.Resource
	if err := m.exec.RunJSON(ctx, &resources, args...); err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	return resources, nil
}

func (m *ManagerImpl) GetResource(ctx context.Context, resourceID string) (*models.Resource, error) {
	if resourceID == "" {
		return nil, &utils.ValidationError{Field: "resourceID", Message: "resource ID is required"}
	}
	var res models.Resource
	if err := m.exec.RunJSON(ctx, &res, "resource", "show", "--ids", resourceID); err != nil {
		return nil, fmt.Errorf("get resource %q: %w", resourceID, err)
	}
	return &res, nil
}

func (m *ManagerImpl) DeleteResource(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return &utils.ValidationError{Field: "resourceID", Message: "resource ID is required"}
	}
	_, err := m.exec.Run(ctx, "resource", "delete", "--ids", resourceID, "--yes")
	if err != nil {
		return fmt.Errorf("delete resource %q: %w", resourceID, err)
	}
	return nil
}

func (m *ManagerImpl) TagResource(ctx context.Context, resourceID string, tags map[string]string) error {
	if resourceID == "" {
		return &utils.ValidationError{Field: "resourceID", Message: "resource ID is required"}
	}
	if len(tags) == 0 {
		return &utils.ValidationError{Field: "tags", Message: "at least one tag is required"}
	}

	parts := make([]string, 0, len(tags))
	for k, v := range tags {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	tagStr := strings.Join(parts, " ")

	_, err := m.exec.Run(ctx, "resource", "tag", "--ids", resourceID, "--tags", tagStr)
	if err != nil {
		return fmt.Errorf("tag resource %q: %w", resourceID, err)
	}
	return nil
}
