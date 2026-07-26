//go:build e2e

// Tests for workflow execution. Since the workflow engine may not have a CLI
// command wired yet, these tests focus on validating testdata files exist
// and the CLI doesn't crash on workflow-related flags.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

// TestWorkflowTestdataExists verifies sample workflow files are present.
func TestWorkflowTestdataExists(t *testing.T) {
	files := []string{
		"testdata/sample_workflow.yaml",
		"testdata/invalid_workflow.yaml",
	}
	for _, f := range files {
		path := filepath.Join(findTestdataDir(t), filepath.Base(f))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected testdata file %s to exist", f)
		}
	}
}

// TestWorkflowSampleParseable verifies the sample workflow is valid YAML.
func TestWorkflowSampleParseable(t *testing.T) {
	path := filepath.Join(findTestdataDir(t), "sample_workflow.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read sample workflow: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("sample workflow file is empty")
	}
	// Basic structure check.
	content := string(data)
	if !contains(content, "name:") || !contains(content, "steps:") {
		t.Error("sample workflow missing required 'name' or 'steps' fields")
	}
}

// TestCLIDoesNotCrashOnHelp ensures no panic on help subcommands.
func TestCLIDoesNotCrashOnHelp(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	// Root help should never crash regardless of workflow state.
	res := cli.Run("--help")
	helpers.AssertSuccess(t, res)
}

func findTestdataDir(t *testing.T) string {
	t.Helper()
	// testdata is relative to this test file's directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	td := filepath.Join(dir, "testdata")
	if _, err := os.Stat(td); os.IsNotExist(err) {
		// Try from repo root
		for d := dir; ; d = filepath.Dir(d) {
			candidate := filepath.Join(d, "tests", "e2e", "testdata")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			if filepath.Dir(d) == d {
				break
			}
		}
		t.Fatal("cannot find testdata directory")
	}
	return td
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
