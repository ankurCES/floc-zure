package handlers

import (
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterFunctionAppHandlers wires functionapp commands into the router.
func RegisterFunctionAppHandlers(r *router.Router, store *state.Store) {
	// az functionapp ...
	r.Register("functionapp create", faCreate(store))
	r.Register("functionapp show", faShow(store))
	r.Register("functionapp list", faList(store))
	r.Register("functionapp delete", faDelete(store))

	// az functionapp function ...
	r.Register("functionapp function create", funcCreate(store))
	r.Register("functionapp function show", funcShow(store))
	r.Register("functionapp function list", funcList(store))
	r.Register("functionapp function delete", funcDelete(store))

	// invoke + invocation history
	r.Register("functionapp function invoke", funcInvoke(store))
	r.Register("functionapp function invocations", funcInvocations(store))
}

// ---------------------------------------------------------------------------
// Function App CRUD
// ---------------------------------------------------------------------------

func faCreate(store *state.Store) router.HandlerFunc {
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
		runtime, _ := router.ParseFlag(args, "--runtime")
		runtimeVer, _ := router.ParseFlag(args, "--runtime-version")
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		fa, err := store.CreateFunctionApp(name, rg, location, runtime, runtimeVer, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(fa)
	}
}

func faShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		fa := store.GetFunctionApp(name)
		if fa == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Function app '%s' not found.\n", name)
			return 1
		}
		return writeJSON(fa)
	}
}

func faList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListFunctionApps(rg))
	}
}

func faDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteFunctionApp(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Function CRUD
// ---------------------------------------------------------------------------

func funcCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		appName, ok := router.ParseFlag(args, "--function-app-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --function-app-name is required.")
			return 2
		}
		triggerType, _ := router.ParseFlag(args, "--trigger-type")
		language, _ := router.ParseFlag(args, "--language")
		f, err := store.CreateFunction(appName, name, triggerType, language, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(f)
	}
}

func funcShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		appName, ok := router.ParseFlag(args, "--function-app-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --function-app-name is required.")
			return 2
		}
		f := store.GetFunction(appName, name)
		if f == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Function '%s' not found in app '%s'.\n", name, appName)
			return 1
		}
		return writeJSON(f)
	}
}

func funcList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		appName, ok := router.ParseFlag(args, "--function-app-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --function-app-name is required.")
			return 2
		}
		return writeJSON(store.ListFunctions(appName))
	}
}

func funcDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		appName, ok := router.ParseFlag(args, "--function-app-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --function-app-name is required.")
			return 2
		}
		if err := store.DeleteFunction(appName, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Invoke + History
// ---------------------------------------------------------------------------

func funcInvoke(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		appName, ok := router.ParseFlag(args, "--function-app-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --function-app-name is required.")
			return 2
		}
		input, _ := router.ParseFlag(args, "--input")
		if input == "" {
			input = "{}"
		}
		inv, err := store.InvokeFunction(appName, name, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(inv)
	}
}

func funcInvocations(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		appName, ok := router.ParseFlag(args, "--function-app-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --function-app-name is required.")
			return 2
		}
		return writeJSON(store.ListInvocations(appName))
	}
}
