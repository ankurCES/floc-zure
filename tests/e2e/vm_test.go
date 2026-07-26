//go:build e2e

// Tests for VM simulator commands: create, show, list, delete, start, stop, restart, deallocate.
// Uses the az-simulator binary via AZFLOCI_SIM_STATE — no real Azure subscription needed.
package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSimVMLifecycle(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "vm-rg", "-l", "eastus")

	// Create VM
	out := runSim(t, simBin, stateFile, "vm", "create",
		"-n", "test-vm", "-g", "vm-rg", "-l", "eastus",
		"--image", "Ubuntu2204", "--size", "Standard_B2s")
	var vm map[string]interface{}
	if err := json.Unmarshal([]byte(out), &vm); err != nil {
		t.Fatalf("create: invalid JSON: %v", err)
	}
	if vm["name"] != "test-vm" {
		t.Errorf("expected test-vm, got %v", vm["name"])
	}

	// Show — should be Running
	out = runSim(t, simBin, stateFile, "vm", "show", "-n", "test-vm", "-g", "vm-rg")
	if err := json.Unmarshal([]byte(out), &vm); err != nil {
		t.Fatalf("show: invalid JSON: %v", err)
	}
	if ps, ok := vm["powerState"]; !ok || ps != "Running" {
		t.Errorf("expected powerState 'Running', got %v", vm["powerState"])
	}

	// Stop
	out = runSim(t, simBin, stateFile, "vm", "stop", "-n", "test-vm", "-g", "vm-rg")
	if err := json.Unmarshal([]byte(out), &vm); err != nil {
		t.Fatalf("stop: invalid JSON: %v", err)
	}
	if ps := vm["powerState"]; ps != "Stopped" {
		t.Errorf("expected 'Stopped' after stop, got %v", ps)
	}

	// Start from stopped (valid: Stopped→Running)
	out = runSim(t, simBin, stateFile, "vm", "start", "-n", "test-vm", "-g", "vm-rg")
	if err := json.Unmarshal([]byte(out), &vm); err != nil {
		t.Fatalf("restart: invalid JSON: %v", err)
	}
	if ps := vm["powerState"]; ps != "Running" {
		t.Errorf("expected 'Running' after restart, got %v", ps)
	}

	// Deallocate
	out = runSim(t, simBin, stateFile, "vm", "deallocate", "-n", "test-vm", "-g", "vm-rg")
	if err := json.Unmarshal([]byte(out), &vm); err != nil {
		t.Fatalf("deallocate: invalid JSON: %v", err)
	}
	if ps := vm["powerState"]; ps != "Deallocated" {
		t.Errorf("expected 'Deallocated', got %v", ps)
	}

	// Start from deallocated (valid: Deallocated→Running)
	out = runSim(t, simBin, stateFile, "vm", "start", "-n", "test-vm", "-g", "vm-rg")
	if err := json.Unmarshal([]byte(out), &vm); err != nil {
		t.Fatalf("start: invalid JSON: %v", err)
	}
	if ps := vm["powerState"]; ps != "Running" {
		t.Errorf("expected 'Running' after start from deallocated, got %v", ps)
	}

	// List
	out = runSim(t, simBin, stateFile, "vm", "list", "-g", "vm-rg")
	var vms []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &vms); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(vms) != 1 {
		t.Errorf("expected 1 VM, got %d", len(vms))
	}

	// Delete
	runSim(t, simBin, stateFile, "vm", "delete", "-n", "test-vm", "-g", "vm-rg", "--yes")

	out = runSim(t, simBin, stateFile, "vm", "list", "-g", "vm-rg")
	if err := json.Unmarshal([]byte(out), &vms); err != nil {
		t.Fatalf("list after delete: invalid JSON: %v", err)
	}
	if len(vms) != 0 {
		t.Errorf("expected 0 VMs after delete, got %d", len(vms))
	}
}

func TestSimVMNotFound(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	cmd := exec.Command(simBin, "vm", "show", "-n", "nope", "-g", "nope-rg")
	cmd.Env = append(os.Environ(), "AZFLOCI_SIM_STATE="+stateFile)
	err := cmd.Run()
	if err == nil {
		t.Error("expected error for non-existent VM")
	}
}

func TestSimMultipleVMs(t *testing.T) {
	simBin := buildSimulator(t)
	stateFile := filepath.Join(t.TempDir(), "state.json")

	runSim(t, simBin, stateFile, "group", "create", "-n", "multi-rg", "-l", "westus")

	for _, name := range []string{"vm-a", "vm-b", "vm-c"} {
		runSim(t, simBin, stateFile, "vm", "create",
			"-n", name, "-g", "multi-rg", "-l", "westus",
			"--image", "Ubuntu2204", "--size", "Standard_B1s")
	}

	out := runSim(t, simBin, stateFile, "vm", "list", "-g", "multi-rg")
	var vms []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &vms); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(vms) != 3 {
		t.Errorf("expected 3 VMs, got %d", len(vms))
	}
}
