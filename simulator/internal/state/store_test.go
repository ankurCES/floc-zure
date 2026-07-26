package state

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")
	s, err := NewStore(p)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestNewStore_SeedsDefaults(t *testing.T) {
	s := tempStore(t)
	subs := s.ListSubscriptions()
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].ID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("unexpected sub ID: %s", subs[0].ID)
	}
	if subs[0].User.Name != "simulator@azfloci.local" {
		t.Errorf("unexpected user: %s", subs[0].User.Name)
	}
}

func TestNewStore_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")

	s, _ := NewStore(p)
	s.CreateResourceGroup("rg1", "eastus", nil)

	// Re-open from same file.
	s2, err := NewStore(p)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	rg := s2.GetResourceGroup("rg1")
	if rg == nil {
		t.Fatal("expected rg1 to persist across re-open")
	}
}

func TestGetActiveSubscription(t *testing.T) {
	s := tempStore(t)
	sub := s.GetActiveSubscription()
	if sub == nil {
		t.Fatal("nil active subscription")
	}
	if sub.State != "Enabled" {
		t.Errorf("expected Enabled, got %s", sub.State)
	}
}

func TestSetActiveSubscription_NotFound(t *testing.T) {
	s := tempStore(t)
	err := s.SetActiveSubscription("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent sub")
	}
}

func TestResourceGroup_CRUD(t *testing.T) {
	s := tempStore(t)

	// Create
	rg := s.CreateResourceGroup("test-rg", "westus2", map[string]string{"env": "test"})
	if rg.Name != "test-rg" {
		t.Errorf("name: got %s", rg.Name)
	}
	if rg.Location != "westus2" {
		t.Errorf("location: got %s", rg.Location)
	}
	if rg.Tags["env"] != "test" {
		t.Errorf("tags: got %v", rg.Tags)
	}
	if rg.Properties.ProvisioningState != "Succeeded" {
		t.Errorf("state: got %s", rg.Properties.ProvisioningState)
	}

	// Get
	got := s.GetResourceGroup("test-rg")
	if got == nil {
		t.Fatal("GetResourceGroup returned nil")
	}

	// List
	all := s.ListResourceGroups()
	if len(all) != 1 {
		t.Fatalf("expected 1, got %d", len(all))
	}

	// Delete
	if err := s.DeleteResourceGroup("test-rg"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if s.GetResourceGroup("test-rg") != nil {
		t.Error("rg should be deleted")
	}
}

func TestDeleteResourceGroup_NotFound(t *testing.T) {
	s := tempStore(t)
	err := s.DeleteResourceGroup("nope")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteResourceGroup_CascadesResources(t *testing.T) {
	s := tempStore(t)
	s.CreateResourceGroup("rg1", "eastus", nil)
	sub := s.GetActiveSubscription()
	resID := GenerateResourceID(sub.ID, "rg1", "Microsoft.Storage", "storageAccounts", "sa1")
	s.AddResource(&Resource{
		ID:       resID,
		Name:     "sa1",
		Type:     "Microsoft.Storage/storageAccounts",
		Location: "eastus",
	})

	if err := s.DeleteResourceGroup("rg1"); err != nil {
		t.Fatal(err)
	}
	if s.GetResource(resID) != nil {
		t.Error("resource should be cascade-deleted")
	}
}

func TestResource_CRUD(t *testing.T) {
	s := tempStore(t)
	s.CreateResourceGroup("rg1", "eastus", nil)
	sub := s.GetActiveSubscription()
	resID := GenerateResourceID(sub.ID, "rg1", "Microsoft.Compute", "virtualMachines", "vm1")

	// Add
	s.AddResource(&Resource{
		ID:       resID,
		Name:     "vm1",
		Type:     "Microsoft.Compute/virtualMachines",
		Location: "eastus",
	})

	// Get
	r := s.GetResource(resID)
	if r == nil {
		t.Fatal("nil resource")
	}
	if r.Name != "vm1" {
		t.Errorf("name: %s", r.Name)
	}

	// List all
	all := s.ListResources("")
	if len(all) != 1 {
		t.Fatalf("expected 1, got %d", len(all))
	}

	// List filtered
	filtered := s.ListResources("rg1")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 in rg1, got %d", len(filtered))
	}
	empty := s.ListResources("rg-other")
	if len(empty) != 0 {
		t.Fatalf("expected 0 in rg-other, got %d", len(empty))
	}

	// Tag
	tagged, err := s.TagResource(resID, map[string]string{"owner": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if tagged.Tags["owner"] != "test" {
		t.Errorf("tag not set: %v", tagged.Tags)
	}

	// Delete
	if err := s.DeleteResource(resID); err != nil {
		t.Fatal(err)
	}
	if s.GetResource(resID) != nil {
		t.Error("should be deleted")
	}
}

func TestDeleteResource_NotFound(t *testing.T) {
	s := tempStore(t)
	err := s.DeleteResource("/nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTagResource_NotFound(t *testing.T) {
	s := tempStore(t)
	_, err := s.TagResource("/nonexistent", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReset(t *testing.T) {
	s := tempStore(t)
	s.CreateResourceGroup("rg1", "eastus", nil)
	if err := s.Reset(); err != nil {
		t.Fatal(err)
	}
	if len(s.ListResourceGroups()) != 0 {
		t.Error("expected empty after reset")
	}
}

func TestDefaultStatePath_EnvOverride(t *testing.T) {
	os.Setenv("AZFLOCI_SIM_STATE", "/custom/path.json")
	defer os.Unsetenv("AZFLOCI_SIM_STATE")
	if got := DefaultStatePath(); got != "/custom/path.json" {
		t.Errorf("expected /custom/path.json, got %s", got)
	}
}

func TestGenerateResourceID(t *testing.T) {
	id := GenerateResourceID("sub1", "rg1", "Microsoft.Storage", "storageAccounts", "sa1")
	expected := "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/sa1"
	if id != expected {
		t.Errorf("got %s", id)
	}
}
