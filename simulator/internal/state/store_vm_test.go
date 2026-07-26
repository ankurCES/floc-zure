package state

import "testing"

func TestVM_CRUD(t *testing.T) {
	s := tempStore(t)
	vm, err := s.CreateVM("vm1", "rg1", "eastus", "Standard_B2s", "Canonical:UbuntuServer:18.04-LTS:latest", "admin", nil)
	if err != nil {
		t.Fatal(err)
	}
	if vm.Name != "vm1" || vm.PowerState != VMStateRunning {
		t.Errorf("got %s/%s", vm.Name, vm.PowerState)
	}
	if s.GetVM("vm1") == nil {
		t.Fatal("nil")
	}
	if len(s.ListVMs("")) != 1 {
		t.Fatal("list")
	}
	if err := s.DeleteVM("vm1"); err != nil {
		t.Fatal(err)
	}
}

func TestVM_StateMachine(t *testing.T) {
	s := tempStore(t)
	s.CreateVM("vm1", "rg1", "eastus", "", "", "", nil)

	// Running -> Stopped
	vm, err := s.TransitionVM("vm1", VMStateStopped)
	if err != nil {
		t.Fatal(err)
	}
	if vm.PowerState != VMStateStopped {
		t.Errorf("expected Stopped, got %s", vm.PowerState)
	}
	// Stopped -> Running
	vm, err = s.TransitionVM("vm1", VMStateRunning)
	if err != nil {
		t.Fatal(err)
	}
	if vm.PowerState != VMStateRunning {
		t.Errorf("expected Running, got %s", vm.PowerState)
	}
	// Running -> Deallocated
	vm, err = s.TransitionVM("vm1", VMStateDeallocated)
	if err != nil {
		t.Fatal(err)
	}
	if vm.PowerState != VMStateDeallocated {
		t.Errorf("expected Deallocated, got %s", vm.PowerState)
	}
	// Deallocated -> Running
	_, err = s.TransitionVM("vm1", VMStateRunning)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVM_InvalidTransition(t *testing.T) {
	s := tempStore(t)
	s.CreateVM("vm1", "rg1", "eastus", "", "", "", nil)
	// Running -> Creating is invalid
	_, err := s.TransitionVM("vm1", VMStateCreating)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
}
