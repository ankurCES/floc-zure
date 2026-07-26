package state

import (
	"testing"
)

func TestVNet_CRUD(t *testing.T) {
	s := tempStore(t)
	v, err := s.CreateVNet("vnet1", "rg1", "eastus", nil, map[string]string{"env": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Name != "vnet1" || v.Location != "eastus" {
		t.Errorf("got %s/%s", v.Name, v.Location)
	}
	if len(v.AddressSpace.AddressPrefixes) != 1 || v.AddressSpace.AddressPrefixes[0] != "10.0.0.0/16" {
		t.Errorf("default prefix: %v", v.AddressSpace.AddressPrefixes)
	}
	if s.GetVNet("vnet1") == nil {
		t.Fatal("GetVNet nil")
	}
	list := s.ListVNets("")
	if len(list) != 1 {
		t.Fatalf("list: %d", len(list))
	}
	if err := s.DeleteVNet("vnet1"); err != nil {
		t.Fatal(err)
	}
	if s.GetVNet("vnet1") != nil {
		t.Error("not deleted")
	}
}

func TestVNet_Duplicate(t *testing.T) {
	s := tempStore(t)
	s.CreateVNet("v1", "rg1", "eastus", nil, nil)
	_, err := s.CreateVNet("v1", "rg1", "eastus", nil, nil)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestSubnet_CRUD(t *testing.T) {
	s := tempStore(t)
	s.CreateVNet("vnet1", "rg1", "eastus", nil, nil)
	sub, err := s.CreateSubnet("vnet1", "sub1", "10.0.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	if sub.Name != "sub1" || sub.AddressPrefix != "10.0.1.0/24" {
		t.Errorf("got %s/%s", sub.Name, sub.AddressPrefix)
	}
	if s.GetSubnet("vnet1", "sub1") == nil {
		t.Fatal("GetSubnet nil")
	}
	if len(s.ListSubnets("vnet1")) != 1 {
		t.Fatal("list subnets")
	}
	if err := s.DeleteSubnet("vnet1", "sub1"); err != nil {
		t.Fatal(err)
	}
	if s.GetSubnet("vnet1", "sub1") != nil {
		t.Error("not deleted")
	}
}

func TestSubnet_VNetNotFound(t *testing.T) {
	s := tempStore(t)
	_, err := s.CreateSubnet("nope", "sub1", "10.0.0.0/24")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNSG_CRUD(t *testing.T) {
	s := tempStore(t)
	nsg, err := s.CreateNSG("nsg1", "rg1", "eastus", nil)
	if err != nil {
		t.Fatal(err)
	}
	if nsg.Name != "nsg1" {
		t.Errorf("name: %s", nsg.Name)
	}
	if s.GetNSG("nsg1") == nil {
		t.Fatal("nil")
	}
	if len(s.ListNSGs("")) != 1 {
		t.Fatal("list")
	}
	if err := s.DeleteNSG("nsg1"); err != nil {
		t.Fatal(err)
	}
}

func TestNSGRule_CRUD(t *testing.T) {
	s := tempStore(t)
	s.CreateNSG("nsg1", "rg1", "eastus", nil)
	rule, err := s.CreateNSGRule("nsg1", NSGRule{Name: "allow-ssh", Priority: 100, Direction: "Inbound", Access: "Allow", Protocol: "Tcp", DestPortRange: "22"})
	if err != nil {
		t.Fatal(err)
	}
	if rule.Name != "allow-ssh" || rule.ProvisioningState != "Succeeded" {
		t.Errorf("rule: %+v", rule)
	}
	if err := s.DeleteNSGRule("nsg1", "allow-ssh"); err != nil {
		t.Fatal(err)
	}
	nsg := s.GetNSG("nsg1")
	if len(nsg.SecurityRules) != 0 {
		t.Error("rule not deleted")
	}
}

func TestPublicIP_CRUD(t *testing.T) {
	s := tempStore(t)
	pip, err := s.CreatePublicIP("pip1", "rg1", "eastus", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if pip.Name != "pip1" || pip.IPAddress == "" {
		t.Errorf("pip: %s / %s", pip.Name, pip.IPAddress)
	}
	if s.GetPublicIP("pip1") == nil {
		t.Fatal("nil")
	}
	if len(s.ListPublicIPs("")) != 1 {
		t.Fatal("list")
	}
	if err := s.DeletePublicIP("pip1"); err != nil {
		t.Fatal(err)
	}
}
