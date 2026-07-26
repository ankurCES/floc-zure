package arm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ankurCES/floc-zure/internal/azure"
)

// Deployment tracks a template deployment's state and created resources.
type Deployment struct {
	Name              string             `json:"name"`
	ResourceGroup     string             `json:"resourceGroup"`
	TemplatePath      string             `json:"templatePath,omitempty"`
	ProvisioningState string             `json:"provisioningState"`
	Timestamp         string             `json:"timestamp"`
	Duration          string             `json:"duration,omitempty"`
	Mode              string             `json:"mode"` // "Incremental" or "Complete"
	Resources         []DeployedResource `json:"resources,omitempty"`
	Outputs           map[string]interface{} `json:"outputs,omitempty"`
	Error             string             `json:"error,omitempty"`
}

// DeployedResource records a resource created by a deployment.
type DeployedResource struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// Deployer creates simulated resources from a parsed ARM template.
type Deployer struct {
	exec azure.CLIExecutor
}

// NewDeployer creates a deployer backed by the given CLI executor.
func NewDeployer(exec azure.CLIExecutor) *Deployer {
	return &Deployer{exec: exec}
}

// Deploy processes a ParseResult and creates resources via the CLI executor.
// It maps ARM resource types to `az` CLI commands.
func (d *Deployer) Deploy(ctx context.Context, result *ParseResult, deployName, resourceGroup string) (*Deployment, error) {
	start := time.Now()
	dep := &Deployment{
		Name:              deployName,
		ResourceGroup:     resourceGroup,
		ProvisioningState: "Running",
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Mode:              "Incremental",
	}

	// Ensure the resource group exists
	_, err := d.exec.Run(ctx, "group", "create", "--name", resourceGroup, "--location", "eastus")
	if err != nil {
		dep.ProvisioningState = "Failed"
		dep.Error = fmt.Sprintf("create resource group: %v", err)
		return dep, fmt.Errorf("create resource group %q: %w", resourceGroup, err)
	}

	// Deploy each resource in order
	for _, res := range result.Resources {
		deployed, err := d.deployResource(ctx, res, resourceGroup)
		if err != nil {
			dep.ProvisioningState = "Failed"
			dep.Error = fmt.Sprintf("deploy %s/%s: %v", res.Type, res.Name, err)
			dep.Duration = time.Since(start).String()
			return dep, err
		}
		dep.Resources = append(dep.Resources, *deployed)
	}

	dep.ProvisioningState = "Succeeded"
	dep.Duration = time.Since(start).String()

	// Resolve outputs
	if result.Template.Outputs != nil {
		dep.Outputs = make(map[string]interface{}, len(result.Template.Outputs))
		for name, out := range result.Template.Outputs {
			dep.Outputs[name] = map[string]interface{}{
				"type":  out.Type,
				"value": out.Value,
			}
		}
	}

	return dep, nil
}

// Validate checks that a template can be deployed without actually creating resources.
func (d *Deployer) Validate(result *ParseResult, resourceGroup string) error {
	for _, res := range result.Resources {
		if res.Name == "" {
			return fmt.Errorf("resource of type %q has no name", res.Type)
		}
		if res.Location == "" {
			return fmt.Errorf("resource %q has no location", res.Name)
		}
		if _, ok := resourceTypeToCommand(res.Type); !ok {
			return fmt.Errorf("unsupported resource type %q for resource %q", res.Type, res.Name)
		}
	}
	return nil
}

// deployResource maps an ARM resource type to az CLI commands and executes them.
func (d *Deployer) deployResource(ctx context.Context, res ResolvedResource, rg string) (*DeployedResource, error) {
	cmdInfo, ok := resourceTypeToCommand(res.Type)
	if !ok {
		return nil, fmt.Errorf("unsupported resource type %q", res.Type)
	}

	args := cmdInfo.buildArgs(res, rg)
	_, err := d.exec.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("deploy %s %q: %w", res.Type, res.Name, err)
	}

	return &DeployedResource{
		ID:       fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/%s/providers/%s/%s", rg, res.Type, res.Name),
		Type:     res.Type,
		Name:     res.Name,
		Location: res.Location,
	}, nil
}

// commandMapping describes how to map an ARM resource type to az CLI args.
type commandMapping struct {
	prefix []string // e.g. ["group", "create"]
}

// buildArgs constructs CLI args for creating this resource.
func (cm *commandMapping) buildArgs(res ResolvedResource, rg string) []string {
	args := make([]string, len(cm.prefix))
	copy(args, cm.prefix)
	args = append(args, "--name", res.Name)

	// Most resources need --resource-group
	switch cm.prefix[0] {
	case "group":
		// group create uses --location directly
	default:
		args = append(args, "--resource-group", rg)
	}

	if res.Location != "" {
		args = append(args, "--location", res.Location)
	}

	// Resource-specific args from properties
	switch cm.prefix[0] {
	case "storage":
		if res.SKU != nil {
			if skuName, ok := res.SKU["name"].(string); ok {
				args = append(args, "--sku", skuName)
			}
		}
		if res.Kind != "" {
			args = append(args, "--kind", res.Kind)
		}
	case "keyvault":
		if res.SKU != nil {
			if skuName, ok := res.SKU["name"].(string); ok {
				args = append(args, "--sku", skuName)
			}
		}
	case "network":
		if strings.Contains(strings.Join(cm.prefix, " "), "vnet") {
			if res.Properties != nil {
				if as, ok := res.Properties["addressSpace"].(map[string]interface{}); ok {
					if prefixes, ok := as["addressPrefixes"].([]interface{}); ok && len(prefixes) > 0 {
						args = append(args, "--address-prefix", fmt.Sprintf("%v", prefixes[0]))
					}
				}
			}
		}
	case "vm":
		if res.Properties != nil {
			if hw, ok := res.Properties["hardwareProfile"].(map[string]interface{}); ok {
				if size, ok := hw["vmSize"].(string); ok {
					args = append(args, "--size", size)
				}
			}
			if img, ok := res.Properties["storageProfile"].(map[string]interface{}); ok {
				if ref, ok := img["imageReference"].(map[string]interface{}); ok {
					urn := fmt.Sprintf("%v:%v:%v:%v",
						ref["publisher"], ref["offer"], ref["sku"], ref["version"])
					args = append(args, "--image", urn)
				}
			}
		}
	}

	// Tags
	if len(res.Tags) > 0 {
		args = append(args, "--tags")
		for k, v := range res.Tags {
			args = append(args, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return args
}

// resourceTypeToCommand maps ARM resource type strings to az CLI commands.
func resourceTypeToCommand(resourceType string) (*commandMapping, bool) {
	rt := strings.ToLower(resourceType)
	mappings := map[string]*commandMapping{
		"microsoft.resources/resourcegroups":           {prefix: []string{"group", "create"}},
		"microsoft.storage/storageaccounts":            {prefix: []string{"storage", "account", "create"}},
		"microsoft.keyvault/vaults":                    {prefix: []string{"keyvault", "create"}},
		"microsoft.network/virtualnetworks":            {prefix: []string{"network", "vnet", "create"}},
		"microsoft.network/networksecuritygroups":      {prefix: []string{"network", "nsg", "create"}},
		"microsoft.network/publicipaddresses":          {prefix: []string{"network", "public-ip", "create"}},
		"microsoft.compute/virtualmachines":            {prefix: []string{"vm", "create"}},
	}
	cm, ok := mappings[rt]
	return cm, ok
}

// DeploymentToJSON serializes a deployment to indented JSON.
func DeploymentToJSON(dep *Deployment) ([]byte, error) {
	return json.MarshalIndent(dep, "", "  ")
}
