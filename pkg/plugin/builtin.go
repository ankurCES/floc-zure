package plugin

import (
	"fmt"
	"sync"
	"time"
)

// BuiltinStorageProvider implements ResourceProvider for Microsoft.Storage/*.
type BuiltinStorageProvider struct {
	mu        sync.RWMutex
	resources map[string]map[string]interface{} // name -> resource
}

// NewBuiltinStorageProvider creates the built-in storage provider.
func NewBuiltinStorageProvider() *BuiltinStorageProvider {
	return &BuiltinStorageProvider{
		resources: make(map[string]map[string]interface{}),
	}
}

func (p *BuiltinStorageProvider) Name() string { return "storage" }

func (p *BuiltinStorageProvider) ResourceTypes() []string {
	return []string{
		"Microsoft.Storage/storageAccounts",
	}
}

func (p *BuiltinStorageProvider) Create(resourceType, name, rg, location string, props map[string]interface{}) (map[string]interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.resources[name]; exists {
		return nil, fmt.Errorf("storage account %q already exists", name)
	}
	res := map[string]interface{}{
		"id":                fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s", rg, name),
		"name":              name,
		"type":              resourceType,
		"location":          location,
		"resourceGroup":     rg,
		"provisioningState": "Succeeded",
		"creationTime":      time.Now().UTC().Format(time.RFC3339),
		"properties":        props,
	}
	p.resources[name] = res
	return res, nil
}

func (p *BuiltinStorageProvider) Get(resourceType, name string) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	res, ok := p.resources[name]
	if !ok {
		return nil, nil
	}
	return res, nil
}

func (p *BuiltinStorageProvider) List(resourceType, rg string) ([]map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var result []map[string]interface{}
	for _, res := range p.resources {
		if rg == "" || res["resourceGroup"] == rg {
			result = append(result, res)
		}
	}
	return result, nil
}

func (p *BuiltinStorageProvider) Delete(resourceType, name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.resources[name]; !ok {
		return fmt.Errorf("storage account %q not found", name)
	}
	delete(p.resources, name)
	return nil
}

// BuiltinComputeProvider implements ResourceProvider for Microsoft.Compute/*.
type BuiltinComputeProvider struct {
	mu        sync.RWMutex
	resources map[string]map[string]interface{}
}

// NewBuiltinComputeProvider creates the built-in compute provider.
func NewBuiltinComputeProvider() *BuiltinComputeProvider {
	return &BuiltinComputeProvider{
		resources: make(map[string]map[string]interface{}),
	}
}

func (p *BuiltinComputeProvider) Name() string { return "compute" }

func (p *BuiltinComputeProvider) ResourceTypes() []string {
	return []string{
		"Microsoft.Compute/virtualMachines",
	}
}

func (p *BuiltinComputeProvider) Create(resourceType, name, rg, location string, props map[string]interface{}) (map[string]interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.resources[name]; exists {
		return nil, fmt.Errorf("VM %q already exists", name)
	}
	res := map[string]interface{}{
		"id":                fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", rg, name),
		"name":              name,
		"type":              resourceType,
		"location":          location,
		"resourceGroup":     rg,
		"provisioningState": "Succeeded",
		"powerState":        "Running",
		"properties":        props,
	}
	p.resources[name] = res
	return res, nil
}

func (p *BuiltinComputeProvider) Get(resourceType, name string) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	res, ok := p.resources[name]
	if !ok {
		return nil, nil
	}
	return res, nil
}

func (p *BuiltinComputeProvider) List(resourceType, rg string) ([]map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var result []map[string]interface{}
	for _, res := range p.resources {
		if rg == "" || res["resourceGroup"] == rg {
			result = append(result, res)
		}
	}
	return result, nil
}

func (p *BuiltinComputeProvider) Delete(resourceType, name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.resources[name]; !ok {
		return fmt.Errorf("VM %q not found", name)
	}
	delete(p.resources, name)
	return nil
}

// RegisterBuiltinProviders registers all built-in providers with the registry.
func RegisterBuiltinProviders(reg *Registry) error {
	providers := []ResourceProvider{
		NewBuiltinStorageProvider(),
		NewBuiltinComputeProvider(),
	}
	for _, p := range providers {
		if err := reg.Register(p); err != nil {
			return fmt.Errorf("register built-in provider %q: %w", p.Name(), err)
		}
	}
	return nil
}
