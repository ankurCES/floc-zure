// Command az is a drop-in replacement for the Azure CLI that operates against
// a local JSON-backed state file. It enables fully offline testing of azfloci.
//
// Usage:
//
//	AZFLOCI_AZ_PATH=$(which az-simulator) azfloci auth status
//
// Environment:
//
//	AZFLOCI_SIM_STATE  — path to the state JSON file (default ~/.azfloci-sim/state.json)
package main

import (
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/simulator/internal/handlers"
	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func main() {
	store, err := state.NewStore(state.DefaultStatePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to load simulator state: %v\n", err)
		os.Exit(1)
	}

	r := router.New()
	handlers.RegisterAccountHandlers(r, store)
	handlers.RegisterGroupHandlers(r, store)
	handlers.RegisterResourceHandlers(r, store)
	handlers.RegisterStorageHandlers(r, store)
	handlers.RegisterKeyVaultHandlers(r, store)

	// os.Args[0] is the binary name; pass the rest as the command line.
	exitCode := r.Dispatch(os.Args[1:])
	os.Exit(exitCode)
}
