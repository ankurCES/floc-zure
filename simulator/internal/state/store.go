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

// StorageAccount mirrors az storage account show JSON output.
type StorageAccount struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	Kind              string            `json:"kind"`
	SKU               StorageSKU        `json:"sku"`
	Tags              map[string]string `json:"tags,omitempty"`
	ProvisioningState string            `json:"provisioningState"`
	CreationTime      string            `json:"creationTime"`
	PrimaryEndpoints  StorageEndpoints  `json:"primaryEndpoints"`
	Type              string            `json:"type"`
}

// StorageSKU represents a storage account SKU.
type StorageSKU struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
}

// StorageEndpoints holds simulated endpoint URLs.
type StorageEndpoints struct {
	Blob  string `json:"blob"`
	Queue string `json:"queue"`
	Table string `json:"table"`
	File  string `json:"file"`
}

// Container mirrors az storage container show JSON output.
type Container struct {
	Name         string            `json:"name"`
	AccountName  string            `json:"accountName"`
	LastModified string            `json:"lastModified"`
	PublicAccess string            `json:"publicAccess"`
	LeaseState   string            `json:"leaseState"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Blob mirrors az storage blob show JSON output.
type Blob struct {
	Name         string            `json:"name"`
	Container    string            `json:"container"`
	AccountName  string            `json:"accountName"`
	ContentType  string            `json:"contentType"`
	ContentLen   int64             `json:"contentLength"`
	LastModified string            `json:"lastModified"`
	BlobType     string            `json:"blobType"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	// LocalPath stores the simulator-local file path backing this blob.
	LocalPath string `json:"_localPath,omitempty"`
}

// KeyVault mirrors az keyvault show JSON output.
type KeyVault struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	ResourceGroup     string           `json:"resourceGroup"`
	Location          string           `json:"location"`
	Tags              map[string]string `json:"tags,omitempty"`
	Properties        VaultProperties  `json:"properties"`
	Type              string           `json:"type"`
}

// VaultProperties holds vault-level settings.
type VaultProperties struct {
	TenantID           string `json:"tenantId"`
	SKU                VaultSKU `json:"sku"`
	VaultURI           string `json:"vaultUri"`
	EnableSoftDelete   bool   `json:"enableSoftDelete"`
	ProvisioningState  string `json:"provisioningState"`
}

// VaultSKU is the key vault pricing tier.
type VaultSKU struct {
	Family string `json:"family"`
	Name   string `json:"name"`
}

// VaultSecret mirrors az keyvault secret show JSON output.
type VaultSecret struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Value       string            `json:"value"`
	VaultName   string            `json:"vaultName"`
	Enabled     bool              `json:"enabled"`
	Created     string            `json:"created"`
	Updated     string            `json:"updated"`
	ContentType string            `json:"contentType,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Version     string            `json:"version"`
}

// VaultKey mirrors az keyvault key show JSON output.
type VaultKey struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	VaultName string            `json:"vaultName"`
	Enabled   bool              `json:"enabled"`
	Created   string            `json:"created"`
	Updated   string            `json:"updated"`
	KeyType   string            `json:"kty"`
	KeySize   int               `json:"key_size"`
	KeyOps    []string          `json:"key_ops"`
	Tags      map[string]string `json:"tags,omitempty"`
	Version   string            `json:"version"`
}

// Data is the top-level persisted state.
type Data struct {
	ActiveSubscription string                   `json:"active_subscription"`
	Subscriptions      []Subscription           `json:"subscriptions"`
	ResourceGroups     map[string]*ResourceGroup `json:"resource_groups"`
	Resources          map[string]*Resource      `json:"resources"`
	// Storage
	StorageAccounts map[string]*StorageAccount          `json:"storage_accounts,omitempty"`
	Containers      map[string]map[string]*Container    `json:"containers,omitempty"`      // acct -> name -> container
	Blobs           map[string]map[string]map[string]*Blob `json:"blobs,omitempty"`        // acct -> container -> blob
	// Key Vault
	KeyVaults    map[string]*KeyVault                `json:"key_vaults,omitempty"`
	VaultSecrets map[string]map[string]*VaultSecret  `json:"vault_secrets,omitempty"`   // vault -> name -> secret
	VaultKeys    map[string]map[string]*VaultKey     `json:"vault_keys,omitempty"`      // vault -> name -> key
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
	if err != nil || len(raw) == 0 {
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read state file: %w", err)
		}
		// File missing or empty — seed defaults.
		s.data = seedData()
		return s, s.persist()
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
	if d.StorageAccounts == nil {
		d.StorageAccounts = make(map[string]*StorageAccount)
	}
	if d.Containers == nil {
		d.Containers = make(map[string]map[string]*Container)
	}
	if d.Blobs == nil {
		d.Blobs = make(map[string]map[string]map[string]*Blob)
	}
	if d.KeyVaults == nil {
		d.KeyVaults = make(map[string]*KeyVault)
	}
	if d.VaultSecrets == nil {
		d.VaultSecrets = make(map[string]map[string]*VaultSecret)
	}
	if d.VaultKeys == nil {
		d.VaultKeys = make(map[string]map[string]*VaultKey)
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
		ResourceGroups:  make(map[string]*ResourceGroup),
		Resources:       make(map[string]*Resource),
		StorageAccounts: make(map[string]*StorageAccount),
		Containers:      make(map[string]map[string]*Container),
		Blobs:           make(map[string]map[string]map[string]*Blob),
		KeyVaults:       make(map[string]*KeyVault),
		VaultSecrets:    make(map[string]map[string]*VaultSecret),
		VaultKeys:       make(map[string]map[string]*VaultKey),
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

// ---------------------------------------------------------------------------
// Storage Account CRUD
// ---------------------------------------------------------------------------

// CreateStorageAccount creates a new simulated storage account.
func (s *Store) CreateStorageAccount(name, rg, location, kind, skuName string, tags map[string]string) (*StorageAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.StorageAccounts[name]; exists {
		return nil, fmt.Errorf("storage account '%s' already exists", name)
	}
	sub := s.activeSub()
	tier := "Standard"
	if len(skuName) > 0 && skuName[0] == 'P' {
		tier = "Premium"
	}
	sa := &StorageAccount{
		ID:            GenerateResourceID(sub.ID, rg, "Microsoft.Storage", "storageAccounts", name),
		Name:          name,
		ResourceGroup: rg,
		Location:      location,
		Kind:          kind,
		SKU:           StorageSKU{Name: skuName, Tier: tier},
		Tags:          tags,
		ProvisioningState: "Succeeded",
		CreationTime:  Timestamp(),
		PrimaryEndpoints: StorageEndpoints{
			Blob:  fmt.Sprintf("https://%s.blob.core.windows.net/", name),
			Queue: fmt.Sprintf("https://%s.queue.core.windows.net/", name),
			Table: fmt.Sprintf("https://%s.table.core.windows.net/", name),
			File:  fmt.Sprintf("https://%s.file.core.windows.net/", name),
		},
		Type: "Microsoft.Storage/storageAccounts",
	}
	s.data.StorageAccounts[name] = sa
	_ = s.persist()
	copy := *sa
	return &copy, nil
}

// GetStorageAccount returns nil if not found.
func (s *Store) GetStorageAccount(name string) *StorageAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sa := s.data.StorageAccounts[name]
	if sa == nil {
		return nil
	}
	copy := *sa
	return &copy
}

// ListStorageAccounts returns all storage accounts, optionally filtered by RG.
func (s *Store) ListStorageAccounts(rg string) []*StorageAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*StorageAccount
	for _, sa := range s.data.StorageAccounts {
		if rg == "" || sa.ResourceGroup == rg {
			copy := *sa
			out = append(out, &copy)
		}
	}
	return out
}

// DeleteStorageAccount removes a storage account and all its containers/blobs.
func (s *Store) DeleteStorageAccount(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.StorageAccounts[name]; !ok {
		return fmt.Errorf("storage account '%s' not found", name)
	}
	delete(s.data.StorageAccounts, name)
	delete(s.data.Containers, name)
	delete(s.data.Blobs, name)
	return s.persist()
}

// ---------------------------------------------------------------------------
// Container CRUD
// ---------------------------------------------------------------------------

// CreateContainer creates a blob container inside a storage account.
func (s *Store) CreateContainer(account, name string) (*Container, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.StorageAccounts[account]; !ok {
		return nil, fmt.Errorf("storage account '%s' not found", account)
	}
	if s.data.Containers[account] == nil {
		s.data.Containers[account] = make(map[string]*Container)
	}
	if _, exists := s.data.Containers[account][name]; exists {
		return nil, fmt.Errorf("container '%s' already exists in account '%s'", name, account)
	}
	c := &Container{
		Name:         name,
		AccountName:  account,
		LastModified: Timestamp(),
		PublicAccess: "off",
		LeaseState:   "available",
	}
	s.data.Containers[account][name] = c
	_ = s.persist()
	copy := *c
	return &copy, nil
}

// GetContainer returns nil if not found.
func (s *Store) GetContainer(account, name string) *Container {
	s.mu.RLock()
	defer s.mu.RUnlock()
	acctMap := s.data.Containers[account]
	if acctMap == nil {
		return nil
	}
	c := acctMap[name]
	if c == nil {
		return nil
	}
	copy := *c
	return &copy
}

// ListContainers lists all containers in a storage account.
func (s *Store) ListContainers(account string) []*Container {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Container
	for _, c := range s.data.Containers[account] {
		copy := *c
		out = append(out, &copy)
	}
	return out
}

// DeleteContainer removes a container and all its blobs.
func (s *Store) DeleteContainer(account, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	acctMap := s.data.Containers[account]
	if acctMap == nil || acctMap[name] == nil {
		return fmt.Errorf("container '%s' not found in account '%s'", name, account)
	}
	delete(acctMap, name)
	if blobMap, ok := s.data.Blobs[account]; ok {
		delete(blobMap, name)
	}
	return s.persist()
}

// ---------------------------------------------------------------------------
// Blob CRUD
// ---------------------------------------------------------------------------

// CreateBlob creates (or overwrites) a blob in a container.
func (s *Store) CreateBlob(account, container, name, contentType string, size int64, localPath string) (*Blob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acctContainers := s.data.Containers[account]
	if acctContainers == nil || acctContainers[container] == nil {
		return nil, fmt.Errorf("container '%s' not found in account '%s'", container, account)
	}
	if s.data.Blobs[account] == nil {
		s.data.Blobs[account] = make(map[string]map[string]*Blob)
	}
	if s.data.Blobs[account][container] == nil {
		s.data.Blobs[account][container] = make(map[string]*Blob)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	b := &Blob{
		Name:         name,
		Container:    container,
		AccountName:  account,
		ContentType:  contentType,
		ContentLen:   size,
		LastModified: Timestamp(),
		BlobType:     "BlockBlob",
		LocalPath:    localPath,
	}
	s.data.Blobs[account][container][name] = b
	_ = s.persist()
	copy := *b
	return &copy, nil
}

// GetBlob returns nil if not found.
func (s *Store) GetBlob(account, container, name string) *Blob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	acct := s.data.Blobs[account]
	if acct == nil {
		return nil
	}
	cont := acct[container]
	if cont == nil {
		return nil
	}
	b := cont[name]
	if b == nil {
		return nil
	}
	copy := *b
	return &copy
}

// ListBlobs lists all blobs in a container.
func (s *Store) ListBlobs(account, container string) []*Blob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Blob
	acct := s.data.Blobs[account]
	if acct == nil {
		return out
	}
	for _, b := range acct[container] {
		copy := *b
		out = append(out, &copy)
	}
	return out
}

// DeleteBlob removes a blob.
func (s *Store) DeleteBlob(account, container, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	acct := s.data.Blobs[account]
	if acct == nil {
		return fmt.Errorf("blob '%s' not found", name)
	}
	cont := acct[container]
	if cont == nil || cont[name] == nil {
		return fmt.Errorf("blob '%s' not found in container '%s'", name, container)
	}
	delete(cont, name)
	return s.persist()
}

// ---------------------------------------------------------------------------
// Key Vault CRUD
// ---------------------------------------------------------------------------

// CreateKeyVault creates a new simulated key vault.
func (s *Store) CreateKeyVault(name, rg, location, skuName string, tags map[string]string) (*KeyVault, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.KeyVaults[name]; exists {
		return nil, fmt.Errorf("vault '%s' already exists", name)
	}
	sub := s.activeSub()
	kv := &KeyVault{
		ID:            GenerateResourceID(sub.ID, rg, "Microsoft.KeyVault", "vaults", name),
		Name:          name,
		ResourceGroup: rg,
		Location:      location,
		Tags:          tags,
		Properties: VaultProperties{
			TenantID:          sub.TenantID,
			SKU:               VaultSKU{Family: "A", Name: skuName},
			VaultURI:          fmt.Sprintf("https://%s.vault.azure.net/", name),
			EnableSoftDelete:  true,
			ProvisioningState: "Succeeded",
		},
		Type: "Microsoft.KeyVault/vaults",
	}
	s.data.KeyVaults[name] = kv
	_ = s.persist()
	copy := *kv
	return &copy, nil
}

// GetKeyVault returns nil if not found.
func (s *Store) GetKeyVault(name string) *KeyVault {
	s.mu.RLock()
	defer s.mu.RUnlock()
	kv := s.data.KeyVaults[name]
	if kv == nil {
		return nil
	}
	copy := *kv
	return &copy
}

// ListKeyVaults returns all vaults, optionally filtered by RG.
func (s *Store) ListKeyVaults(rg string) []*KeyVault {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*KeyVault
	for _, kv := range s.data.KeyVaults {
		if rg == "" || kv.ResourceGroup == rg {
			copy := *kv
			out = append(out, &copy)
		}
	}
	return out
}

// DeleteKeyVault removes a vault and all its secrets/keys.
func (s *Store) DeleteKeyVault(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.KeyVaults[name]; !ok {
		return fmt.Errorf("vault '%s' not found", name)
	}
	delete(s.data.KeyVaults, name)
	delete(s.data.VaultSecrets, name)
	delete(s.data.VaultKeys, name)
	return s.persist()
}

// ---------------------------------------------------------------------------
// Vault Secret CRUD
// ---------------------------------------------------------------------------

// SetSecret creates or updates a secret in a vault (returns the new version).
func (s *Store) SetSecret(vaultName, name, value, contentType string, tags map[string]string) (*VaultSecret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.KeyVaults[vaultName]; !ok {
		return nil, fmt.Errorf("vault '%s' not found", vaultName)
	}
	if s.data.VaultSecrets[vaultName] == nil {
		s.data.VaultSecrets[vaultName] = make(map[string]*VaultSecret)
	}
	now := Timestamp()
	version := fmt.Sprintf("%d", time.Now().UnixNano())
	sec := &VaultSecret{
		ID:          fmt.Sprintf("https://%s.vault.azure.net/secrets/%s/%s", vaultName, name, version),
		Name:        name,
		Value:       value,
		VaultName:   vaultName,
		Enabled:     true,
		Created:     now,
		Updated:     now,
		ContentType: contentType,
		Tags:        tags,
		Version:     version,
	}
	s.data.VaultSecrets[vaultName][name] = sec
	_ = s.persist()
	copy := *sec
	return &copy, nil
}

// GetSecret returns nil if not found.
func (s *Store) GetSecret(vaultName, name string) *VaultSecret {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vm := s.data.VaultSecrets[vaultName]
	if vm == nil {
		return nil
	}
	sec := vm[name]
	if sec == nil {
		return nil
	}
	copy := *sec
	return &copy
}

// ListSecrets lists all secrets in a vault.
func (s *Store) ListSecrets(vaultName string) []*VaultSecret {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*VaultSecret
	for _, sec := range s.data.VaultSecrets[vaultName] {
		copy := *sec
		out = append(out, &copy)
	}
	return out
}

// DeleteSecret removes a secret from a vault.
func (s *Store) DeleteSecret(vaultName, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vm := s.data.VaultSecrets[vaultName]
	if vm == nil || vm[name] == nil {
		return fmt.Errorf("secret '%s' not found in vault '%s'", name, vaultName)
	}
	delete(vm, name)
	return s.persist()
}

// ---------------------------------------------------------------------------
// Vault Key CRUD
// ---------------------------------------------------------------------------

// CreateKey creates a new key in a vault.
func (s *Store) CreateKey(vaultName, name, kty string, keySize int, tags map[string]string) (*VaultKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.KeyVaults[vaultName]; !ok {
		return nil, fmt.Errorf("vault '%s' not found", vaultName)
	}
	if s.data.VaultKeys[vaultName] == nil {
		s.data.VaultKeys[vaultName] = make(map[string]*VaultKey)
	}
	now := Timestamp()
	version := fmt.Sprintf("%d", time.Now().UnixNano())
	ops := []string{"encrypt", "decrypt", "sign", "verify", "wrapKey", "unwrapKey"}
	key := &VaultKey{
		ID:        fmt.Sprintf("https://%s.vault.azure.net/keys/%s/%s", vaultName, name, version),
		Name:      name,
		VaultName: vaultName,
		Enabled:   true,
		Created:   now,
		Updated:   now,
		KeyType:   kty,
		KeySize:   keySize,
		KeyOps:    ops,
		Tags:      tags,
		Version:   version,
	}
	s.data.VaultKeys[vaultName][name] = key
	_ = s.persist()
	copy := *key
	return &copy, nil
}

// GetKey returns nil if not found.
func (s *Store) GetKey(vaultName, name string) *VaultKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vm := s.data.VaultKeys[vaultName]
	if vm == nil {
		return nil
	}
	key := vm[name]
	if key == nil {
		return nil
	}
	copy := *key
	return &copy
}

// ListKeys lists all keys in a vault.
func (s *Store) ListKeys(vaultName string) []*VaultKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*VaultKey
	for _, k := range s.data.VaultKeys[vaultName] {
		copy := *k
		out = append(out, &copy)
	}
	return out
}

// DeleteKey removes a key from a vault.
func (s *Store) DeleteKey(vaultName, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vm := s.data.VaultKeys[vaultName]
	if vm == nil || vm[name] == nil {
		return fmt.Errorf("key '%s' not found in vault '%s'", name, vaultName)
	}
	delete(vm, name)
	return s.persist()
}

// activeSub returns the active subscription (caller must hold at least RLock).
func (s *Store) activeSub() *Subscription {
	for i := range s.data.Subscriptions {
		if s.data.Subscriptions[i].IsDefault {
			return &s.data.Subscriptions[i]
		}
	}
	if len(s.data.Subscriptions) > 0 {
		return &s.data.Subscriptions[0]
	}
	return nil
}
