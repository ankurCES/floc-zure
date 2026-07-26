package arm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/pkg/models"
)

// mockExecutor records CLI calls and returns success.
type mockExecutor struct {
	calls [][]string
}

func (m *mockExecutor) Run(_ context.Context, args ...string) (*azure.CLIResult, error) {
	m.calls = append(m.calls, args)
	return &azure.CLIResult{Stdout: "{}", ExitCode: 0}, nil
}

func (m *mockExecutor) RunJSON(_ context.Context, v interface{}, args ...string) error {
	m.calls = append(m.calls, args)
	return nil
}

func (m *mockExecutor) GetAccount(_ context.Context) (*models.AzureAccount, error) {
	return &models.AzureAccount{ID: "test"}, nil
}

func (m *mockExecutor) IsAuthenticated(_ context.Context) (bool, error) {
	return true, nil
}

func (m *mockExecutor) SetSubscription(_ context.Context, _ string) error {
	return nil
}

func TestDeploy_storageAccount(t *testing.T) {
	mock := &mockExecutor{}
	deployer := NewDeployer(mock)

	result := &ParseResult{
		Template: &Template{},
		Resources: []ResolvedResource{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "mystorage",
				Location: "westus2",
				Kind:     "StorageV2",
				SKU:      map[string]interface{}{"name": "Standard_LRS"},
				Tags:     map[string]string{"env": "test"},
			},
		},
	}

	dep, err := deployer.Deploy(context.Background(), result, "test-deploy", "my-rg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dep.ProvisioningState != "Succeeded" {
		t.Errorf("expected Succeeded, got %s", dep.ProvisioningState)
	}
	if len(dep.Resources) != 1 {
		t.Fatalf("expected 1 deployed resource, got %d", len(dep.Resources))
	}
	if dep.Resources[0].Name != "mystorage" {
		t.Errorf("expected name 'mystorage', got %q", dep.Resources[0].Name)
	}

	// Verify CLI calls: group create + storage account create
	if len(mock.calls) != 2 {
		t.Fatalf("expected 2 CLI calls, got %d", len(mock.calls))
	}
	// First call: group create
	if mock.calls[0][0] != "group" || mock.calls[0][1] != "create" {
		t.Errorf("expected 'group create', got %v", mock.calls[0][:2])
	}
	// Second call: storage account create
	if mock.calls[1][0] != "storage" || mock.calls[1][1] != "account" || mock.calls[1][2] != "create" {
		t.Errorf("expected 'storage account create', got %v", mock.calls[1][:3])
	}
}

func TestDeploy_multipleResources(t *testing.T) {
	mock := &mockExecutor{}
	deployer := NewDeployer(mock)

	result := &ParseResult{
		Template: &Template{},
		Resources: []ResolvedResource{
			{Type: "Microsoft.Storage/storageAccounts", Name: "sa1", Location: "eastus"},
			{Type: "Microsoft.KeyVault/vaults", Name: "kv1", Location: "eastus"},
			{Type: "Microsoft.Network/virtualNetworks", Name: "vnet1", Location: "eastus"},
		},
	}

	dep, err := deployer.Deploy(context.Background(), result, "multi-deploy", "rg1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dep.ProvisioningState != "Succeeded" {
		t.Errorf("expected Succeeded, got %s", dep.ProvisioningState)
	}
	if len(dep.Resources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(dep.Resources))
	}
	// 1 group create + 3 resource creates = 4 calls
	if len(mock.calls) != 4 {
		t.Errorf("expected 4 CLI calls, got %d", len(mock.calls))
	}
}

func TestValidate_valid(t *testing.T) {
	deployer := NewDeployer(&mockExecutor{})
	result := &ParseResult{
		Resources: []ResolvedResource{
			{Type: "Microsoft.Storage/storageAccounts", Name: "sa1", Location: "eastus"},
		},
	}
	if err := deployer.Validate(result, "rg1"); err != nil {
		t.Fatalf("expected valid, got error: %v", err)
	}
}

func TestValidate_noName(t *testing.T) {
	deployer := NewDeployer(&mockExecutor{})
	result := &ParseResult{
		Resources: []ResolvedResource{
			{Type: "Microsoft.Storage/storageAccounts", Name: "", Location: "eastus"},
		},
	}
	err := deployer.Validate(result, "rg1")
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidate_unsupportedType(t *testing.T) {
	deployer := NewDeployer(&mockExecutor{})
	result := &ParseResult{
		Resources: []ResolvedResource{
			{Type: "Microsoft.Foo/bars", Name: "test", Location: "eastus"},
		},
	}
	err := deployer.Validate(result, "rg1")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error, got %q", err.Error())
	}
}

// failExecutor fails on a specific call pattern.
type failExecutor struct {
	failOn string
	calls  [][]string
}

func (f *failExecutor) Run(_ context.Context, args ...string) (*azure.CLIResult, error) {
	f.calls = append(f.calls, args)
	joined := strings.Join(args, " ")
	if strings.Contains(joined, f.failOn) {
		return nil, fmt.Errorf("simulated failure on %q", f.failOn)
	}
	return &azure.CLIResult{Stdout: "{}", ExitCode: 0}, nil
}

func (f *failExecutor) RunJSON(_ context.Context, v interface{}, args ...string) error {
	return nil
}

func (f *failExecutor) GetAccount(_ context.Context) (*models.AzureAccount, error) {
	return &models.AzureAccount{ID: "test"}, nil
}

func (f *failExecutor) IsAuthenticated(_ context.Context) (bool, error) {
	return true, nil
}

func (f *failExecutor) SetSubscription(_ context.Context, _ string) error {
	return nil
}

func TestDeploy_failedResource(t *testing.T) {
	exec := &failExecutor{failOn: "keyvault"}
	deployer := NewDeployer(exec)

	result := &ParseResult{
		Template: &Template{},
		Resources: []ResolvedResource{
			{Type: "Microsoft.Storage/storageAccounts", Name: "sa1", Location: "eastus"},
			{Type: "Microsoft.KeyVault/vaults", Name: "kv1", Location: "eastus"},
		},
	}

	dep, err := deployer.Deploy(context.Background(), result, "fail-deploy", "rg1")
	if err == nil {
		t.Fatal("expected deployment error")
	}
	if dep.ProvisioningState != "Failed" {
		t.Errorf("expected Failed, got %s", dep.ProvisioningState)
	}
	if dep.Error == "" {
		t.Error("expected error message in deployment")
	}
}

func TestDeploymentToJSON(t *testing.T) {
	dep := &Deployment{
		Name:              "test",
		ResourceGroup:     "rg1",
		ProvisioningState: "Succeeded",
	}
	data, err := DeploymentToJSON(dep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "Succeeded") {
		t.Error("expected 'Succeeded' in JSON output")
	}
}
