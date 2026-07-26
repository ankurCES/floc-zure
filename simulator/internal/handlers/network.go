package handlers

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterNetworkHandlers wires network vnet/subnet/nsg/public-ip commands.
func RegisterNetworkHandlers(r *router.Router, store *state.Store) {
	// az network vnet ...
	r.Register("network vnet create", vnetCreate(store))
	r.Register("network vnet show", vnetShow(store))
	r.Register("network vnet list", vnetList(store))
	r.Register("network vnet delete", vnetDelete(store))

	// az network vnet subnet ...
	r.Register("network vnet subnet create", subnetCreate(store))  // 4 words — router tries 3 max
	r.Register("network subnet create", subnetCreate(store))       // fallback 3-word
	r.Register("network vnet subnet show", subnetShow(store))
	r.Register("network subnet show", subnetShow(store))
	r.Register("network vnet subnet list", subnetList(store))
	r.Register("network subnet list", subnetList(store))
	r.Register("network vnet subnet delete", subnetDelete(store))
	r.Register("network subnet delete", subnetDelete(store))

	// az network nsg ...
	r.Register("network nsg create", nsgCreate(store))
	r.Register("network nsg show", nsgShow(store))
	r.Register("network nsg list", nsgList(store))
	r.Register("network nsg delete", nsgDelete(store))

	// az network nsg rule ...
	r.Register("network nsg rule create", nsgRuleCreate(store))  // 4 words
	r.Register("network rule create", nsgRuleCreate(store))      // fallback
	r.Register("network nsg rule delete", nsgRuleDelete(store))
	r.Register("network rule delete", nsgRuleDelete(store))

	// az network public-ip ...
	r.Register("network public-ip create", publicIPCreate(store))
	r.Register("network public-ip show", publicIPShow(store))
	r.Register("network public-ip list", publicIPList(store))
	r.Register("network public-ip delete", publicIPDelete(store))
}

// ---------------------------------------------------------------------------
// VNet handlers
// ---------------------------------------------------------------------------

func vnetCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		if rg == "" {
			fmt.Fprintln(os.Stderr, "ERROR: --resource-group is required.")
			return 2
		}
		location, _ := router.ParseFlag(args, "--location")
		if location == "" {
			location, _ = router.ParseFlag(args, "-l")
		}
		if location == "" {
			location = "eastus"
		}
		prefixes := []string{"10.0.0.0/16"}
		if ap, ok := router.ParseFlag(args, "--address-prefixes"); ok {
			prefixes = strings.Fields(ap)
		}
		if ap, ok := router.ParseFlag(args, "--address-prefix"); ok {
			prefixes = []string{ap}
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		vnet, err := store.CreateVNet(name, rg, location, prefixes, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(vnet)
	}
}

func vnetShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vnet := store.GetVNet(name)
		if vnet == nil {
			fmt.Fprintf(os.Stderr, "ERROR: VNet '%s' not found.\n", name)
			return 1
		}
		return writeJSON(vnet)
	}
}

func vnetList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListVNets(rg))
	}
}

func vnetDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteVNet(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Subnet handlers
// ---------------------------------------------------------------------------

func subnetCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vnet, ok := router.ParseFlag(args, "--vnet-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vnet-name is required.")
			return 2
		}
		prefix, _ := router.ParseFlag(args, "--address-prefix")
		if prefix == "" {
			prefix, _ = router.ParseFlag(args, "--address-prefixes")
		}
		if prefix == "" {
			prefix = "10.0.0.0/24"
		}
		subnet, err := store.CreateSubnet(vnet, name, prefix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(subnet)
	}
}

func subnetShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vnet, ok := router.ParseFlag(args, "--vnet-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vnet-name is required.")
			return 2
		}
		subnet := store.GetSubnet(vnet, name)
		if subnet == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Subnet '%s' not found in vnet '%s'.\n", name, vnet)
			return 1
		}
		return writeJSON(subnet)
	}
}

func subnetList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		vnet, ok := router.ParseFlag(args, "--vnet-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vnet-name is required.")
			return 2
		}
		return writeJSON(store.ListSubnets(vnet))
	}
}

func subnetDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vnet, ok := router.ParseFlag(args, "--vnet-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vnet-name is required.")
			return 2
		}
		if err := store.DeleteSubnet(vnet, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// NSG handlers
// ---------------------------------------------------------------------------

func nsgCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		if rg == "" {
			fmt.Fprintln(os.Stderr, "ERROR: --resource-group is required.")
			return 2
		}
		location, _ := router.ParseFlag(args, "--location")
		if location == "" {
			location, _ = router.ParseFlag(args, "-l")
		}
		if location == "" {
			location = "eastus"
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		nsg, err := store.CreateNSG(name, rg, location, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(nsg)
	}
}

func nsgShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsg := store.GetNSG(name)
		if nsg == nil {
			fmt.Fprintf(os.Stderr, "ERROR: NSG '%s' not found.\n", name)
			return 1
		}
		return writeJSON(nsg)
	}
}

func nsgList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListNSGs(rg))
	}
}

func nsgDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteNSG(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// NSG Rule handlers
// ---------------------------------------------------------------------------

func nsgRuleCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsgName, ok := router.ParseFlag(args, "--nsg-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --nsg-name is required.")
			return 2
		}
		priority := 100
		if ps, ok := router.ParseFlag(args, "--priority"); ok {
			if p, err := strconv.Atoi(ps); err == nil {
				priority = p
			}
		}
		direction, _ := router.ParseFlag(args, "--direction")
		if direction == "" {
			direction = "Inbound"
		}
		access, _ := router.ParseFlag(args, "--access")
		if access == "" {
			access = "Allow"
		}
		protocol, _ := router.ParseFlag(args, "--protocol")
		if protocol == "" {
			protocol = "*"
		}
		srcAddr, _ := router.ParseFlag(args, "--source-address-prefix")
		if srcAddr == "" {
			srcAddr = "*"
		}
		srcPort, _ := router.ParseFlag(args, "--source-port-range")
		if srcPort == "" {
			srcPort = "*"
		}
		dstAddr, _ := router.ParseFlag(args, "--destination-address-prefix")
		if dstAddr == "" {
			dstAddr = "*"
		}
		dstPort, _ := router.ParseFlag(args, "--destination-port-range")
		if dstPort == "" {
			dstPort = "*"
		}

		rule := state.NSGRule{
			Name:                name,
			Priority:            priority,
			Direction:           direction,
			Access:              access,
			Protocol:            protocol,
			SourceAddressPrefix: srcAddr,
			SourcePortRange:     srcPort,
			DestAddressPrefix:   dstAddr,
			DestPortRange:       dstPort,
		}
		created, err := store.CreateNSGRule(nsgName, rule)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(created)
	}
}

func nsgRuleDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		nsgName, ok := router.ParseFlag(args, "--nsg-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --nsg-name is required.")
			return 2
		}
		if err := store.DeleteNSGRule(nsgName, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Public IP handlers
// ---------------------------------------------------------------------------

func publicIPCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		if rg == "" {
			fmt.Fprintln(os.Stderr, "ERROR: --resource-group is required.")
			return 2
		}
		location, _ := router.ParseFlag(args, "--location")
		if location == "" {
			location, _ = router.ParseFlag(args, "-l")
		}
		if location == "" {
			location = "eastus"
		}
		sku, _ := router.ParseFlag(args, "--sku")
		allocation, _ := router.ParseFlag(args, "--allocation-method")
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		pip, err := store.CreatePublicIP(name, rg, location, sku, allocation, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(pip)
	}
}

func publicIPShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		pip := store.GetPublicIP(name)
		if pip == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Public IP '%s' not found.\n", name)
			return 1
		}
		return writeJSON(pip)
	}
}

func publicIPList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListPublicIPs(rg))
	}
}

func publicIPDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeletePublicIP(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}
