package handlers

import (
	"encoding/json"
	"testing"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func TestVMCreate(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterVMHandlers(rtr, store)
	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"vm", "create", "--name", "vm1", "--resource-group", "rg1", "--image", "Canonical:UbuntuServer:18.04-LTS:latest"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	var vm state.VirtualMachine
	if err := json.Unmarshal([]byte(out), &vm); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if vm.Name != "vm1" || vm.PowerState != state.VMStateRunning {
		t.Errorf("got %s/%s", vm.Name, vm.PowerState)
	}
}

func TestVMStopStart(t *testing.T) {
	store := tempStore(t)
	store.CreateVM("vm1", "rg1", "eastus", "", "", "", nil)
	rtr := router.New()
	RegisterVMHandlers(rtr, store)
	code := rtr.Dispatch([]string{"vm", "stop", "--name", "vm1"})
	if code != 0 {
		t.Fatalf("stop exit %d", code)
	}
	if store.GetVM("vm1").PowerState != state.VMStateStopped {
		t.Error("not stopped")
	}
	captureStdout(t, func() {
		code = rtr.Dispatch([]string{"vm", "start", "--name", "vm1"})
	})
	if code != 0 {
		t.Fatalf("start exit %d", code)
	}
	if store.GetVM("vm1").PowerState != state.VMStateRunning {
		t.Error("not running")
	}
}

func TestVMDeallocate(t *testing.T) {
	store := tempStore(t)
	store.CreateVM("vm1", "rg1", "eastus", "", "", "", nil)
	rtr := router.New()
	RegisterVMHandlers(rtr, store)
	captureStdout(t, func() {
		code := rtr.Dispatch([]string{"vm", "deallocate", "--name", "vm1"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	if store.GetVM("vm1").PowerState != state.VMStateDeallocated {
		t.Error("not deallocated")
	}
}

func TestVMRestart(t *testing.T) {
	store := tempStore(t)
	store.CreateVM("vm1", "rg1", "eastus", "", "", "", nil)
	rtr := router.New()
	RegisterVMHandlers(rtr, store)
	captureStdout(t, func() {
		code := rtr.Dispatch([]string{"vm", "restart", "--name", "vm1"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	if store.GetVM("vm1").PowerState != state.VMStateRunning {
		t.Error("not running after restart")
	}
}

func TestVMDelete(t *testing.T) {
	store := tempStore(t)
	store.CreateVM("vm1", "rg1", "eastus", "", "", "", nil)
	rtr := router.New()
	RegisterVMHandlers(rtr, store)
	code := rtr.Dispatch([]string{"vm", "delete", "--name", "vm1"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if store.GetVM("vm1") != nil {
		t.Error("not deleted")
	}
}

func TestVMShowNotFound(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterVMHandlers(rtr, store)
	code := rtr.Dispatch([]string{"vm", "show", "--name", "nope"})
	if code == 0 {
		t.Fatal("expected error")
	}
}
