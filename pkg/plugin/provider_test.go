package plugin

import (
	"fmt"
	"sort"
	"testing"
)

// testProvider is a minimal custom provider for testing.
type testProvider struct {
	name  string
	types []string
	data  map[string]map[string]interface{}
}

func newTestProvider(name string, types []string) *testProvider {
	return &testProvider{name: name, types: types, data: make(map[string]map[string]interface{})}
}

func (p *testProvider) Name() string              { return p.name }
func (p *testProvider) ResourceTypes() []string    { return p.types }

func (p *testProvider) Create(rt, name, rg, loc string, props map[string]interface{}) (map[string]interface{}, error) {
	if _, ok := p.data[name]; ok {
		return nil, fmt.Errorf("%s %q already exists", rt, name)
	}
	res := map[string]interface{}{"name": name, "type": rt, "resourceGroup": rg, "location": loc, "properties": props}
	p.data[name] = res
	return res, nil
}

func (p *testProvider) Get(rt, name string) (map[string]interface{}, error) {
	return p.data[name], nil
}

func (p *testProvider) List(rt, rg string) ([]map[string]interface{}, error) {
	var out []map[string]interface{}
	for _, r := range p.data {
		if rg == "" || r["resourceGroup"] == rg {
			out = append(out, r)
		}
	}
	return out, nil
}

func (p *testProvider) Delete(rt, name string) error {
	if _, ok := p.data[name]; !ok {
		return fmt.Errorf("%q not found", name)
	}
	delete(p.data, name)
	return nil
}

func TestRegistry_RegisterAndLookup(t *testing.T) {
	reg := NewRegistry()
	p := newTestProvider("custom", []string{"Microsoft.Custom/widgets"})
	if err := reg.Register(p); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	found, ok := reg.GetProvider("Microsoft.Custom/widgets")
	if !ok {
		t.Fatal("expected to find provider")
	}
	if found.Name() != "custom" {
		t.Errorf("expected 'custom', got %q", found.Name())
	}
}

func TestRegistry_CaseInsensitiveLookup(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newTestProvider("custom", []string{"Microsoft.Custom/Widgets"}))

	_, ok := reg.GetProvider("microsoft.custom/widgets")
	if !ok {
		t.Fatal("case-insensitive lookup should work")
	}
}

func TestRegistry_DuplicateProvider(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newTestProvider("a", []string{"Microsoft.A/items"}))
	err := reg.Register(newTestProvider("a", []string{"Microsoft.B/items"}))
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func TestRegistry_DuplicateResourceType(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newTestProvider("a", []string{"Microsoft.X/items"}))
	err := reg.Register(newTestProvider("b", []string{"Microsoft.X/items"}))
	if err == nil {
		t.Fatal("expected duplicate resource type error")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newTestProvider("a", []string{"Microsoft.A/items"}))
	if err := reg.Unregister("a"); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}
	_, ok := reg.GetProvider("Microsoft.A/items")
	if ok {
		t.Fatal("expected provider to be removed")
	}
}

func TestRegistry_UnregisterNotFound(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Unregister("nope"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRegistry_CRUD(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newTestProvider("custom", []string{"Microsoft.Custom/widgets"}))

	// Create
	res, err := reg.Create("Microsoft.Custom/widgets", "w1", "rg1", "eastus", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res["name"] != "w1" {
		t.Errorf("expected name 'w1', got %v", res["name"])
	}

	// Get
	got, err := reg.Get("Microsoft.Custom/widgets", "w1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// List
	list, err := reg.List("Microsoft.Custom/widgets", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1, got %d", len(list))
	}

	// Delete
	if err := reg.Delete("Microsoft.Custom/widgets", "w1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ = reg.Get("Microsoft.Custom/widgets", "w1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestRegistry_NoProvider(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Create("Microsoft.Unknown/foo", "x", "rg", "eastus", nil)
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestRegistry_ListProviders(t *testing.T) {
	reg := NewRegistry()
	reg.Register(newTestProvider("alpha", []string{"Microsoft.A/x"}))
	reg.Register(newTestProvider("beta", []string{"Microsoft.B/x"}))
	names := reg.ListProviders()
	sort.Strings(names)
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("unexpected providers: %v", names)
	}
}

func TestBuiltinProviders(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterBuiltinProviders(reg); err != nil {
		t.Fatalf("RegisterBuiltinProviders: %v", err)
	}

	// Storage should be registered
	p, ok := reg.GetProvider("Microsoft.Storage/storageAccounts")
	if !ok {
		t.Fatal("storage provider not registered")
	}
	if p.Name() != "storage" {
		t.Errorf("expected 'storage', got %q", p.Name())
	}

	// Compute should be registered
	_, ok = reg.GetProvider("Microsoft.Compute/virtualMachines")
	if !ok {
		t.Fatal("compute provider not registered")
	}

	// CRUD through built-in
	res, err := reg.Create("Microsoft.Storage/storageAccounts", "testsa", "rg1", "eastus", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if res["name"] != "testsa" {
		t.Errorf("expected 'testsa', got %v", res["name"])
	}
}

func TestBuiltinStorage_DuplicateCreate(t *testing.T) {
	p := NewBuiltinStorageProvider()
	p.Create("Microsoft.Storage/storageAccounts", "sa1", "rg", "eastus", nil)
	_, err := p.Create("Microsoft.Storage/storageAccounts", "sa1", "rg", "eastus", nil)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestBuiltinStorage_DeleteNotFound(t *testing.T) {
	p := NewBuiltinStorageProvider()
	err := p.Delete("Microsoft.Storage/storageAccounts", "nope")
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestBuiltinCompute_ListByRG(t *testing.T) {
	p := NewBuiltinComputeProvider()
	p.Create("Microsoft.Compute/virtualMachines", "vm1", "rg1", "eastus", nil)
	p.Create("Microsoft.Compute/virtualMachines", "vm2", "rg2", "westus", nil)

	list, _ := p.List("Microsoft.Compute/virtualMachines", "rg1")
	if len(list) != 1 {
		t.Errorf("expected 1, got %d", len(list))
	}

	allList, _ := p.List("Microsoft.Compute/virtualMachines", "")
	if len(allList) != 2 {
		t.Errorf("expected 2, got %d", len(allList))
	}
}
