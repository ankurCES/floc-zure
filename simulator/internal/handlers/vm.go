package handlers

import (
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterVMHandlers wires VM lifecycle commands into the router.
func RegisterVMHandlers(r *router.Router, store *state.Store) {
	r.Register("vm create", vmCreate(store))
	r.Register("vm show", vmShow(store))
	r.Register("vm list", vmList(store))
	r.Register("vm delete", vmDelete(store))
	r.Register("vm start", vmStart(store))
	r.Register("vm stop", vmStop(store))
	r.Register("vm restart", vmRestart(store))
	r.Register("vm deallocate", vmDeallocate(store))
}

func vmCreate(store *state.Store) router.HandlerFunc {
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
		size, _ := router.ParseFlag(args, "--size")
		image, _ := router.ParseFlag(args, "--image")
		adminUser, _ := router.ParseFlag(args, "--admin-username")
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		vm, err := store.CreateVM(name, rg, location, size, image, adminUser, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(vm)
	}
}

func vmShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vm := store.GetVM(name)
		if vm == nil {
			fmt.Fprintf(os.Stderr, "ERROR: VM '%s' not found.\n", name)
			return 1
		}
		return writeJSON(vm)
	}
}

func vmList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListVMs(rg))
	}
}

func vmDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteVM(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

func vmStart(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vm, err := store.TransitionVM(name, state.VMStateRunning)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(vm)
	}
}

func vmStop(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vm, err := store.TransitionVM(name, state.VMStateStopped)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(vm)
	}
}

func vmRestart(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		// restart = stop then start (only valid from Running)
		_, err := store.TransitionVM(name, state.VMStateStopped)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		vm, err := store.TransitionVM(name, state.VMStateRunning)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(vm)
	}
}

func vmDeallocate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vm, err := store.TransitionVM(name, state.VMStateDeallocated)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(vm)
	}
}
