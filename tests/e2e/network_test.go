//go:build e2e

// Tests for networking simulator commands: VNet, Subnet, NSG, NSG Rule, Public IP.
// Uses the az-simulator binary via AZFLOCI_AZ_PATH — no real Azure subscription needed.
package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ankurCES/floc-zure/tests/e2e/helpers"
)

// buildSimulator compiles the az-simulator binary and returns its path.
func buildSimulator(t *testing.T) string {
	t.Helper()
	repoRoot := findRepoRoot2(t)
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "az-simulator")
	cmd := exec.Command("go", "build", "-o", binPath, "./simulator/cmd/az")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build az-simulator: %v\n%s", err, string(out))
	}
	return binPath
}

// findRepoRoot2 walks up to find go.mod.
func findRepoRoot2(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get wd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find repo root")
		}
		dir = parent
	}
}

// simRunner returns a CLIRunner wired to the az-simulator.
func simRunner(t *testing.T) (*helpers.CLIRunner, string) {
	t.Helper()
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	runner := helpers.NewCLIRunner(t)
	runner.Env = append(runner.Env,
		"AZFLOCI_AZ_PATH="+simBin,
		"AZFLOCI_SIM_STATE="+stateFile,
	)
	return runner, stateFile
}

// runSim executes a raw az-simulator command and returns stdout.
func runSim(t *testing.T, simBin, stateFile string, args ...string) string {
	t.Helper()
	cmd := exec.Command(simBin, args...)
	cmd.Env = append(os.Environ(), "AZFLOCI_SIM_STATE="+stateFile)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("az-sim %v failed: %s", args, string(ee.Stderr))
		}
		t.Fatalf("az-sim %v failed: %v", args, err)
	}
	return string(out)
}

// --- VNet tests ---

func TestSimVNetCreateShowListDelete(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	// Create a resource group first
	runSim(t, simBin, stateFile, "group", "create", "-n", "net-rg", "-l", "eastus")

	// Create VNet
	out := runSim(t, simBin, stateFile, "network", "vnet", "create",
		"-n", "my-vnet", "-g", "net-rg", "-l", "eastus",
		"--address-prefixes", "10.0.0.0/16")
	var vnet map[string]interface{}
	if err := json.Unmarshal([]byte(out), &vnet); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if vnet["name"] != "my-vnet" {
		t.Errorf("expected name my-vnet, got %v", vnet["name"])
	}

	// Show
	out = runSim(t, simBin, stateFile, "network", "vnet", "show",
		"-n", "my-vnet", "-g", "net-rg")
	if err := json.Unmarshal([]byte(out), &vnet); err != nil {
		t.Fatalf("show: invalid JSON: %v", err)
	}
	if vnet["name"] != "my-vnet" {
		t.Errorf("show: expected my-vnet, got %v", vnet["name"])
	}

	// List
	out = runSim(t, simBin, stateFile, "network", "vnet", "list", "-g", "net-rg")
	var vnets []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &vnets); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(vnets) != 1 {
		t.Errorf("expected 1 VNet, got %d", len(vnets))
	}

	// Delete
	runSim(t, simBin, stateFile, "network", "vnet", "delete",
		"-n", "my-vnet", "-g", "net-rg")

	// Verify gone
	out = runSim(t, simBin, stateFile, "network", "vnet", "list", "-g", "net-rg")
	if err := json.Unmarshal([]byte(out), &vnets); err != nil {
		t.Fatalf("list after delete: invalid JSON: %v", err)
	}
	if len(vnets) != 0 {
		t.Errorf("expected 0 VNets after delete, got %d", len(vnets))
	}
}

// --- Subnet tests ---

func TestSimSubnetCreateShowListDelete(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "sub-rg", "-l", "eastus")
	runSim(t, simBin, stateFile, "network", "vnet", "create",
		"-n", "sub-vnet", "-g", "sub-rg", "-l", "eastus",
		"--address-prefixes", "10.0.0.0/16")

	// Create subnet
	out := runSim(t, simBin, stateFile, "network", "vnet", "subnet", "create",
		"-n", "sub1", "--vnet-name", "sub-vnet", "-g", "sub-rg",
		"--address-prefixes", "10.0.1.0/24")
	var subnet map[string]interface{}
	if err := json.Unmarshal([]byte(out), &subnet); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if subnet["name"] != "sub1" {
		t.Errorf("expected sub1, got %v", subnet["name"])
	}

	// Show
	out = runSim(t, simBin, stateFile, "network", "vnet", "subnet", "show",
		"-n", "sub1", "--vnet-name", "sub-vnet", "-g", "sub-rg")
	if err := json.Unmarshal([]byte(out), &subnet); err != nil {
		t.Fatalf("show: invalid JSON: %v", err)
	}

	// List
	out = runSim(t, simBin, stateFile, "network", "vnet", "subnet", "list",
		"--vnet-name", "sub-vnet", "-g", "sub-rg")
	var subnets []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &subnets); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(subnets) != 1 {
		t.Errorf("expected 1 subnet, got %d", len(subnets))
	}

	// Delete
	runSim(t, simBin, stateFile, "network", "vnet", "subnet", "delete",
		"-n", "sub1", "--vnet-name", "sub-vnet", "-g", "sub-rg")

	out = runSim(t, simBin, stateFile, "network", "vnet", "subnet", "list",
		"--vnet-name", "sub-vnet", "-g", "sub-rg")
	if err := json.Unmarshal([]byte(out), &subnets); err != nil {
		t.Fatalf("list after delete: invalid JSON: %v", err)
	}
	if len(subnets) != 0 {
		t.Errorf("expected 0 subnets after delete, got %d", len(subnets))
	}
}

// --- NSG tests ---

func TestSimNSGCreateShowListDelete(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "nsg-rg", "-l", "westus")

	// Create NSG
	out := runSim(t, simBin, stateFile, "network", "nsg", "create",
		"-n", "my-nsg", "-g", "nsg-rg", "-l", "westus")
	var nsg map[string]interface{}
	if err := json.Unmarshal([]byte(out), &nsg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if nsg["name"] != "my-nsg" {
		t.Errorf("expected my-nsg, got %v", nsg["name"])
	}

	// Show
	runSim(t, simBin, stateFile, "network", "nsg", "show", "-n", "my-nsg", "-g", "nsg-rg")

	// List
	out = runSim(t, simBin, stateFile, "network", "nsg", "list", "-g", "nsg-rg")
	var nsgs []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &nsgs); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(nsgs) != 1 {
		t.Errorf("expected 1 NSG, got %d", len(nsgs))
	}

	// Delete
	runSim(t, simBin, stateFile, "network", "nsg", "delete", "-n", "my-nsg", "-g", "nsg-rg")

	out = runSim(t, simBin, stateFile, "network", "nsg", "list", "-g", "nsg-rg")
	if err := json.Unmarshal([]byte(out), &nsgs); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(nsgs) != 0 {
		t.Errorf("expected 0 NSGs after delete, got %d", len(nsgs))
	}
}

// --- NSG Rule tests ---

func TestSimNSGRuleCreateDelete(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "rule-rg", "-l", "eastus")
	runSim(t, simBin, stateFile, "network", "nsg", "create",
		"-n", "rule-nsg", "-g", "rule-rg", "-l", "eastus")

	// Create rule
	out := runSim(t, simBin, stateFile, "network", "nsg", "rule", "create",
		"-n", "allow-http", "--nsg-name", "rule-nsg", "-g", "rule-rg",
		"--priority", "100", "--access", "Allow", "--protocol", "Tcp",
		"--direction", "Inbound",
		"--source-address-prefixes", "*",
		"--destination-port-ranges", "80")
	var rule map[string]interface{}
	if err := json.Unmarshal([]byte(out), &rule); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if rule["name"] != "allow-http" {
		t.Errorf("expected allow-http, got %v", rule["name"])
	}

	// Delete rule
	runSim(t, simBin, stateFile, "network", "nsg", "rule", "delete",
		"-n", "allow-http", "--nsg-name", "rule-nsg", "-g", "rule-rg")
}

// --- Public IP tests ---

func TestSimPublicIPCreateShowListDelete(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "pip-rg", "-l", "eastus")

	// Create
	out := runSim(t, simBin, stateFile, "network", "public-ip", "create",
		"-n", "my-pip", "-g", "pip-rg", "-l", "eastus")
	var pip map[string]interface{}
	if err := json.Unmarshal([]byte(out), &pip); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if pip["name"] != "my-pip" {
		t.Errorf("expected my-pip, got %v", pip["name"])
	}

	// Show
	runSim(t, simBin, stateFile, "network", "public-ip", "show",
		"-n", "my-pip", "-g", "pip-rg")

	// List
	out = runSim(t, simBin, stateFile, "network", "public-ip", "list", "-g", "pip-rg")
	var pips []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &pips); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(pips) != 1 {
		t.Errorf("expected 1 PIP, got %d", len(pips))
	}

	// Delete
	runSim(t, simBin, stateFile, "network", "public-ip", "delete",
		"-n", "my-pip", "-g", "pip-rg")
}
