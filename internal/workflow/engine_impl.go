package workflow

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/pkg/models"
	"github.com/ankurCES/floc-zure/pkg/utils"
	"gopkg.in/yaml.v3"
)

// EngineImpl is the default workflow engine backed by an Azure CLI executor.
type EngineImpl struct {
	exec azure.CLIExecutor
}

// NewEngine creates a workflow engine that delegates to the given executor.
func NewEngine(exec azure.CLIExecutor) *EngineImpl {
	return &EngineImpl{exec: exec}
}

// LoadWorkflow reads and parses a YAML workflow definition.
func (e *EngineImpl) LoadWorkflow(path string) (*models.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}
	var wf models.Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parse workflow YAML: %w", err)
	}
	return &wf, nil
}

// Validate checks a workflow for structural errors:
//   - at least one step
//   - unique step names
//   - all depends_on references resolve
//   - no dependency cycles
func (e *EngineImpl) Validate(wf *models.Workflow) error {
	if len(wf.Steps) == 0 {
		return &utils.ValidationError{Field: "steps", Message: "workflow has no steps"}
	}

	names := make(map[string]bool, len(wf.Steps))
	for _, s := range wf.Steps {
		if s.Name == "" {
			return &utils.ValidationError{Field: "step.name", Message: "step name is required"}
		}
		if names[s.Name] {
			return &utils.ValidationError{Field: "step.name", Message: fmt.Sprintf("duplicate step name %q", s.Name)}
		}
		names[s.Name] = true
	}

	// Check depends_on references
	for _, s := range wf.Steps {
		for _, dep := range s.DependsOn {
			if !names[dep] {
				return &utils.ValidationError{
					Field:   "step.depends_on",
					Message: fmt.Sprintf("step %q depends on unknown step %q", s.Name, dep),
				}
			}
		}
	}

	// Cycle detection via topological sort
	_, err := e.topoSort(wf.Steps)
	return err
}

// Execute runs all steps in topological (dependency) order.
// Independent steps run concurrently. Respects on_error and retry policies.
func (e *EngineImpl) Execute(ctx context.Context, wf *models.Workflow) (*models.WorkflowResult, error) {
	start := time.Now()

	sorted, err := e.topoSort(wf.Steps)
	if err != nil {
		return nil, err
	}

	// Build step map and result map
	stepMap := make(map[string]*models.WorkflowStep, len(wf.Steps))
	for i := range wf.Steps {
		stepMap[wf.Steps[i].Name] = &wf.Steps[i]
	}

	results := &sync.Map{}    // name -> *models.StepResult
	completed := &sync.Map{}  // name -> bool (true = done)

	// Group steps by "level" — steps whose deps are all in earlier levels
	levels := e.buildLevels(sorted, stepMap)

	var allResults []models.StepResult
	aborted := false

	for _, level := range levels {
		if aborted {
			// Skip remaining levels, mark as skipped
			for _, name := range level {
				allResults = append(allResults, models.StepResult{
					StepName: name,
					Status:   "skipped",
				})
			}
			continue
		}

		var wg sync.WaitGroup
		levelResults := make([]*models.StepResult, len(level))

		for i, name := range level {
			wg.Add(1)
			go func(idx int, stepName string) {
				defer wg.Done()
				step := stepMap[stepName]
				sr := e.executeWithRetry(ctx, step)
				levelResults[idx] = sr
				results.Store(stepName, sr)
				completed.Store(stepName, true)
			}(i, name)
		}
		wg.Wait()

		for _, sr := range levelResults {
			allResults = append(allResults, *sr)
			if sr.Status == "failed" {
				step := stepMap[sr.StepName]
				if step.OnError == "" || step.OnError == "abort" {
					aborted = true
				}
				// "continue" → keep going; "retry" already handled in executeWithRetry
			}
		}
	}

	status := "success"
	for _, r := range allResults {
		if r.Status == "failed" {
			status = "failed"
			break
		}
		if r.Status == "skipped" && status == "success" {
			status = "partial"
		}
	}

	return &models.WorkflowResult{
		WorkflowName: wf.Name,
		Status:       status,
		Steps:        allResults,
		Duration:     time.Since(start),
	}, nil
}

// ExecuteStep runs a single workflow step against the Azure CLI.
func (e *EngineImpl) ExecuteStep(ctx context.Context, step *models.WorkflowStep) (*models.StepResult, error) {
	sr := e.executeOne(ctx, step)
	if sr.Status == "failed" {
		return sr, fmt.Errorf("step %q failed: %s", step.Name, sr.Error)
	}
	return sr, nil
}

// --- internal helpers ---

func (e *EngineImpl) executeOne(ctx context.Context, step *models.WorkflowStep) *models.StepResult {
	start := time.Now()

	args := e.buildArgs(step)
	result, err := e.exec.Run(ctx, args...)

	sr := &models.StepResult{
		StepName: step.Name,
		Duration: time.Since(start),
	}
	if err != nil {
		sr.Status = "failed"
		sr.Error = err.Error()
		if result != nil {
			sr.Output = result.Stderr
		}
	} else {
		sr.Status = "success"
		if result != nil {
			sr.Output = result.Stdout
		}
	}
	return sr
}

func (e *EngineImpl) executeWithRetry(ctx context.Context, step *models.WorkflowStep) *models.StepResult {
	maxRetries := step.MaxRetries
	if step.OnError != "retry" {
		maxRetries = 0
	}

	var sr *models.StepResult
	for attempt := 0; attempt <= maxRetries; attempt++ {
		sr = e.executeOne(ctx, step)
		sr.Retries = attempt
		if sr.Status == "success" {
			return sr
		}
		if attempt < maxRetries {
			// Exponential backoff: 1s, 2s, 4s...
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
		}
	}
	return sr
}

// buildArgs converts a WorkflowStep into az CLI arguments.
// Command format: "group create" → ["group", "create"]
// Args map: {"--name": "rg1"} → ["--name", "rg1"]
func (e *EngineImpl) buildArgs(step *models.WorkflowStep) []string {
	parts := strings.Fields(step.Command)
	// Sort arg keys for deterministic ordering in tests
	keys := make([]string, 0, len(step.Args))
	for k := range step.Args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.HasPrefix(k, "--") {
			parts = append(parts, k, step.Args[k])
		} else {
			parts = append(parts, "--"+k, step.Args[k])
		}
	}
	return parts
}

// topoSort performs Kahn's algorithm for topological ordering.
// Returns an error if a cycle is detected.
func (e *EngineImpl) topoSort(steps []models.WorkflowStep) ([]string, error) {
	inDegree := make(map[string]int)
	graph := make(map[string][]string) // dep -> dependents
	for _, s := range steps {
		if _, ok := inDegree[s.Name]; !ok {
			inDegree[s.Name] = 0
		}
		for _, dep := range s.DependsOn {
			graph[dep] = append(graph[dep], s.Name)
			inDegree[s.Name]++
		}
	}

	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue) // deterministic

	var sorted []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, node)

		for _, dependent := range graph[node] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				sort.Strings(queue)
			}
		}
	}

	if len(sorted) != len(steps) {
		return nil, &utils.ValidationError{
			Field:   "workflow",
			Message: "dependency cycle detected",
		}
	}
	return sorted, nil
}

// buildLevels groups steps into concurrent execution levels.
// Level N contains steps whose dependencies are all in levels < N.
func (e *EngineImpl) buildLevels(sorted []string, stepMap map[string]*models.WorkflowStep) [][]string {
	level := make(map[string]int)
	for _, name := range sorted {
		step := stepMap[name]
		maxDepLevel := -1
		for _, dep := range step.DependsOn {
			if l, ok := level[dep]; ok && l > maxDepLevel {
				maxDepLevel = l
			}
		}
		level[name] = maxDepLevel + 1
	}

	// Group by level
	maxLevel := 0
	for _, l := range level {
		if l > maxLevel {
			maxLevel = l
		}
	}
	levels := make([][]string, maxLevel+1)
	for _, name := range sorted {
		l := level[name]
		levels[l] = append(levels[l], name)
	}
	return levels
}
