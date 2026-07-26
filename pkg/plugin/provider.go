// Package plugin defines the resource provider interface for extending
// the azfloci simulator with custom resource types.
//
// Built-in providers handle Microsoft.Storage, Microsoft.KeyVault, etc.
// Third-party providers can register additional resource types by implementing
// the ResourceProvider interface and registering with the Registry.
package plugin

import (
	"fmt"
	"strings"
	"sync"
)

// ResourceProvider is the interface that resource type handlers must implement.
// Each provider handles one or more Azure resource types (e.g. "Microsoft.Storage/storageAccounts").
type ResourceProvider interface {
	// Name returns the provider's unique identifier (e.g. "storage", "keyvault").
	Name() string

	// ResourceTypes returns the ARM resource types this provider handles.
	ResourceTypes() []string

	// Create creates a resource of the given type with the given properties.
	// Returns the created resource as a JSON-serializable map.
	Create(resourceType, name, resourceGroup, location string, properties map[string]interface{}) (map[string]interface{}, error)

	// Get retrieves a resource by name. Returns nil if not found.
	Get(resourceType, name string) (map[string]interface{}, error)

	// List returns all resources of the given type, optionally filtered by resource group.
	List(resourceType, resourceGroup string) ([]map[string]interface{}, error)

	// Delete removes a resource by name.
	Delete(resourceType, name string) error
}

// Registry holds all registered resource providers and routes resource
// operations to the correct provider based on the ARM resource type.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ResourceProvider // name -> provider
	typeIndex map[string]string           // lowercase(resourceType) -> provider name
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ResourceProvider),
		typeIndex: make(map[string]string),
	}
}

// Register adds a provider to the registry. Panics if any of its resource
// types are already registered by another provider.
func (r *Registry) Register(p ResourceProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}

	for _, rt := range p.ResourceTypes() {
		key := strings.ToLower(rt)
		if existingProvider, ok := r.typeIndex[key]; ok {
			return fmt.Errorf("resource type %q already registered by provider %q", rt, existingProvider)
		}
		r.typeIndex[key] = name
	}

	r.providers[name] = p
	return nil
}

// Unregister removes a provider and all its type mappings.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.providers[name]
	if !ok {
		return fmt.Errorf("provider %q not found", name)
	}

	for _, rt := range p.ResourceTypes() {
		delete(r.typeIndex, strings.ToLower(rt))
	}
	delete(r.providers, name)
	return nil
}

// GetProvider returns the provider registered for the given resource type.
func (r *Registry) GetProvider(resourceType string) (ResourceProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := strings.ToLower(resourceType)
	name, ok := r.typeIndex[key]
	if !ok {
		return nil, false
	}
	return r.providers[name], true
}

// GetProviderByName returns a provider by its name.
func (r *Registry) GetProviderByName(name string) (ResourceProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// ListProviders returns all registered provider names.
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListResourceTypes returns all registered resource types.
func (r *Registry) ListResourceTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.typeIndex))
	for rt := range r.typeIndex {
		types = append(types, rt)
	}
	return types
}

// Create routes a create operation to the appropriate provider.
func (r *Registry) Create(resourceType, name, resourceGroup, location string, properties map[string]interface{}) (map[string]interface{}, error) {
	p, ok := r.GetProvider(resourceType)
	if !ok {
		return nil, fmt.Errorf("no provider registered for resource type %q", resourceType)
	}
	return p.Create(resourceType, name, resourceGroup, location, properties)
}

// Get routes a get operation to the appropriate provider.
func (r *Registry) Get(resourceType, name string) (map[string]interface{}, error) {
	p, ok := r.GetProvider(resourceType)
	if !ok {
		return nil, fmt.Errorf("no provider registered for resource type %q", resourceType)
	}
	return p.Get(resourceType, name)
}

// List routes a list operation to the appropriate provider.
func (r *Registry) List(resourceType, resourceGroup string) ([]map[string]interface{}, error) {
	p, ok := r.GetProvider(resourceType)
	if !ok {
		return nil, fmt.Errorf("no provider registered for resource type %q", resourceType)
	}
	return p.List(resourceType, resourceGroup)
}

// Delete routes a delete operation to the appropriate provider.
func (r *Registry) Delete(resourceType, name string) error {
	p, ok := r.GetProvider(resourceType)
	if !ok {
		return fmt.Errorf("no provider registered for resource type %q", resourceType)
	}
	return p.Delete(resourceType, name)
}
