// Package handlers implements az-command handlers backed by the state store.
package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterAccountHandlers wires account commands into the router.
func RegisterAccountHandlers(r *router.Router, store *state.Store) {
	r.Register("account show", accountShow(store))
	r.Register("account set", accountSet(store))
	r.Register("account list", accountList(store))
}

func accountShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		sub := store.GetActiveSubscription()
		if sub == nil {
			fmt.Fprintln(os.Stderr, "ERROR: No subscription found. Run 'az login' to log in.")
			return 1
		}
		return writeJSON(sub)
	}
}

func accountSet(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		id, ok := router.ParseFlag(args, "--subscription")
		if !ok {
			// Also try -s shorthand
			id, ok = router.ParseFlag(args, "-s")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --subscription is required.")
			return 2
		}
		if err := store.SetActiveSubscription(id); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

func accountList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		subs := store.ListSubscriptions()
		return writeJSON(subs)
	}
}

// writeJSON marshals v to indented JSON on stdout. Returns exit code.
func writeJSON(v interface{}) int {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: marshal JSON: %v\n", err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}
