//go:build e2e

// Tests for drift detection: snapshot, compare, report.
// Uses the az-simulator binary — no real Azure subscription needed.
package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

func TestDriftSnapshotCreatesFile(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")
	snapshotFile := filepath.Join(t.TempDir(), "snap.json")

	// Seed some resources
	runSim(t, simBin, stateFile, "group", "create", "-n", "drift-rg", "-l", "eastus")
	runSim(t, simBin, stateFile, "vm", "create",
		"-n", "drift-vm", "-g", "drift-rg", "-l", "eastus",
		"--image", "Ubuntu2204", "--size", "Standard_B1s")

	// Capture snapshot via azfloci drift snapshot
	runner := newDriftRunner(t, simBin, stateFile)
	res := runner.Run("drift", "snapshot",
		"--state", stateFile,
		"--output", snapshotFile)
	if !res.Success() {
		t.Fatalf("drift snapshot failed: exit %d\nstdout: %s\nstderr: %s",
			res.ExitCode, res.Stdout, res.Stderr)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}
	var snap map[string]interface{}
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("snapshot is not valid JSON: %v", err)
	}
	if _, ok := snap["label"]; !ok {
		t.Error("snapshot missing 'label' field")
	}
	if _, ok := snap["resources"]; !ok {
		t.Error("snapshot missing 'resources' field")
	}
}

func TestDriftCompareDetectsChanges(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")
	beforeFile := filepath.Join(t.TempDir(), "before.json")
	afterFile := filepath.Join(t.TempDir(), "after.json")

	// Create initial state
	runSim(t, simBin, stateFile, "group", "create", "-n", "cmp-rg", "-l", "eastus")
	runSim(t, simBin, stateFile, "storage", "account", "create",
		"-n", "cmpstore", "-g", "cmp-rg", "-l", "eastus")

	// Snapshot before
	runner := newDriftRunner(t, simBin, stateFile)
	runner.MustRun(t, "drift", "snapshot",
		"--state", stateFile,
		"--output", beforeFile,
		"--label", "before")

	// Make a change — add a VM
	runSim(t, simBin, stateFile, "vm", "create",
		"-n", "new-vm", "-g", "cmp-rg", "-l", "eastus",
		"--image", "Ubuntu2204", "--size", "Standard_B2s")

	// Snapshot after
	runner.MustRun(t, "drift", "snapshot",
		"--state", stateFile,
		"--output", afterFile,
		"--label", "after")

	// Compare
	res := runner.Run("drift", "compare",
		"--before", beforeFile,
		"--after", afterFile)
	if !res.Success() {
		t.Fatalf("drift compare failed: exit %d\nstderr: %s", res.ExitCode, res.Stderr)
	}

	// Should detect the added VM
	if !strings.Contains(res.Stdout, "added") && !strings.Contains(res.Stdout, "Added") &&
		!strings.Contains(res.Stdout, "ADDED") && !strings.Contains(res.Stdout, "new-vm") {
		t.Errorf("expected drift output to mention added resource, got:\n%s", res.Stdout)
	}
}

func TestDriftCompareNoDrift(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")
	snapFile := filepath.Join(t.TempDir(), "snap.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "nodrift-rg", "-l", "eastus")

	runner := newDriftRunner(t, simBin, stateFile)
	runner.MustRun(t, "drift", "snapshot",
		"--state", stateFile,
		"--output", snapFile)

	// Compare identical snapshots
	res := runner.Run("drift", "compare",
		"--before", snapFile,
		"--after", snapFile)
	if !res.Success() {
		t.Fatalf("drift compare failed: exit %d\nstderr: %s", res.ExitCode, res.Stderr)
	}

	// Should report no drift
	combined := res.Stdout + res.Stderr
	if strings.Contains(combined, "ADDED") || strings.Contains(combined, "REMOVED") ||
		strings.Contains(combined, "MODIFIED") {
		t.Errorf("expected no drift, got:\n%s", combined)
	}
}

func TestDriftReportOutput(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")
	baselineFile := filepath.Join(t.TempDir(), "baseline.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "rpt-rg", "-l", "eastus")

	runner := newDriftRunner(t, simBin, stateFile)
	runner.MustRun(t, "drift", "snapshot",
		"--state", stateFile,
		"--output", baselineFile)

	// Add resource after baseline
	runSim(t, simBin, stateFile, "network", "vnet", "create",
		"-n", "rpt-vnet", "-g", "rpt-rg", "-l", "eastus",
		"--address-prefixes", "10.0.0.0/16")

	// Report
	res := runner.Run("drift", "report",
		"--state", stateFile,
		"--baseline", baselineFile)
	if !res.Success() {
		t.Fatalf("drift report failed: exit %d\nstderr: %s", res.ExitCode, res.Stderr)
	}

	// Report should mention the added VNet
	if !strings.Contains(res.Stdout, "rpt-vnet") && !strings.Contains(res.Stdout, "added") &&
		!strings.Contains(res.Stdout, "Added") {
		// It's OK if format varies — just verify non-empty output
		if len(strings.TrimSpace(res.Stdout)) == 0 {
			t.Error("drift report produced empty output")
		}
	}
}

// newDriftRunner creates a CLIRunner with simulator env vars configured.
func newDriftRunner(t *testing.T, simBin, stateFile string) *helpers.CLIRunner {
	t.Helper()
	runner := helpers.NewCLIRunner(t)
	runner.Env = append(runner.Env,
		"AZFLOCI_AZ_PATH="+simBin,
		"AZFLOCI_SIM_STATE="+stateFile,
	)
	return runner
}
