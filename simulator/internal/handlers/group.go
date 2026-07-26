package handlers

import (
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterGroupHandlers wires resource group commands into the router.
func RegisterGroupHandlers(r *router.Router, store *state.Store) {
	r.Register("group create", groupCreate(store))
	r.Register("group show", groupShow(store))
	r.Register("group list", groupList(store))
	r.Register("group delete", groupDelete(store))
}

func groupCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		location, ok := router.ParseFlag(args, "--location")
		if !ok {
			location, ok = router.ParseFlag(args, "-l")
		}
		if !ok {
			location = "eastus"
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		rg := store.CreateResourceGroup(name, location, tags)
		return writeJSON(rg)
	}
}

func groupShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		rg := store.GetResourceGroup(name)
		if rg == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Resource group '%s' could not be found.\n", name)
			return 1
		}
		return writeJSON(rg)
	}
}

func groupList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		groups := store.ListResourceGroups()
		return writeJSON(groups)
	}
}

func groupDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteResourceGroup(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		// az group delete with --yes produces no output on success.
		return 0
	}
}
