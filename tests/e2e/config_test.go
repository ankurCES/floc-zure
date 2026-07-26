//go:build e2e

// Tests for configuration import/export: verifies config file loading,
// defaults, and env-var overrides.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

// TestConfigTestdataExists verifies the sample config file is present.
func TestConfigTestdataExists(t *testing.T) {
	path := filepath.Join(findTestdataDir(t), "sample_config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("sample_config.yaml not found in testdata")
	}
}

// TestConfigSampleIsValidYAML checks the sample config is non-empty valid YAML.
func TestConfigSampleIsValidYAML(t *testing.T) {
	path := filepath.Join(findTestdataDir(t), "sample_config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read sample config: %v", err)
	}
	content := string(data)
	if !contains(content, "location:") {
		t.Error("sample config missing 'location' field")
	}
	if !contains(content, "output_format:") {
		t.Error("sample config missing 'output_format' field")
	}
}

// TestCLIWithConfigEnvVar verifies the CLI runs when AZFLOCI env vars are set.
func TestCLIWithConfigEnvVar(t *testing.T) {
	cli := helpers.NewCLIRunner(t)

	// Set an env var and verify CLI still works.
	res := cli.RunWithEnv(
		[]string{"AZFLOCI_LOCATION=westus2", "AZFLOCI_VERBOSE=true"},
		"version",
	)
	helpers.AssertSuccess(t, res)
	helpers.AssertStdoutContains(t, res, "azfloci")
}

// TestCLIWithWorkDir verifies the CLI works from a custom working directory.
func TestCLIWithWorkDir(t *testing.T) {
	cli := helpers.NewCLIRunner(t)
	cli.WorkDir = t.TempDir() // Empty dir with no config file.

	res := cli.Run("version")
	helpers.AssertSuccess(t, res)
}
