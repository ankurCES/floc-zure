// Package workflow implements the pipeline execution engine.
package workflow

import (
	"context"

	"github.com/ankurCES/floc-zure/pkg/models"
)

// Engine orchestrates multi-step Azure provisioning workflows.
// Steps can declare dependencies (DAG), error handling policies,
// and retry logic. The engine resolves execution order via
// topological sort and runs independent steps concurrently.
type Engine interface {
	// LoadWorkflow parses a workflow definition from a YAML file.
	LoadWorkflow(path string) (*models.Workflow, error)

	// Validate checks a workflow for cycles, missing deps, and bad commands.
	Validate(wf *models.Workflow) error

	// Execute runs a workflow to completion, respecting step dependencies.
	Execute(ctx context.Context, wf *models.Workflow) (*models.WorkflowResult, error)

	// ExecuteStep runs a single step (used internally and for retries).
	ExecuteStep(ctx context.Context, step *models.WorkflowStep) (*models.StepResult, error)
}

// StepExecutor is the lower-level interface for running a single step's
// command against Azure. The Engine delegates to this.
type StepExecutor interface {
	Execute(ctx context.Context, command string, args map[string]string) (string, error)
}
