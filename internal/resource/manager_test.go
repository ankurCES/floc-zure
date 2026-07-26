package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/pkg/models"
)

// mockExec is a test-only CLIExecutor.
type mockExec struct {
	runFunc     func(ctx context.Context, args ...string) (*azure.CLIResult, error)
	runJSONFunc func(ctx context.Context, dest interface{}, args ...string) error
	calls       [][]string
}

func (m *mockExec) Run(ctx context.Context, args ...string) (*azure.CLIResult, error) {
	m.calls = append(m.calls, args)
	if m.runFunc != nil {
		return m.runFunc(ctx, args...)
	}
	return &azure.CLIResult{}, nil
}

func (m *mockExec) RunJSON(ctx context.Context, dest interface{}, args ...string) error {
	m.calls = append(m.calls, args)
	if m.runJSONFunc != nil {
		return m.runJSONFunc(ctx, dest, args...)
	}
	return json.Unmarshal([]byte("{}"), dest)
}

func (m *mockExec) GetAccount(ctx context.Context) (*models.AzureAccount, error) {
	return &models.AzureAccount{ID: "mock"}, nil
}
func (m *mockExec) IsAuthenticated(ctx context.Context) (bool, error) { return true, nil }
func (m *mockExec) SetSubscription(ctx context.Context, id string) error { return nil }

func (m *mockExec) lastCall() []string {
	if len(m.calls) == 0 {
		return nil
	}
	return m.calls[len(m.calls)-1]
}

// --- Resource Group Tests ---

func TestCreateResourceGroup(t *testing.T) {
	mock := &mockExec{
		runJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			data := `{"id":"/subscriptions/s/resourceGroups/rg1","name":"rg1","location":"eastus"}`
			return json.Unmarshal([]byte(data), dest)
		},
	}
	mgr := NewManager(mock)
	rg, err := mgr.CreateResourceGroup(context.Background(), "rg1", "eastus", map[string]string{"env": "dev"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rg.Name != "rg1" {
		t.Errorf("expected name rg1, got %s", rg.Name)
	}
	// Verify CLI args contain tags
	call := mock.lastCall()
	joined := strings.Join(call, " ")
	if !strings.Contains(joined, "--tags") {
		t.Errorf("expected --tags in args: %v", call)
	}
	if !strings.Contains(joined, "env=dev") {
		t.Errorf("expected env=dev in args: %v", call)
	}
}

func TestCreateResourceGroup_EmptyName(t *testing.T) {
	mgr := NewManager(&mockExec{})
	_, err := mgr.CreateResourceGroup(context.Background(), "", "eastus", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected 'name' in error, got: %v", err)
	}
}

func TestCreateResourceGroup_EmptyLocation(t *testing.T) {
	mgr := NewManager(&mockExec{})
	_, err := mgr.CreateResourceGroup(context.Background(), "rg1", "", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetResourceGroup(t *testing.T) {
	mock := &mockExec{
		runJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			data := `{"name":"rg2","location":"westus2"}`
			return json.Unmarshal([]byte(data), dest)
		},
	}
	mgr := NewManager(mock)
	rg, err := mgr.GetResourceGroup(context.Background(), "rg2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rg.Location != "westus2" {
		t.Errorf("expected westus2, got %s", rg.Location)
	}
}

func TestGetResourceGroup_EmptyName(t *testing.T) {
	mgr := NewManager(&mockExec{})
	_, err := mgr.GetResourceGroup(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestListResourceGroups(t *testing.T) {
	mock := &mockExec{
		runJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			data := `[{"name":"rg1"},{"name":"rg2"}]`
			return json.Unmarshal([]byte(data), dest)
		},
	}
	mgr := NewManager(mock)
	groups, err := mgr.ListResourceGroups(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestDeleteResourceGroup(t *testing.T) {
	mock := &mockExec{}
	mgr := NewManager(mock)
	err := mgr.DeleteResourceGroup(context.Background(), "rg-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	call := mock.lastCall()
	joined := strings.Join(call, " ")
	if !strings.Contains(joined, "--yes") {
		t.Error("expected --yes flag in delete")
	}
	if !strings.Contains(joined, "--no-wait") {
		t.Error("expected --no-wait flag in delete")
	}
}

func TestDeleteResourceGroup_EmptyName(t *testing.T) {
	mgr := NewManager(&mockExec{})
	err := mgr.DeleteResourceGroup(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- Resource Tests ---

func TestListResources_NoFilter(t *testing.T) {
	mock := &mockExec{
		runJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			return json.Unmarshal([]byte(`[{"name":"vm1","type":"Microsoft.Compute/virtualMachines"}]`), dest)
		},
	}
	mgr := NewManager(mock)
	res, err := mgr.ListResources(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("expected 1 resource, got %d", len(res))
	}
	// Should NOT have --resource-group flag
	call := mock.lastCall()
	joined := strings.Join(call, " ")
	if strings.Contains(joined, "--resource-group") {
		t.Error("did not expect --resource-group when no filter")
	}
}

func TestListResources_WithFilter(t *testing.T) {
	mock := &mockExec{
		runJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			return json.Unmarshal([]byte(`[]`), dest)
		},
	}
	mgr := NewManager(mock)
	_, _ = mgr.ListResources(context.Background(), "my-rg")
	call := mock.lastCall()
	joined := strings.Join(call, " ")
	if !strings.Contains(joined, "--resource-group my-rg") {
		t.Errorf("expected --resource-group my-rg: %v", call)
	}
}

func TestGetResource(t *testing.T) {
	mock := &mockExec{
		runJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			return json.Unmarshal([]byte(`{"id":"/sub/rg/res1","name":"res1","type":"Microsoft.Storage/storageAccounts"}`), dest)
		},
	}
	mgr := NewManager(mock)
	res, err := mgr.GetResource(context.Background(), "/sub/rg/res1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Type != "Microsoft.Storage/storageAccounts" {
		t.Errorf("expected storage account type, got %s", res.Type)
	}
}

func TestGetResource_EmptyID(t *testing.T) {
	mgr := NewManager(&mockExec{})
	_, err := mgr.GetResource(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDeleteResource(t *testing.T) {
	mock := &mockExec{}
	mgr := NewManager(mock)
	err := mgr.DeleteResource(context.Background(), "/sub/rg/res1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(mock.lastCall(), " ")
	if !strings.Contains(joined, "--yes") {
		t.Error("expected --yes in delete")
	}
}

func TestDeleteResource_EmptyID(t *testing.T) {
	mgr := NewManager(&mockExec{})
	err := mgr.DeleteResource(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTagResource(t *testing.T) {
	mock := &mockExec{}
	mgr := NewManager(mock)
	err := mgr.TagResource(context.Background(), "/sub/rg/res1", map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(mock.lastCall(), " ")
	if !strings.Contains(joined, "--tags") {
		t.Error("expected --tags in tag command")
	}
}

func TestTagResource_EmptyID(t *testing.T) {
	mgr := NewManager(&mockExec{})
	err := mgr.TagResource(context.Background(), "", map[string]string{"a": "b"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTagResource_EmptyTags(t *testing.T) {
	mgr := NewManager(&mockExec{})
	err := mgr.TagResource(context.Background(), "/id", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateResourceGroup_CLIError(t *testing.T) {
	mock := &mockExec{
		runJSONFunc: func(ctx context.Context, dest interface{}, args ...string) error {
			return fmt.Errorf("az command failed")
		},
	}
	mgr := NewManager(mock)
	_, err := mgr.CreateResourceGroup(context.Background(), "rg1", "eastus", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "create resource group") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}
