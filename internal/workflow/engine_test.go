package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/pkg/models"
)

// mockExec is a test-only CLIExecutor.
type mockExec struct {
	runFunc func(ctx context.Context, args ...string) (*azure.CLIResult, error)
	calls   [][]string
}

func (m *mockExec) Run(ctx context.Context, args ...string) (*azure.CLIResult, error) {
	m.calls = append(m.calls, args)
	if m.runFunc != nil {
		return m.runFunc(ctx, args...)
	}
	return &azure.CLIResult{Stdout: "ok"}, nil
}
func (m *mockExec) RunJSON(ctx context.Context, dest interface{}, args ...string) error {
	m.calls = append(m.calls, args)
	return json.Unmarshal([]byte("{}"), dest)
}
func (m *mockExec) GetAccount(ctx context.Context) (*models.AzureAccount, error) {
	return &models.AzureAccount{ID: "mock"}, nil
}
func (m *mockExec) IsAuthenticated(ctx context.Context) (bool, error) { return true, nil }
func (m *mockExec) SetSubscription(ctx context.Context, id string) error { return nil }

// --- YAML Loading Tests ---

func writeWorkflowYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadWorkflow_Valid(t *testing.T) {
	yaml := `
name: test-wf
description: A test workflow
steps:
  - name: create-rg
    command: group create
    args:
      name: rg1
      location: eastus
  - name: create-sa
    command: storage account create
    args:
      name: sa1
    depends_on:
      - create-rg
`
	path := writeWorkflowYAML(t, yaml)
	eng := NewEngine(&mockExec{})
	wf, err := eng.LoadWorkflow(path)
	if err != nil {
		t.Fatalf("LoadWorkflow failed: %v", err)
	}
	if wf.Name != "test-wf" {
		t.Errorf("expected name test-wf, got %s", wf.Name)
	}
	if len(wf.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(wf.Steps))
	}
	if wf.Steps[1].DependsOn[0] != "create-rg" {
		t.Errorf("expected depends_on create-rg, got %v", wf.Steps[1].DependsOn)
	}
}

func TestLoadWorkflow_FileNotFound(t *testing.T) {
	eng := NewEngine(&mockExec{})
	_, err := eng.LoadWorkflow("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadWorkflow_InvalidYAML(t *testing.T) {
	path := writeWorkflowYAML(t, "{{{{invalid yaml")
	eng := NewEngine(&mockExec{})
	_, err := eng.LoadWorkflow(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// --- Validation Tests ---

func TestValidate_EmptySteps(t *testing.T) {
	eng := NewEngine(&mockExec{})
	err := eng.Validate(&models.Workflow{Name: "empty", Steps: nil})
	if err == nil {
		t.Fatal("expected error for empty steps")
	}
}

func TestValidate_DuplicateNames(t *testing.T) {
	eng := NewEngine(&mockExec{})
	wf := &models.Workflow{
		Name: "dup",
		Steps: []models.WorkflowStep{
			{Name: "a", Command: "x"},
			{Name: "a", Command: "y"},
		},
	}
	err := eng.Validate(wf)
	if err == nil {
		t.Fatal("expected error for duplicate names")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected 'duplicate' in error, got: %v", err)
	}
}

func TestValidate_UnknownDependency(t *testing.T) {
	eng := NewEngine(&mockExec{})
	wf := &models.Workflow{
		Name: "bad-dep",
		Steps: []models.WorkflowStep{
			{Name: "a", Command: "x", DependsOn: []string{"nonexistent"}},
		},
	}
	err := eng.Validate(wf)
	if err == nil {
		t.Fatal("expected error for unknown dependency")
	}
	if !strings.Contains(err.Error(), "unknown step") {
		t.Errorf("expected 'unknown step' in error, got: %v", err)
	}
}

func TestValidate_CycleDetection(t *testing.T) {
	eng := NewEngine(&mockExec{})
	wf := &models.Workflow{
		Name: "cycle",
		Steps: []models.WorkflowStep{
			{Name: "a", Command: "x", DependsOn: []string{"b"}},
			{Name: "b", Command: "y", DependsOn: []string{"a"}},
		},
	}
	err := eng.Validate(wf)
	if err == nil {
		t.Fatal("expected error for cycle")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected 'cycle' in error, got: %v", err)
	}
}

func TestValidate_EmptyStepName(t *testing.T) {
	eng := NewEngine(&mockExec{})
	wf := &models.Workflow{
		Name:  "noname",
		Steps: []models.WorkflowStep{{Name: "", Command: "x"}},
	}
	err := eng.Validate(wf)
	if err == nil {
		t.Fatal("expected error for empty step name")
	}
}

func TestValidate_ValidDAG(t *testing.T) {
	eng := NewEngine(&mockExec{})
	wf := &models.Workflow{
		Name: "diamond",
		Steps: []models.WorkflowStep{
			{Name: "a", Command: "x"},
			{Name: "b", Command: "y", DependsOn: []string{"a"}},
			{Name: "c", Command: "z", DependsOn: []string{"a"}},
			{Name: "d", Command: "w", DependsOn: []string{"b", "c"}},
		},
	}
	err := eng.Validate(wf)
	if err != nil {
		t.Fatalf("valid DAG should pass: %v", err)
	}
}

// --- Execution Tests ---

func TestExecute_SimpleLinear(t *testing.T) {
	mock := &mockExec{}
	eng := NewEngine(mock)
	wf := &models.Workflow{
		Name: "linear",
		Steps: []models.WorkflowStep{
			{Name: "step1", Command: "group create", Args: map[string]string{"name": "rg1"}},
			{Name: "step2", Command: "group show", DependsOn: []string{"step1"}},
		},
	}
	result, err := eng.Execute(context.Background(), wf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Status != "success" {
		t.Errorf("expected success, got %s", result.Status)
	}
	if len(result.Steps) != 2 {
		t.Errorf("expected 2 step results, got %d", len(result.Steps))
	}
	// Verify both steps succeeded
	for _, sr := range result.Steps {
		if sr.Status != "success" {
			t.Errorf("step %s: expected success, got %s", sr.StepName, sr.Status)
		}
	}
}

func TestExecute_ParallelSteps(t *testing.T) {
	callOrder := make([]string, 0)
	mock := &mockExec{
		runFunc: func(ctx context.Context, args ...string) (*azure.CLIResult, error) {
			callOrder = append(callOrder, args[0])
			return &azure.CLIResult{Stdout: "ok"}, nil
		},
	}
	eng := NewEngine(mock)
	wf := &models.Workflow{
		Name: "parallel",
		Steps: []models.WorkflowStep{
			{Name: "a", Command: "cmd-a"},
			{Name: "b", Command: "cmd-b"},
			{Name: "c", Command: "cmd-c"},
		},
	}
	result, err := eng.Execute(context.Background(), wf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Status != "success" {
		t.Errorf("expected success, got %s", result.Status)
	}
	if len(result.Steps) != 3 {
		t.Errorf("expected 3 step results, got %d", len(result.Steps))
	}
}

func TestExecute_AbortOnFailure(t *testing.T) {
	mock := &mockExec{
		runFunc: func(ctx context.Context, args ...string) (*azure.CLIResult, error) {
			if args[0] == "fail-cmd" {
				return &azure.CLIResult{ExitCode: 1, Stderr: "boom"}, fmt.Errorf("command failed")
			}
			return &azure.CLIResult{Stdout: "ok"}, nil
		},
	}
	eng := NewEngine(mock)
	wf := &models.Workflow{
		Name: "abort-test",
		Steps: []models.WorkflowStep{
			{Name: "step1", Command: "fail-cmd"},                               // fails, default on_error=abort
			{Name: "step2", Command: "good-cmd", DependsOn: []string{"step1"}}, // should be skipped
		},
	}
	result, err := eng.Execute(context.Background(), wf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Status != "failed" {
		t.Errorf("expected failed, got %s", result.Status)
	}
	if result.Steps[1].Status != "skipped" {
		t.Errorf("step2 should be skipped, got %s", result.Steps[1].Status)
	}
}

func TestExecute_ContinueOnFailure(t *testing.T) {
	mock := &mockExec{
		runFunc: func(ctx context.Context, args ...string) (*azure.CLIResult, error) {
			if args[0] == "fail-cmd" {
				return nil, fmt.Errorf("step failed")
			}
			return &azure.CLIResult{Stdout: "ok"}, nil
		},
	}
	eng := NewEngine(mock)
	wf := &models.Workflow{
		Name: "continue-test",
		Steps: []models.WorkflowStep{
			{Name: "step1", Command: "fail-cmd", OnError: "continue"},
			{Name: "step2", Command: "good-cmd", DependsOn: []string{"step1"}},
		},
	}
	result, err := eng.Execute(context.Background(), wf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	// step2 should still execute because step1 has on_error=continue
	if result.Steps[1].Status != "success" {
		t.Errorf("step2 should succeed, got %s", result.Steps[1].Status)
	}
}

func TestExecuteStep_Success(t *testing.T) {
	mock := &mockExec{}
	eng := NewEngine(mock)
	step := &models.WorkflowStep{
		Name:    "single",
		Command: "group list",
	}
	sr, err := eng.ExecuteStep(context.Background(), step)
	if err != nil {
		t.Fatalf("ExecuteStep failed: %v", err)
	}
	if sr.Status != "success" {
		t.Errorf("expected success, got %s", sr.Status)
	}
}

func TestExecuteStep_Failure(t *testing.T) {
	mock := &mockExec{
		runFunc: func(ctx context.Context, args ...string) (*azure.CLIResult, error) {
			return nil, fmt.Errorf("execution failed")
		},
	}
	eng := NewEngine(mock)
	step := &models.WorkflowStep{Name: "bad", Command: "bad cmd"}
	sr, err := eng.ExecuteStep(context.Background(), step)
	if err == nil {
		t.Fatal("expected error")
	}
	if sr.Status != "failed" {
		t.Errorf("expected failed, got %s", sr.Status)
	}
}

// --- buildArgs tests ---

func TestBuildArgs(t *testing.T) {
	eng := NewEngine(&mockExec{})
	step := &models.WorkflowStep{
		Command: "group create",
		Args: map[string]string{
			"name":     "rg1",
			"location": "eastus",
		},
	}
	args := eng.buildArgs(step)
	joined := strings.Join(args, " ")
	if !strings.HasPrefix(joined, "group create") {
		t.Errorf("expected prefix 'group create', got: %s", joined)
	}
	if !strings.Contains(joined, "--location eastus") {
		t.Errorf("expected --location eastus: %s", joined)
	}
	if !strings.Contains(joined, "--name rg1") {
		t.Errorf("expected --name rg1: %s", joined)
	}
}

func TestBuildArgs_ExplicitDashes(t *testing.T) {
	eng := NewEngine(&mockExec{})
	step := &models.WorkflowStep{
		Command: "resource show",
		Args:    map[string]string{"--ids": "/sub/rg/res"},
	}
	args := eng.buildArgs(step)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--ids /sub/rg/res") {
		t.Errorf("expected --ids with dashes preserved: %s", joined)
	}
}

// --- topoSort tests ---

func TestTopoSort_Linear(t *testing.T) {
	eng := NewEngine(&mockExec{})
	steps := []models.WorkflowStep{
		{Name: "c", Command: "x", DependsOn: []string{"b"}},
		{Name: "b", Command: "y", DependsOn: []string{"a"}},
		{Name: "a", Command: "z"},
	}
	sorted, err := eng.topoSort(steps)
	if err != nil {
		t.Fatalf("topoSort failed: %v", err)
	}
	if sorted[0] != "a" || sorted[1] != "b" || sorted[2] != "c" {
		t.Errorf("expected [a, b, c], got %v", sorted)
	}
}

func TestTopoSort_Diamond(t *testing.T) {
	eng := NewEngine(&mockExec{})
	steps := []models.WorkflowStep{
		{Name: "a", Command: "x"},
		{Name: "b", Command: "y", DependsOn: []string{"a"}},
		{Name: "c", Command: "z", DependsOn: []string{"a"}},
		{Name: "d", Command: "w", DependsOn: []string{"b", "c"}},
	}
	sorted, err := eng.topoSort(steps)
	if err != nil {
		t.Fatalf("topoSort failed: %v", err)
	}
	if sorted[0] != "a" || sorted[len(sorted)-1] != "d" {
		t.Errorf("expected a first, d last: %v", sorted)
	}
}

// --- buildLevels tests ---

func TestBuildLevels_Diamond(t *testing.T) {
	eng := NewEngine(&mockExec{})
	steps := []models.WorkflowStep{
		{Name: "a", Command: "x"},
		{Name: "b", Command: "y", DependsOn: []string{"a"}},
		{Name: "c", Command: "z", DependsOn: []string{"a"}},
		{Name: "d", Command: "w", DependsOn: []string{"b", "c"}},
	}
	stepMap := make(map[string]*models.WorkflowStep)
	for i := range steps {
		stepMap[steps[i].Name] = &steps[i]
	}
	levels := eng.buildLevels([]string{"a", "b", "c", "d"}, stepMap)
	if len(levels) != 3 {
		t.Fatalf("expected 3 levels, got %d", len(levels))
	}
	if levels[0][0] != "a" {
		t.Errorf("level 0 should be [a], got %v", levels[0])
	}
	if len(levels[1]) != 2 {
		t.Errorf("level 1 should have 2 items (b,c), got %v", levels[1])
	}
	if levels[2][0] != "d" {
		t.Errorf("level 2 should be [d], got %v", levels[2])
	}
}

// --- Integration: Load + Validate + Execute ---

func TestFullWorkflow_LoadValidateExecute(t *testing.T) {
	yaml := `
name: full-test
steps:
  - name: init
    command: group create
    args:
      name: rg-test
      location: eastus
  - name: deploy
    command: webapp create
    depends_on: [init]
  - name: verify
    command: webapp show
    depends_on: [deploy]
`
	path := writeWorkflowYAML(t, yaml)
	mock := &mockExec{}
	eng := NewEngine(mock)

	wf, err := eng.LoadWorkflow(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := eng.Validate(wf); err != nil {
		t.Fatalf("validate: %v", err)
	}
	result, err := eng.Execute(context.Background(), wf)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result.Status != "success" {
		t.Errorf("expected success, got %s", result.Status)
	}
	if len(result.Steps) != 3 {
		t.Errorf("expected 3 step results, got %d", len(result.Steps))
	}
	if len(mock.calls) != 3 {
		t.Errorf("expected 3 CLI calls, got %d", len(mock.calls))
	}
}
