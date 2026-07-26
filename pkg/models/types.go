// Package models defines shared data types used across azfloci.
package models

import "time"

// AzureAccount represents the output of `az account show`.
type AzureAccount struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	State            string `json:"state"`
	IsDefault        bool   `json:"isDefault"`
	TenantID         string `json:"tenantId"`
	HomeTenantID     string `json:"homeTenantId"`
	EnvironmentName  string `json:"environmentName"`
	User             AzureUser `json:"user"`
}

// AzureUser represents the user block in az account show.
type AzureUser struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ResourceGroup represents an Azure resource group.
type ResourceGroup struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Location   string            `json:"location"`
	Tags       map[string]string `json:"tags,omitempty"`
	Properties ResourceGroupProperties `json:"properties,omitempty"`
}

// ResourceGroupProperties holds provisioning state.
type ResourceGroupProperties struct {
	ProvisioningState string `json:"provisioningState"`
}

// Resource represents a generic Azure resource.
type Resource struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Location   string            `json:"location"`
	Tags       map[string]string `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// WorkflowStep defines a single step in a workflow pipeline.
type WorkflowStep struct {
	Name       string            `yaml:"name" json:"name"`
	Command    string            `yaml:"command" json:"command"`
	Args       map[string]string `yaml:"args,omitempty" json:"args,omitempty"`
	DependsOn  []string          `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	OnError    string            `yaml:"on_error,omitempty" json:"on_error,omitempty"` // "continue", "abort", "retry"
	MaxRetries int               `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
}

// Workflow defines a pipeline of steps to provision/manage Azure resources.
type Workflow struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Steps       []WorkflowStep `yaml:"steps" json:"steps"`
}

// StepResult captures the outcome of executing a workflow step.
type StepResult struct {
	StepName  string        `json:"step_name"`
	Status    string        `json:"status"` // "success", "failed", "skipped"
	Output    string        `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Retries   int           `json:"retries"`
}

// WorkflowResult captures the outcome of a full workflow execution.
type WorkflowResult struct {
	WorkflowName string       `json:"workflow_name"`
	Status       string       `json:"status"` // "success", "failed", "partial"
	Steps        []StepResult `json:"steps"`
	Duration     time.Duration `json:"duration"`
}

// Config holds the azfloci configuration.
type Config struct {
	Subscription string            `yaml:"subscription" json:"subscription" mapstructure:"subscription"`
	Location     string            `yaml:"location" json:"location" mapstructure:"location"`
	OutputFormat string            `yaml:"output_format" json:"output_format" mapstructure:"output_format"`
	Verbose      bool              `yaml:"verbose" json:"verbose" mapstructure:"verbose"`
	Tags         map[string]string `yaml:"tags,omitempty" json:"tags,omitempty" mapstructure:"tags"`
	Defaults     ConfigDefaults    `yaml:"defaults,omitempty" json:"defaults,omitempty" mapstructure:"defaults"`
}

// ConfigDefaults holds default values for resource creation.
type ConfigDefaults struct {
	ResourceGroup string `yaml:"resource_group" json:"resource_group" mapstructure:"resource_group"`
	Location      string `yaml:"location" json:"location" mapstructure:"location"`
}
