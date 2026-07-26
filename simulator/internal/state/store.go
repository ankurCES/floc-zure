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

// --- Networking types ---

// VNet mirrors az network vnet show JSON output.
type VNet struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	AddressSpace      AddressSpace      `json:"addressSpace"`
	Subnets           []Subnet          `json:"subnets"`
	Tags              map[string]string `json:"tags,omitempty"`
	ProvisioningState string            `json:"provisioningState"`
	Type              string            `json:"type"`
}

// AddressSpace holds CIDR prefixes for a VNet.
type AddressSpace struct {
	AddressPrefixes []string `json:"addressPrefixes"`
}

// Subnet mirrors az network vnet subnet show JSON output.
type Subnet struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	AddressPrefix     string `json:"addressPrefix"`
	ProvisioningState string `json:"provisioningState"`
	NSG               *NSGRef `json:"networkSecurityGroup,omitempty"`
}

// NSGRef is a reference to an NSG by ID.
type NSGRef struct {
	ID string `json:"id"`
}

// NSG mirrors az network nsg show JSON output.
type NSG struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	SecurityRules     []NSGRule         `json:"securityRules"`
	Tags              map[string]string `json:"tags,omitempty"`
	ProvisioningState string            `json:"provisioningState"`
	Type              string            `json:"type"`
}

// NSGRule mirrors az network nsg rule show JSON output.
type NSGRule struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Priority               int    `json:"priority"`
	Direction              string `json:"direction"`
	Access                 string `json:"access"`
	Protocol               string `json:"protocol"`
	SourceAddressPrefix    string `json:"sourceAddressPrefix"`
	SourcePortRange        string `json:"sourcePortRange"`
	DestAddressPrefix      string `json:"destinationAddressPrefix"`
	DestPortRange          string `json:"destinationPortRange"`
	ProvisioningState      string `json:"provisioningState"`
}

// PublicIP mirrors az network public-ip show JSON output.
type PublicIP struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	IPAddress         string            `json:"ipAddress"`
	PublicIPVersion   string            `json:"publicIPAllocationMethod"`
	AllocationMethod  string            `json:"publicIpAllocationMethod"`
	SKU               PublicIPSKU       `json:"sku"`
	Tags              map[string]string `json:"tags,omitempty"`
	ProvisioningState string            `json:"provisioningState"`
	Type              string            `json:"type"`
}

// PublicIPSKU is the public IP pricing tier.
type PublicIPSKU struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
}

// --- VM types ---

// VMState represents the power state of a virtual machine.
type VMState string

const (
	VMStateCreating     VMState = "Creating"
	VMStateRunning      VMState = "Running"
	VMStateStopped      VMState = "Stopped"
	VMStateDeallocated  VMState = "Deallocated"
	VMStateDeleting     VMState = "Deleting"
)

// VirtualMachine mirrors az vm show JSON output.
type VirtualMachine struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	VMSize            string            `json:"hardwareProfile.vmSize"`
	HardwareProfile   HardwareProfile   `json:"hardwareProfile"`
	StorageProfile    VMStorageProfile  `json:"storageProfile"`
	OSProfile         OSProfile         `json:"osProfile"`
	NetworkProfile    NetworkProfile    `json:"networkProfile"`
	Tags              map[string]string `json:"tags,omitempty"`
	ProvisioningState string            `json:"provisioningState"`
	PowerState        VMState           `json:"powerState"`
	Type              string            `json:"type"`
}

// HardwareProfile holds the VM size.
type HardwareProfile struct {
	VMSize string `json:"vmSize"`
}

// VMStorageProfile holds OS disk and image reference.
type VMStorageProfile struct {
	ImageReference ImageReference `json:"imageReference"`
	OSDisk         OSDisk         `json:"osDisk"`
}

// ImageReference identifies the VM image.
type ImageReference struct {
	Publisher string `json:"publisher"`
	Offer     string `json:"offer"`
	SKU       string `json:"sku"`
	Version   string `json:"version"`
}

// OSDisk holds OS disk properties.
type OSDisk struct {
	Name         string `json:"name"`
	CreateOption string `json:"createOption"`
	DiskSizeGB   int    `json:"diskSizeGB"`
	OSType       string `json:"osType"`
}

// OSProfile holds computer name and admin user.
type OSProfile struct {
	ComputerName  string `json:"computerName"`
	AdminUsername string `json:"adminUsername"`
}

// NetworkProfile holds NIC references.
type NetworkProfile struct {
	NetworkInterfaces []NICRef `json:"networkInterfaces"`
}

// NICRef is a reference to a NIC by ID.
type NICRef struct {
	ID string `json:"id"`
}

// --- Service Bus types ---

// ServiceBusNamespace mirrors az servicebus namespace show JSON output.
type ServiceBusNamespace struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	SKU               ServiceBusSKU     `json:"sku"`
	Tags              map[string]string `json:"tags,omitempty"`
	ProvisioningState string            `json:"provisioningState"`
	CreatedAt         string            `json:"createdAt"`
	ServiceBusEndpoint string           `json:"serviceBusEndpoint"`
	Type              string            `json:"type"`
}

// ServiceBusSKU is the pricing tier for a namespace.
type ServiceBusSKU struct {
	Name     string `json:"name"`
	Tier     string `json:"tier"`
	Capacity int    `json:"capacity"`
}

// ServiceBusQueue mirrors az servicebus queue show JSON output.
type ServiceBusQueue struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	NamespaceName            string `json:"namespaceName"`
	MaxSizeInMegabytes       int    `json:"maxSizeInMegabytes"`
	MessageCount             int64  `json:"messageCount"`
	DeadLetterMessageCount   int64  `json:"countDetails.deadLetterMessageCount"`
	DefaultMessageTimeToLive string `json:"defaultMessageTimeToLive"`
	LockDuration             string `json:"lockDuration"`
	RequiresSession          bool   `json:"requiresSession"`
	Status                   string `json:"status"`
	CreatedAt                string `json:"createdAt"`
	Type                     string `json:"type"`
}

// ServiceBusTopic mirrors az servicebus topic show JSON output.
type ServiceBusTopic struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	NamespaceName            string `json:"namespaceName"`
	MaxSizeInMegabytes       int    `json:"maxSizeInMegabytes"`
	DefaultMessageTimeToLive string `json:"defaultMessageTimeToLive"`
	SubscriptionCount        int    `json:"subscriptionCount"`
	Status                   string `json:"status"`
	CreatedAt                string `json:"createdAt"`
	Type                     string `json:"type"`
}

// ServiceBusSubscription mirrors az servicebus topic subscription show JSON output.
type ServiceBusSubscription struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	TopicName                string `json:"topicName"`
	NamespaceName            string `json:"namespaceName"`
	MaxDeliveryCount         int    `json:"maxDeliveryCount"`
	DefaultMessageTimeToLive string `json:"defaultMessageTimeToLive"`
	LockDuration             string `json:"lockDuration"`
	MessageCount             int64  `json:"messageCount"`
	RequiresSession          bool   `json:"requiresSession"`
	Status                   string `json:"status"`
	CreatedAt                string `json:"createdAt"`
	Type                     string `json:"type"`
}

// ServiceBusMessage represents a message in a queue or topic subscription.
type ServiceBusMessage struct {
	MessageID   string                 `json:"messageId"`
	Body        string                 `json:"body"`
	ContentType string                 `json:"contentType,omitempty"`
	Label       string                 `json:"label,omitempty"`
	Properties  map[string]interface{} `json:"userProperties,omitempty"`
	EnqueuedAt  string                 `json:"enqueuedTimeUtc"`
}

// --- Function App types ---

// FunctionApp mirrors az functionapp show JSON output.
type FunctionApp struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	Kind              string            `json:"kind"`
	State             string            `json:"state"`
	DefaultHostName   string            `json:"defaultHostName"`
	Runtime           string            `json:"runtime"`
	RuntimeVersion    string            `json:"runtimeVersion"`
	Tags              map[string]string `json:"tags,omitempty"`
	ProvisioningState string            `json:"provisioningState"`
	Type              string            `json:"type"`
}

// Function represents a single function within a function app.
type Function struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	FunctionApp  string          `json:"functionApp"`
	TriggerType  string          `json:"triggerType"`
	IsDisabled   bool            `json:"isDisabled"`
	Language     string          `json:"language"`
	ScriptFile   string          `json:"scriptFile,omitempty"`
	Bindings     []FuncBinding   `json:"bindings,omitempty"`
	InvokeURL    string          `json:"invokeUrlTemplate"`
}

// FuncBinding represents an input/output binding on a function.
type FuncBinding struct {
	Type      string `json:"type"`
	Direction string `json:"direction"`
	Name      string `json:"name"`
}

// FunctionInvocation records a simulated function invocation.
type FunctionInvocation struct {
	ID            string `json:"id"`
	FunctionName  string `json:"functionName"`
	FunctionApp   string `json:"functionApp"`
	Status        string `json:"status"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	DurationMs    int64  `json:"duration"`
	Input         string `json:"input,omitempty"`
	Output        string `json:"output,omitempty"`
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
	// Networking
	VNets     map[string]*VNet     `json:"vnets,omitempty"`
	NSGs      map[string]*NSG      `json:"nsgs,omitempty"`
	PublicIPs map[string]*PublicIP `json:"public_ips,omitempty"`
	// VMs
	VMs map[string]*VirtualMachine `json:"vms,omitempty"`
	// Service Bus
	ServiceBusNamespaces map[string]*ServiceBusNamespace                              `json:"servicebus_namespaces,omitempty"`
	ServiceBusQueues     map[string]map[string]*ServiceBusQueue                       `json:"servicebus_queues,omitempty"`     // ns -> name -> queue
	ServiceBusTopics     map[string]map[string]*ServiceBusTopic                       `json:"servicebus_topics,omitempty"`     // ns -> name -> topic
	ServiceBusSubs       map[string]map[string]map[string]*ServiceBusSubscription     `json:"servicebus_subs,omitempty"`       // ns -> topic -> name -> sub
	ServiceBusMessages   map[string]map[string][]ServiceBusMessage                    `json:"servicebus_messages,omitempty"`   // ns -> queue -> messages
	// Function Apps
	FunctionApps       map[string]*FunctionApp                    `json:"function_apps,omitempty"`
	Functions          map[string]map[string]*Function            `json:"functions,omitempty"`           // app -> name -> function
	FunctionInvocations map[string][]FunctionInvocation           `json:"function_invocations,omitempty"` // app -> invocations
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
	if d.VNets == nil {
		d.VNets = make(map[string]*VNet)
	}
	if d.NSGs == nil {
		d.NSGs = make(map[string]*NSG)
	}
	if d.PublicIPs == nil {
		d.PublicIPs = make(map[string]*PublicIP)
	}
	if d.VMs == nil {
		d.VMs = make(map[string]*VirtualMachine)
	}
	if d.ServiceBusNamespaces == nil {
		d.ServiceBusNamespaces = make(map[string]*ServiceBusNamespace)
	}
	if d.ServiceBusQueues == nil {
		d.ServiceBusQueues = make(map[string]map[string]*ServiceBusQueue)
	}
	if d.ServiceBusTopics == nil {
		d.ServiceBusTopics = make(map[string]map[string]*ServiceBusTopic)
	}
	if d.ServiceBusSubs == nil {
		d.ServiceBusSubs = make(map[string]map[string]map[string]*ServiceBusSubscription)
	}
	if d.ServiceBusMessages == nil {
		d.ServiceBusMessages = make(map[string]map[string][]ServiceBusMessage)
	}
	if d.FunctionApps == nil {
		d.FunctionApps = make(map[string]*FunctionApp)
	}
	if d.Functions == nil {
		d.Functions = make(map[string]map[string]*Function)
	}
	if d.FunctionInvocations == nil {
		d.FunctionInvocations = make(map[string][]FunctionInvocation)
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
		VNets:           make(map[string]*VNet),
		NSGs:            make(map[string]*NSG),
		PublicIPs:       make(map[string]*PublicIP),
		VMs:                  make(map[string]*VirtualMachine),
		ServiceBusNamespaces: make(map[string]*ServiceBusNamespace),
		ServiceBusQueues:     make(map[string]map[string]*ServiceBusQueue),
		ServiceBusTopics:     make(map[string]map[string]*ServiceBusTopic),
		ServiceBusSubs:       make(map[string]map[string]map[string]*ServiceBusSubscription),
		ServiceBusMessages:   make(map[string]map[string][]ServiceBusMessage),
		FunctionApps:         make(map[string]*FunctionApp),
		Functions:            make(map[string]map[string]*Function),
		FunctionInvocations:  make(map[string][]FunctionInvocation),
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

// ---------------------------------------------------------------------------
// VNet CRUD
// ---------------------------------------------------------------------------

// CreateVNet creates a virtual network. addressPrefixes defaults to ["10.0.0.0/16"].
func (s *Store) CreateVNet(name, rg, location string, addressPrefixes []string, tags map[string]string) (*VNet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.VNets[name]; exists {
		return nil, fmt.Errorf("vnet '%s' already exists", name)
	}
	sub := s.activeSub()
	if sub == nil {
		return nil, fmt.Errorf("no active subscription")
	}
	if len(addressPrefixes) == 0 {
		addressPrefixes = []string{"10.0.0.0/16"}
	}
	vnet := &VNet{
		ID:            fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", sub.ID, rg, name),
		Name:          name,
		ResourceGroup: rg,
		Location:      location,
		AddressSpace:  AddressSpace{AddressPrefixes: addressPrefixes},
		Subnets:       []Subnet{},
		Tags:          tags,
		ProvisioningState: "Succeeded",
		Type:          "Microsoft.Network/virtualNetworks",
	}
	s.data.VNets[name] = vnet
	return vnet, s.persist()
}

// GetVNet returns a VNet by name.
func (s *Store) GetVNet(name string) *VNet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.VNets[name]
}

// ListVNets lists VNets, optionally filtered by resource group.
func (s *Store) ListVNets(rg string) []VNet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []VNet
	for _, v := range s.data.VNets {
		if rg == "" || v.ResourceGroup == rg {
			out = append(out, *v)
		}
	}
	return out
}

// DeleteVNet removes a VNet.
func (s *Store) DeleteVNet(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.VNets[name]; !ok {
		return fmt.Errorf("vnet '%s' not found", name)
	}
	delete(s.data.VNets, name)
	return s.persist()
}

// CreateSubnet adds a subnet to a VNet.
func (s *Store) CreateSubnet(vnetName, subnetName, addressPrefix string) (*Subnet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vnet, ok := s.data.VNets[vnetName]
	if !ok {
		return nil, fmt.Errorf("vnet '%s' not found", vnetName)
	}
	for _, sub := range vnet.Subnets {
		if sub.Name == subnetName {
			return nil, fmt.Errorf("subnet '%s' already exists in vnet '%s'", subnetName, vnetName)
		}
	}
	subnet := Subnet{
		ID:                vnet.ID + "/subnets/" + subnetName,
		Name:              subnetName,
		AddressPrefix:     addressPrefix,
		ProvisioningState: "Succeeded",
	}
	vnet.Subnets = append(vnet.Subnets, subnet)
	return &subnet, s.persist()
}

// GetSubnet returns a subnet from a VNet.
func (s *Store) GetSubnet(vnetName, subnetName string) *Subnet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vnet, ok := s.data.VNets[vnetName]
	if !ok {
		return nil
	}
	for i := range vnet.Subnets {
		if vnet.Subnets[i].Name == subnetName {
			return &vnet.Subnets[i]
		}
	}
	return nil
}

// ListSubnets returns all subnets of a VNet.
func (s *Store) ListSubnets(vnetName string) []Subnet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vnet, ok := s.data.VNets[vnetName]
	if !ok {
		return nil
	}
	return vnet.Subnets
}

// DeleteSubnet removes a subnet from a VNet.
func (s *Store) DeleteSubnet(vnetName, subnetName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vnet, ok := s.data.VNets[vnetName]
	if !ok {
		return fmt.Errorf("vnet '%s' not found", vnetName)
	}
	for i, sub := range vnet.Subnets {
		if sub.Name == subnetName {
			vnet.Subnets = append(vnet.Subnets[:i], vnet.Subnets[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("subnet '%s' not found in vnet '%s'", subnetName, vnetName)
}

// ---------------------------------------------------------------------------
// NSG CRUD
// ---------------------------------------------------------------------------

// CreateNSG creates a network security group.
func (s *Store) CreateNSG(name, rg, location string, tags map[string]string) (*NSG, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.NSGs[name]; exists {
		return nil, fmt.Errorf("nsg '%s' already exists", name)
	}
	sub := s.activeSub()
	if sub == nil {
		return nil, fmt.Errorf("no active subscription")
	}
	nsg := &NSG{
		ID:                fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/networkSecurityGroups/%s", sub.ID, rg, name),
		Name:              name,
		ResourceGroup:     rg,
		Location:          location,
		SecurityRules:     []NSGRule{},
		Tags:              tags,
		ProvisioningState: "Succeeded",
		Type:              "Microsoft.Network/networkSecurityGroups",
	}
	s.data.NSGs[name] = nsg
	return nsg, s.persist()
}

// GetNSG returns an NSG by name.
func (s *Store) GetNSG(name string) *NSG {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.NSGs[name]
}

// ListNSGs lists NSGs, optionally filtered by resource group.
func (s *Store) ListNSGs(rg string) []NSG {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []NSG
	for _, n := range s.data.NSGs {
		if rg == "" || n.ResourceGroup == rg {
			out = append(out, *n)
		}
	}
	return out
}

// DeleteNSG removes an NSG.
func (s *Store) DeleteNSG(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.NSGs[name]; !ok {
		return fmt.Errorf("nsg '%s' not found", name)
	}
	delete(s.data.NSGs, name)
	return s.persist()
}

// CreateNSGRule adds a security rule to an NSG.
func (s *Store) CreateNSGRule(nsgName string, rule NSGRule) (*NSGRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nsg, ok := s.data.NSGs[nsgName]
	if !ok {
		return nil, fmt.Errorf("nsg '%s' not found", nsgName)
	}
	for _, r := range nsg.SecurityRules {
		if r.Name == rule.Name {
			return nil, fmt.Errorf("rule '%s' already exists in nsg '%s'", rule.Name, nsgName)
		}
	}
	rule.ID = nsg.ID + "/securityRules/" + rule.Name
	rule.ProvisioningState = "Succeeded"
	nsg.SecurityRules = append(nsg.SecurityRules, rule)
	return &rule, s.persist()
}

// DeleteNSGRule removes a security rule from an NSG.
func (s *Store) DeleteNSGRule(nsgName, ruleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	nsg, ok := s.data.NSGs[nsgName]
	if !ok {
		return fmt.Errorf("nsg '%s' not found", nsgName)
	}
	for i, r := range nsg.SecurityRules {
		if r.Name == ruleName {
			nsg.SecurityRules = append(nsg.SecurityRules[:i], nsg.SecurityRules[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("rule '%s' not found in nsg '%s'", ruleName, nsgName)
}

// ---------------------------------------------------------------------------
// Public IP CRUD
// ---------------------------------------------------------------------------

// fakeIPCounter is used to generate sequential fake IPs.
var fakeIPCounter int

// CreatePublicIP creates a public IP address.
func (s *Store) CreatePublicIP(name, rg, location, sku, allocation string, tags map[string]string) (*PublicIP, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.PublicIPs[name]; exists {
		return nil, fmt.Errorf("public-ip '%s' already exists", name)
	}
	sub := s.activeSub()
	if sub == nil {
		return nil, fmt.Errorf("no active subscription")
	}
	if sku == "" {
		sku = "Standard"
	}
	if allocation == "" {
		allocation = "Static"
	}
	fakeIPCounter++
	pip := &PublicIP{
		ID:                fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/publicIPAddresses/%s", sub.ID, rg, name),
		Name:              name,
		ResourceGroup:     rg,
		Location:          location,
		IPAddress:         fmt.Sprintf("20.0.%d.%d", fakeIPCounter/256, fakeIPCounter%256),
		PublicIPVersion:   allocation,
		AllocationMethod:  allocation,
		SKU:               PublicIPSKU{Name: sku, Tier: "Regional"},
		Tags:              tags,
		ProvisioningState: "Succeeded",
		Type:              "Microsoft.Network/publicIPAddresses",
	}
	s.data.PublicIPs[name] = pip
	return pip, s.persist()
}

// GetPublicIP returns a PublicIP by name.
func (s *Store) GetPublicIP(name string) *PublicIP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.PublicIPs[name]
}

// ListPublicIPs lists public IPs, optionally filtered by resource group.
func (s *Store) ListPublicIPs(rg string) []PublicIP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []PublicIP
	for _, p := range s.data.PublicIPs {
		if rg == "" || p.ResourceGroup == rg {
			out = append(out, *p)
		}
	}
	return out
}

// DeletePublicIP removes a public IP.
func (s *Store) DeletePublicIP(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.PublicIPs[name]; !ok {
		return fmt.Errorf("public-ip '%s' not found", name)
	}
	delete(s.data.PublicIPs, name)
	return s.persist()
}

// ---------------------------------------------------------------------------
// VM CRUD + State Machine
// ---------------------------------------------------------------------------

// validVMTransitions defines allowed power state transitions.
var validVMTransitions = map[VMState][]VMState{
	VMStateCreating:    {VMStateRunning},
	VMStateRunning:     {VMStateStopped, VMStateDeallocated, VMStateDeleting},
	VMStateStopped:     {VMStateRunning, VMStateDeallocated, VMStateDeleting},
	VMStateDeallocated: {VMStateRunning, VMStateDeleting},
}

// CreateVM creates a virtual machine in the Running state.
func (s *Store) CreateVM(name, rg, location, vmSize, image, adminUser string, tags map[string]string) (*VirtualMachine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.VMs[name]; exists {
		return nil, fmt.Errorf("vm '%s' already exists", name)
	}
	sub := s.activeSub()
	if sub == nil {
		return nil, fmt.Errorf("no active subscription")
	}
	if vmSize == "" {
		vmSize = "Standard_B1s"
	}
	if adminUser == "" {
		adminUser = "azureuser"
	}
	// Parse image URN: publisher:offer:sku:version
	publisher, offer, sku, version := "Canonical", "UbuntuServer", "18.04-LTS", "latest"
	parts := splitImage(image)
	if len(parts) == 4 {
		publisher, offer, sku, version = parts[0], parts[1], parts[2], parts[3]
	}
	vm := &VirtualMachine{
		ID:            fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", sub.ID, rg, name),
		Name:          name,
		ResourceGroup: rg,
		Location:      location,
		VMSize:        vmSize,
		HardwareProfile: HardwareProfile{VMSize: vmSize},
		StorageProfile: VMStorageProfile{
			ImageReference: ImageReference{Publisher: publisher, Offer: offer, SKU: sku, Version: version},
			OSDisk:         OSDisk{Name: name + "-osdisk", CreateOption: "FromImage", DiskSizeGB: 30, OSType: "Linux"},
		},
		OSProfile:     OSProfile{ComputerName: name, AdminUsername: adminUser},
		NetworkProfile: NetworkProfile{NetworkInterfaces: []NICRef{}},
		Tags:          tags,
		ProvisioningState: "Succeeded",
		PowerState:    VMStateRunning,
		Type:          "Microsoft.Compute/virtualMachines",
	}
	s.data.VMs[name] = vm
	return vm, s.persist()
}

// GetVM returns a VM by name.
func (s *Store) GetVM(name string) *VirtualMachine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.VMs[name]
}

// ListVMs lists VMs, optionally filtered by resource group.
func (s *Store) ListVMs(rg string) []VirtualMachine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []VirtualMachine
	for _, vm := range s.data.VMs {
		if rg == "" || vm.ResourceGroup == rg {
			out = append(out, *vm)
		}
	}
	return out
}

// DeleteVM removes a VM.
func (s *Store) DeleteVM(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vm, ok := s.data.VMs[name]
	if !ok {
		return fmt.Errorf("vm '%s' not found", name)
	}
	// Must be in a state that allows deletion
	if !isValidTransition(vm.PowerState, VMStateDeleting) {
		return fmt.Errorf("cannot delete vm '%s' in state '%s'", name, vm.PowerState)
	}
	delete(s.data.VMs, name)
	return s.persist()
}

// TransitionVM changes a VM's power state with state machine validation.
func (s *Store) TransitionVM(name string, target VMState) (*VirtualMachine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vm, ok := s.data.VMs[name]
	if !ok {
		return nil, fmt.Errorf("vm '%s' not found", name)
	}
	if !isValidTransition(vm.PowerState, target) {
		return nil, fmt.Errorf("invalid transition: %s -> %s for vm '%s'", vm.PowerState, target, name)
	}
	vm.PowerState = target
	return vm, s.persist()
}

// isValidTransition checks if a state transition is allowed.
func isValidTransition(from, to VMState) bool {
	allowed, ok := validVMTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

// splitImage splits a URN like "publisher:offer:sku:version".
func splitImage(image string) []string {
	if image == "" {
		return nil
	}
	parts := make([]string, 0, 4)
	start := 0
	for i := 0; i < len(image); i++ {
		if image[i] == ':' {
			parts = append(parts, image[start:i])
			start = i + 1
		}
	}
	parts = append(parts, image[start:])
	return parts
}

// --- Service Bus ---

// CreateServiceBusNamespace creates a new Service Bus namespace.
func (s *Store) CreateServiceBusNamespace(name, rg, location, sku string, tags map[string]string) (*ServiceBusNamespace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.ServiceBusNamespaces[name]; exists {
		return nil, fmt.Errorf("namespace '%s' already exists", name)
	}
	sub := s.data.Subscriptions[0]
	ns := &ServiceBusNamespace{
		ID:                 GenerateResourceID(sub.ID, rg, "Microsoft.ServiceBus", "namespaces", name),
		Name:               name,
		ResourceGroup:      rg,
		Location:           location,
		SKU:                ServiceBusSKU{Name: sku, Tier: sku, Capacity: 1},
		Tags:               tags,
		ProvisioningState:  "Succeeded",
		CreatedAt:          Timestamp(),
		ServiceBusEndpoint: fmt.Sprintf("https://%s.servicebus.windows.net:443/", name),
		Type:               "Microsoft.ServiceBus/namespaces",
	}
	s.data.ServiceBusNamespaces[name] = ns
	_ = s.persist()
	return ns, nil
}

// GetServiceBusNamespace returns a namespace by name or nil.
func (s *Store) GetServiceBusNamespace(name string) *ServiceBusNamespace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.ServiceBusNamespaces[name]
}

// ListServiceBusNamespaces returns all namespaces, optionally filtered by RG.
func (s *Store) ListServiceBusNamespaces(rg string) []ServiceBusNamespace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ServiceBusNamespace
	for _, ns := range s.data.ServiceBusNamespaces {
		if rg == "" || ns.ResourceGroup == rg {
			out = append(out, *ns)
		}
	}
	return out
}

// DeleteServiceBusNamespace deletes a namespace and all its queues/topics/subs/messages.
func (s *Store) DeleteServiceBusNamespace(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.ServiceBusNamespaces[name]; !exists {
		return fmt.Errorf("namespace '%s' not found", name)
	}
	delete(s.data.ServiceBusNamespaces, name)
	delete(s.data.ServiceBusQueues, name)
	delete(s.data.ServiceBusTopics, name)
	delete(s.data.ServiceBusSubs, name)
	delete(s.data.ServiceBusMessages, name)
	_ = s.persist()
	return nil
}

// CreateServiceBusQueue creates a queue in a namespace.
func (s *Store) CreateServiceBusQueue(nsName, queueName string, maxSize int) (*ServiceBusQueue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ns, ok := s.data.ServiceBusNamespaces[nsName]
	if !ok {
		return nil, fmt.Errorf("namespace '%s' not found", nsName)
	}
	if s.data.ServiceBusQueues[nsName] == nil {
		s.data.ServiceBusQueues[nsName] = make(map[string]*ServiceBusQueue)
	}
	if _, exists := s.data.ServiceBusQueues[nsName][queueName]; exists {
		return nil, fmt.Errorf("queue '%s' already exists in namespace '%s'", queueName, nsName)
	}
	if maxSize <= 0 {
		maxSize = 1024
	}
	q := &ServiceBusQueue{
		ID:                       ns.ID + "/queues/" + queueName,
		Name:                     queueName,
		NamespaceName:            nsName,
		MaxSizeInMegabytes:       maxSize,
		DefaultMessageTimeToLive: "P14D",
		LockDuration:             "PT30S",
		Status:                   "Active",
		CreatedAt:                Timestamp(),
		Type:                     "Microsoft.ServiceBus/namespaces/queues",
	}
	s.data.ServiceBusQueues[nsName][queueName] = q
	_ = s.persist()
	return q, nil
}

// GetServiceBusQueue returns a queue by namespace and name.
func (s *Store) GetServiceBusQueue(nsName, queueName string) *ServiceBusQueue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m := s.data.ServiceBusQueues[nsName]; m != nil {
		return m[queueName]
	}
	return nil
}

// ListServiceBusQueues lists all queues in a namespace.
func (s *Store) ListServiceBusQueues(nsName string) []ServiceBusQueue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ServiceBusQueue
	for _, q := range s.data.ServiceBusQueues[nsName] {
		out = append(out, *q)
	}
	return out
}

// DeleteServiceBusQueue deletes a queue and its messages.
func (s *Store) DeleteServiceBusQueue(nsName, queueName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.data.ServiceBusQueues[nsName]
	if m == nil || m[queueName] == nil {
		return fmt.Errorf("queue '%s' not found in namespace '%s'", queueName, nsName)
	}
	delete(m, queueName)
	if msgs := s.data.ServiceBusMessages[nsName]; msgs != nil {
		delete(msgs, queueName)
	}
	_ = s.persist()
	return nil
}

// CreateServiceBusTopic creates a topic in a namespace.
func (s *Store) CreateServiceBusTopic(nsName, topicName string, maxSize int) (*ServiceBusTopic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ns, ok := s.data.ServiceBusNamespaces[nsName]
	if !ok {
		return nil, fmt.Errorf("namespace '%s' not found", nsName)
	}
	if s.data.ServiceBusTopics[nsName] == nil {
		s.data.ServiceBusTopics[nsName] = make(map[string]*ServiceBusTopic)
	}
	if _, exists := s.data.ServiceBusTopics[nsName][topicName]; exists {
		return nil, fmt.Errorf("topic '%s' already exists in namespace '%s'", topicName, nsName)
	}
	if maxSize <= 0 {
		maxSize = 1024
	}
	t := &ServiceBusTopic{
		ID:                       ns.ID + "/topics/" + topicName,
		Name:                     topicName,
		NamespaceName:            nsName,
		MaxSizeInMegabytes:       maxSize,
		DefaultMessageTimeToLive: "P14D",
		Status:                   "Active",
		CreatedAt:                Timestamp(),
		Type:                     "Microsoft.ServiceBus/namespaces/topics",
	}
	s.data.ServiceBusTopics[nsName][topicName] = t
	_ = s.persist()
	return t, nil
}

// GetServiceBusTopic returns a topic by namespace and name.
func (s *Store) GetServiceBusTopic(nsName, topicName string) *ServiceBusTopic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m := s.data.ServiceBusTopics[nsName]; m != nil {
		return m[topicName]
	}
	return nil
}

// ListServiceBusTopics lists all topics in a namespace.
func (s *Store) ListServiceBusTopics(nsName string) []ServiceBusTopic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ServiceBusTopic
	for _, t := range s.data.ServiceBusTopics[nsName] {
		out = append(out, *t)
	}
	return out
}

// DeleteServiceBusTopic deletes a topic and its subscriptions.
func (s *Store) DeleteServiceBusTopic(nsName, topicName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.data.ServiceBusTopics[nsName]
	if m == nil || m[topicName] == nil {
		return fmt.Errorf("topic '%s' not found in namespace '%s'", topicName, nsName)
	}
	delete(m, topicName)
	if subs := s.data.ServiceBusSubs[nsName]; subs != nil {
		delete(subs, topicName)
	}
	_ = s.persist()
	return nil
}

// CreateServiceBusSub creates a subscription on a topic.
func (s *Store) CreateServiceBusSub(nsName, topicName, subName string, maxDelivery int) (*ServiceBusSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.ServiceBusTopics[nsName] == nil || s.data.ServiceBusTopics[nsName][topicName] == nil {
		return nil, fmt.Errorf("topic '%s' not found in namespace '%s'", topicName, nsName)
	}
	if s.data.ServiceBusSubs[nsName] == nil {
		s.data.ServiceBusSubs[nsName] = make(map[string]map[string]*ServiceBusSubscription)
	}
	if s.data.ServiceBusSubs[nsName][topicName] == nil {
		s.data.ServiceBusSubs[nsName][topicName] = make(map[string]*ServiceBusSubscription)
	}
	if _, exists := s.data.ServiceBusSubs[nsName][topicName][subName]; exists {
		return nil, fmt.Errorf("subscription '%s' already exists on topic '%s'", subName, topicName)
	}
	if maxDelivery <= 0 {
		maxDelivery = 10
	}
	ns := s.data.ServiceBusNamespaces[nsName]
	sub := &ServiceBusSubscription{
		ID:                       ns.ID + "/topics/" + topicName + "/subscriptions/" + subName,
		Name:                     subName,
		TopicName:                topicName,
		NamespaceName:            nsName,
		MaxDeliveryCount:         maxDelivery,
		DefaultMessageTimeToLive: "P14D",
		LockDuration:             "PT30S",
		Status:                   "Active",
		CreatedAt:                Timestamp(),
		Type:                     "Microsoft.ServiceBus/namespaces/topics/subscriptions",
	}
	s.data.ServiceBusSubs[nsName][topicName][subName] = sub
	// Update topic subscription count.
	s.data.ServiceBusTopics[nsName][topicName].SubscriptionCount++
	_ = s.persist()
	return sub, nil
}

// GetServiceBusSub returns a subscription by namespace, topic, and name.
func (s *Store) GetServiceBusSub(nsName, topicName, subName string) *ServiceBusSubscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m := s.data.ServiceBusSubs[nsName]; m != nil {
		if tm := m[topicName]; tm != nil {
			return tm[subName]
		}
	}
	return nil
}

// ListServiceBusSubs lists subscriptions on a topic.
func (s *Store) ListServiceBusSubs(nsName, topicName string) []ServiceBusSubscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ServiceBusSubscription
	if m := s.data.ServiceBusSubs[nsName]; m != nil {
		for _, sub := range m[topicName] {
			out = append(out, *sub)
		}
	}
	return out
}

// DeleteServiceBusSub deletes a subscription from a topic.
func (s *Store) DeleteServiceBusSub(nsName, topicName, subName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.data.ServiceBusSubs[nsName]
	if m == nil || m[topicName] == nil || m[topicName][subName] == nil {
		return fmt.Errorf("subscription '%s' not found on topic '%s' in namespace '%s'", subName, topicName, nsName)
	}
	delete(m[topicName], subName)
	if t := s.data.ServiceBusTopics[nsName][topicName]; t != nil {
		t.SubscriptionCount--
	}
	_ = s.persist()
	return nil
}

// SendMessage sends a message to a queue.
func (s *Store) SendMessage(nsName, queueName, body, contentType, label string, props map[string]interface{}) (*ServiceBusMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.ServiceBusQueues[nsName] == nil || s.data.ServiceBusQueues[nsName][queueName] == nil {
		return nil, fmt.Errorf("queue '%s' not found in namespace '%s'", queueName, nsName)
	}
	if s.data.ServiceBusMessages[nsName] == nil {
		s.data.ServiceBusMessages[nsName] = make(map[string][]ServiceBusMessage)
	}
	msg := ServiceBusMessage{
		MessageID:   fmt.Sprintf("msg-%d", len(s.data.ServiceBusMessages[nsName][queueName])+1),
		Body:        body,
		ContentType: contentType,
		Label:       label,
		Properties:  props,
		EnqueuedAt:  Timestamp(),
	}
	s.data.ServiceBusMessages[nsName][queueName] = append(s.data.ServiceBusMessages[nsName][queueName], msg)
	s.data.ServiceBusQueues[nsName][queueName].MessageCount++
	_ = s.persist()
	return &msg, nil
}

// ReceiveMessage dequeues the next message from a queue (destructive read).
func (s *Store) ReceiveMessage(nsName, queueName string) (*ServiceBusMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.ServiceBusQueues[nsName] == nil || s.data.ServiceBusQueues[nsName][queueName] == nil {
		return nil, fmt.Errorf("queue '%s' not found in namespace '%s'", queueName, nsName)
	}
	msgs := s.data.ServiceBusMessages[nsName][queueName]
	if len(msgs) == 0 {
		return nil, nil
	}
	msg := msgs[0]
	s.data.ServiceBusMessages[nsName][queueName] = msgs[1:]
	s.data.ServiceBusQueues[nsName][queueName].MessageCount--
	_ = s.persist()
	return &msg, nil
}

// PeekMessage peeks at the next message without removing it.
func (s *Store) PeekMessage(nsName, queueName string) (*ServiceBusMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.data.ServiceBusQueues[nsName] == nil || s.data.ServiceBusQueues[nsName][queueName] == nil {
		return nil, fmt.Errorf("queue '%s' not found in namespace '%s'", queueName, nsName)
	}
	msgs := s.data.ServiceBusMessages[nsName][queueName]
	if len(msgs) == 0 {
		return nil, nil
	}
	m := msgs[0]
	return &m, nil
}

// --- Function Apps ---

// CreateFunctionApp creates a new function app.
func (s *Store) CreateFunctionApp(name, rg, location, runtime, runtimeVersion string, tags map[string]string) (*FunctionApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.FunctionApps[name]; exists {
		return nil, fmt.Errorf("function app '%s' already exists", name)
	}
	sub := s.data.Subscriptions[0]
	if runtime == "" {
		runtime = "node"
	}
	if runtimeVersion == "" {
		runtimeVersion = "18"
	}
	fa := &FunctionApp{
		ID:                GenerateResourceID(sub.ID, rg, "Microsoft.Web", "sites", name),
		Name:              name,
		ResourceGroup:     rg,
		Location:          location,
		Kind:              "functionapp",
		State:             "Running",
		DefaultHostName:   fmt.Sprintf("%s.azurewebsites.net", name),
		Runtime:           runtime,
		RuntimeVersion:    runtimeVersion,
		Tags:              tags,
		ProvisioningState: "Succeeded",
		Type:              "Microsoft.Web/sites",
	}
	s.data.FunctionApps[name] = fa
	_ = s.persist()
	return fa, nil
}

// GetFunctionApp returns a function app by name.
func (s *Store) GetFunctionApp(name string) *FunctionApp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.FunctionApps[name]
}

// ListFunctionApps returns all function apps, optionally filtered by RG.
func (s *Store) ListFunctionApps(rg string) []FunctionApp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []FunctionApp
	for _, fa := range s.data.FunctionApps {
		if rg == "" || fa.ResourceGroup == rg {
			out = append(out, *fa)
		}
	}
	return out
}

// DeleteFunctionApp deletes a function app and its functions/invocations.
func (s *Store) DeleteFunctionApp(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.FunctionApps[name]; !exists {
		return fmt.Errorf("function app '%s' not found", name)
	}
	delete(s.data.FunctionApps, name)
	delete(s.data.Functions, name)
	delete(s.data.FunctionInvocations, name)
	_ = s.persist()
	return nil
}

// CreateFunction creates a function in a function app.
func (s *Store) CreateFunction(appName, funcName, triggerType, language string, bindings []FuncBinding) (*Function, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fa, ok := s.data.FunctionApps[appName]
	if !ok {
		return nil, fmt.Errorf("function app '%s' not found", appName)
	}
	if s.data.Functions[appName] == nil {
		s.data.Functions[appName] = make(map[string]*Function)
	}
	if _, exists := s.data.Functions[appName][funcName]; exists {
		return nil, fmt.Errorf("function '%s' already exists in app '%s'", funcName, appName)
	}
	if triggerType == "" {
		triggerType = "httpTrigger"
	}
	if language == "" {
		language = fa.Runtime
	}
	f := &Function{
		ID:          fa.ID + "/functions/" + funcName,
		Name:        funcName,
		FunctionApp: appName,
		TriggerType: triggerType,
		Language:    language,
		Bindings:    bindings,
		InvokeURL:   fmt.Sprintf("https://%s/api/%s", fa.DefaultHostName, funcName),
	}
	s.data.Functions[appName][funcName] = f
	_ = s.persist()
	return f, nil
}

// GetFunction returns a function by app and name.
func (s *Store) GetFunction(appName, funcName string) *Function {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m := s.data.Functions[appName]; m != nil {
		return m[funcName]
	}
	return nil
}

// ListFunctions lists functions in a function app.
func (s *Store) ListFunctions(appName string) []Function {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Function
	for _, f := range s.data.Functions[appName] {
		out = append(out, *f)
	}
	return out
}

// DeleteFunction deletes a function from an app.
func (s *Store) DeleteFunction(appName, funcName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.data.Functions[appName]
	if m == nil || m[funcName] == nil {
		return fmt.Errorf("function '%s' not found in app '%s'", funcName, appName)
	}
	delete(m, funcName)
	_ = s.persist()
	return nil
}

// InvokeFunction simulates invoking a function and records the result.
func (s *Store) InvokeFunction(appName, funcName, input string) (*FunctionInvocation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Functions[appName] == nil || s.data.Functions[appName][funcName] == nil {
		return nil, fmt.Errorf("function '%s' not found in app '%s'", funcName, appName)
	}
	f := s.data.Functions[appName][funcName]
	if f.IsDisabled {
		return nil, fmt.Errorf("function '%s' is disabled", funcName)
	}
	now := Timestamp()
	inv := FunctionInvocation{
		ID:           fmt.Sprintf("inv-%d", len(s.data.FunctionInvocations[appName])+1),
		FunctionName: funcName,
		FunctionApp:  appName,
		Status:       "Succeeded",
		StartTime:    now,
		EndTime:      now,
		DurationMs:   42, // simulated
		Input:        input,
		Output:       fmt.Sprintf("{\"status\":\"ok\",\"function\":\"%s\",\"echo\":%s}", funcName, input),
	}
	s.data.FunctionInvocations[appName] = append(s.data.FunctionInvocations[appName], inv)
	_ = s.persist()
	return &inv, nil
}

// ListInvocations returns function invocation history for an app.
func (s *Store) ListInvocations(appName string) []FunctionInvocation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.FunctionInvocations[appName]
}
