// Package state provides a JSON-file-backed state store for the Azure simulator.
// It holds subscriptions, resource groups, and resources in memory and persists
// to disk after every mutation.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Subscription mirrors az account show JSON output.
type Subscription struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	State           string `json:"state"`
	IsDefault       bool   `json:"isDefault"`
	TenantID        string `json:"tenantId"`
	HomeTenantID    string `json:"homeTenantId"`
	EnvironmentName string `json:"environmentName"`
	User            User   `json:"user"`
}

// User is the user block inside a subscription.
type User struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ResourceGroup mirrors az group show JSON output.
type ResourceGroup struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Location   string                `json:"location"`
	Tags       map[string]string     `json:"tags,omitempty"`
	Properties ResourceGroupProps    `json:"properties"`
	ManagedBy  string                `json:"managedBy,omitempty"`
	Type       string                `json:"type"`
}

// ResourceGroupProps holds provisioning state.
type ResourceGroupProps struct {
	ProvisioningState string `json:"provisioningState"`
}

// Resource mirrors az resource show JSON output.
type Resource struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Location   string                 `json:"location"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Kind       string                 `json:"kind,omitempty"`
	SKU        map[string]interface{} `json:"sku,omitempty"`
}

// Data is the top-level persisted state.
type Data struct {
	ActiveSubscription string                   `json:"active_subscription"`
	Subscriptions      []Subscription           `json:"subscriptions"`
	ResourceGroups     map[string]*ResourceGroup `json:"resource_groups"`
	Resources          map[string]*Resource      `json:"resources"`
}

// Store is a thread-safe, JSON-file-backed Azure state store.
type Store struct {
	mu       sync.RWMutex
	data     *Data
	filePath string
}

// DefaultStatePath returns ~/.azfloci-sim/state.json, honoring AZFLOCI_SIM_STATE.
func DefaultStatePath() string {
	if p := os.Getenv("AZFLOCI_SIM_STATE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".azfloci-sim", "state.json")
}

// NewStore loads state from filePath, seeding defaults if the file doesn't exist.
func NewStore(filePath string) (*Store, error) {
	s := &Store{filePath: filePath}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.data = seedData()
			return s, s.persist()
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var d Data
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}
	if d.ResourceGroups == nil {
		d.ResourceGroups = make(map[string]*ResourceGroup)
	}
	if d.Resources == nil {
		d.Resources = make(map[string]*Resource)
	}
	s.data = &d
	return s, nil
}

// seedData creates the initial simulated Azure state.
func seedData() *Data {
	subID := "00000000-0000-0000-0000-000000000001"
	tenantID := "00000000-0000-0000-0000-000000000099"
	return &Data{
		ActiveSubscription: subID,
		Subscriptions: []Subscription{
			{
				ID:              subID,
				Name:            "Simulated-Subscription-1",
				State:           "Enabled",
				IsDefault:       true,
				TenantID:        tenantID,
				HomeTenantID:    tenantID,
				EnvironmentName: "AzureCloud",
				User:            User{Name: "simulator@azfloci.local", Type: "user"},
			},
		},
		ResourceGroups: make(map[string]*ResourceGroup),
		Resources:      make(map[string]*Resource),
	}
}

// persist writes the current state to disk atomically (write-tmp + rename).
func (s *Store) persist() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("write state tmp: %w", err)
	}
	return os.Rename(tmp, s.filePath)
}

// --- Subscriptions ---

// GetActiveSubscription returns the currently active subscription.
func (s *Store) GetActiveSubscription() *Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.data.Subscriptions {
		if s.data.Subscriptions[i].ID == s.data.ActiveSubscription {
			sub := s.data.Subscriptions[i]
			return &sub
		}
	}
	return nil
}

// ListSubscriptions returns all subscriptions.
func (s *Store) ListSubscriptions() []Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Subscription, len(s.data.Subscriptions))
	copy(out, s.data.Subscriptions)
	return out
}

// SetActiveSubscription switches to the given subscription ID.
func (s *Store) SetActiveSubscription(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Subscriptions {
		if s.data.Subscriptions[i].ID == id {
			// Clear all isDefault, set new one.
			for j := range s.data.Subscriptions {
				s.data.Subscriptions[j].IsDefault = false
			}
			s.data.Subscriptions[i].IsDefault = true
			s.data.ActiveSubscription = id
			return s.persist()
		}
	}
	return fmt.Errorf("subscription '%s' not found", id)
}

// --- Resource Groups ---

// CreateResourceGroup creates or updates a resource group.
func (s *Store) CreateResourceGroup(name, location string, tags map[string]string) *ResourceGroup {
	s.mu.Lock()
	defer s.mu.Unlock()

	subID := s.data.ActiveSubscription
	rg := &ResourceGroup{
		ID:       fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subID, name),
		Name:     name,
		Location: location,
		Tags:     tags,
		Type:     "Microsoft.Resources/resourceGroups",
		Properties: ResourceGroupProps{
			ProvisioningState: "Succeeded",
		},
	}
	s.data.ResourceGroups[name] = rg
	_ = s.persist()
	return rg
}

// GetResourceGroup returns a resource group by name, or nil.
func (s *Store) GetResourceGroup(name string) *ResourceGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rg := s.data.ResourceGroups[name]
	if rg == nil {
		return nil
	}
	copy := *rg
	return &copy
}

// ListResourceGroups returns all resource groups.
func (s *Store) ListResourceGroups() []*ResourceGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ResourceGroup, 0, len(s.data.ResourceGroups))
	for _, rg := range s.data.ResourceGroups {
		copy := *rg
		out = append(out, &copy)
	}
	return out
}

// DeleteResourceGroup removes a resource group and all its resources.
func (s *Store) DeleteResourceGroup(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.ResourceGroups[name]; !ok {
		return fmt.Errorf("resource group '%s' could not be found", name)
	}
	delete(s.data.ResourceGroups, name)

	// Cascade: delete resources belonging to this group.
	subID := s.data.ActiveSubscription
	prefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/", subID, name)
	for id := range s.data.Resources {
		if len(id) >= len(prefix) && id[:len(prefix)] == prefix {
			delete(s.data.Resources, id)
		}
	}
	return s.persist()
}

// --- Resources ---

// AddResource adds a resource to the store.
func (s *Store) AddResource(res *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Resources[res.ID] = res
	_ = s.persist()
}

// GetResource returns a resource by ARM ID, or nil.
func (s *Store) GetResource(id string) *Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r := s.data.Resources[id]
	if r == nil {
		return nil
	}
	copy := *r
	return &copy
}

// ListResources returns resources, optionally filtered by resource group name.
func (s *Store) ListResources(resourceGroup string) []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Resource, 0)
	subID := s.data.ActiveSubscription
	for _, r := range s.data.Resources {
		if resourceGroup != "" {
			prefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/", subID, resourceGroup)
			if len(r.ID) < len(prefix) || r.ID[:len(prefix)] != prefix {
				continue
			}
		}
		copy := *r
		out = append(out, &copy)
	}
	return out
}

// DeleteResource removes a resource by ARM ID.
func (s *Store) DeleteResource(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Resources[id]; !ok {
		return fmt.Errorf("resource '%s' could not be found", id)
	}
	delete(s.data.Resources, id)
	return s.persist()
}

// TagResource merges tags into an existing resource.
func (s *Store) TagResource(id string, tags map[string]string) (*Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data.Resources[id]
	if !ok {
		return nil, fmt.Errorf("resource '%s' could not be found", id)
	}
	if r.Tags == nil {
		r.Tags = make(map[string]string)
	}
	for k, v := range tags {
		r.Tags[k] = v
	}
	_ = s.persist()
	copy := *r
	return &copy, nil
}

// Reset clears all state and re-seeds defaults. Useful for test isolation.
func (s *Store) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = seedData()
	return s.persist()
}

// FilePath returns the path to the state file.
func (s *Store) FilePath() string {
	return s.filePath
}

// GenerateResourceID builds a deterministic ARM resource ID.
func GenerateResourceID(subID, rgName, provider, resourceType, name string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s",
		subID, rgName, provider, resourceType, name)
}

// Timestamp returns an Azure-style timestamp string.
func Timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
