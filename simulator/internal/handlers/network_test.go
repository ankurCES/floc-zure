package handlers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func TestVNetCreate(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"network", "vnet", "create", "--name", "vnet1", "--resource-group", "rg1", "--location", "westus2"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	var v state.VNet
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v.Name != "vnet1" {
		t.Errorf("name: %s", v.Name)
	}
}

func TestVNetShowNotFound(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	code := rtr.Dispatch([]string{"network", "vnet", "show", "--name", "nope"})
	if code == 0 {
		t.Fatal("expected error")
	}
}

func TestVNetList(t *testing.T) {
	store := tempStore(t)
	store.CreateVNet("v1", "rg1", "eastus", nil, nil)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"network", "vnet", "list"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	if !strings.Contains(out, "v1") {
		t.Error("missing v1")
	}
}

func TestVNetDelete(t *testing.T) {
	store := tempStore(t)
	store.CreateVNet("v1", "rg1", "eastus", nil, nil)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	code := rtr.Dispatch([]string{"network", "vnet", "delete", "--name", "v1"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if store.GetVNet("v1") != nil {
		t.Error("not deleted")
	}
}

func TestSubnetCreate(t *testing.T) {
	store := tempStore(t)
	store.CreateVNet("vnet1", "rg1", "eastus", nil, nil)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"network", "vnet", "subnet", "create", "--name", "sub1", "--vnet-name", "vnet1", "--address-prefix", "10.0.1.0/24"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	var s state.Subnet
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s.Name != "sub1" {
		t.Errorf("name: %s", s.Name)
	}
}

func TestNSGCreate(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"network", "nsg", "create", "--name", "nsg1", "--resource-group", "rg1"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	var n state.NSG
	if err := json.Unmarshal([]byte(out), &n); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if n.Name != "nsg1" {
		t.Errorf("name: %s", n.Name)
	}
}

func TestNSGRuleCreate(t *testing.T) {
	store := tempStore(t)
	store.CreateNSG("nsg1", "rg1", "eastus", nil)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"network", "nsg", "rule", "create", "--name", "allow-ssh", "--nsg-name", "nsg1", "--priority", "100", "--access", "Allow", "--protocol", "Tcp", "--destination-port-range", "22"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	var r state.NSGRule
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if r.Name != "allow-ssh" {
		t.Errorf("name: %s", r.Name)
	}
}

func TestPublicIPCreate(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterNetworkHandlers(rtr, store)
	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"network", "public-ip", "create", "--name", "pip1", "--resource-group", "rg1"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	var p state.PublicIP
	if err := json.Unmarshal([]byte(out), &p); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if p.Name != "pip1" {
		t.Errorf("name: %s", p.Name)
	}
}
