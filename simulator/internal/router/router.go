// Package router parses az-style CLI arguments and dispatches to the correct handler.
package router

import (
	"fmt"
	"os"
	"strings"
)

// HandlerFunc processes a parsed az command. args contains the remaining
// positional/flag arguments after the command prefix has been consumed.
// It writes JSON to stdout and returns an exit code.
type HandlerFunc func(args []string) int

// Router maps az command prefixes (e.g. "account show") to handlers.
type Router struct {
	routes map[string]HandlerFunc
}

// New creates an empty router.
func New() *Router {
	return &Router{routes: make(map[string]HandlerFunc)}
}

// Register maps a command path to a handler. path is space-separated,
// e.g. "account show", "group create".
func (r *Router) Register(path string, fn HandlerFunc) {
	r.routes[path] = fn
}

// Dispatch parses os.Args-style arguments, finds the longest matching
// command prefix, and calls the handler with remaining args.
// Returns the process exit code.
func (r *Router) Dispatch(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: az: no command specified. Try 'az --help'.")
		return 2
	}

	// Strip global flags that appear before the command: --output, -o, --query
	cleaned := stripGlobalFlags(args)

	// Try longest prefix first (2 words, then 1 word).
	if len(cleaned) >= 2 {
		key := cleaned[0] + " " + cleaned[1]
		if fn, ok := r.routes[key]; ok {
			return fn(cleaned[2:])
		}
	}
	if len(cleaned) >= 1 {
		key := cleaned[0]
		if fn, ok := r.routes[key]; ok {
			return fn(cleaned[1:])
		}
	}

	// Special: --version / version
	for _, a := range cleaned {
		if a == "--version" || a == "version" {
			fmt.Println("azure-cli-simulator 1.0.0 (azfloci)")
			return 0
		}
		if a == "--help" || a == "-h" || a == "help" {
			printHelp()
			return 0
		}
	}

	fmt.Fprintf(os.Stderr, "ERROR: '%s' is not a recognized command.\n", strings.Join(cleaned, " "))
	return 2
}

// ParseFlag extracts the value of a named flag from args.
// Supports both "--name value" and "--name=value" forms.
// Returns ("", false) if not found.
func ParseFlag(args []string, name string) (string, bool) {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1], true
		}
		if strings.HasPrefix(a, name+"=") {
			return a[len(name)+1:], true
		}
	}
	return "", false
}

// ParseFlagBool returns true if the flag is present (no value needed).
func ParseFlagBool(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

// ParseTags extracts tag key=value pairs from args after a --tags flag.
// Collects all non-flag args until the next -- flag or end of args.
func ParseTags(args []string) map[string]string {
	tags := make(map[string]string)
	collecting := false
	for _, a := range args {
		if a == "--tags" {
			collecting = true
			continue
		}
		if collecting {
			if strings.HasPrefix(a, "--") {
				break
			}
			parts := strings.SplitN(a, "=", 2)
			if len(parts) == 2 {
				tags[parts[0]] = parts[1]
			}
		}
	}
	return tags
}

// stripGlobalFlags removes --output/--query flags and their values from args.
func stripGlobalFlags(args []string) []string {
	var out []string
	skip := false
	for i, a := range args {
		if skip {
			skip = false
			continue
		}
		if (a == "--output" || a == "-o" || a == "--query") && i+1 < len(args) {
			skip = true
			continue
		}
		// Also strip --output=json form
		if strings.HasPrefix(a, "--output=") || strings.HasPrefix(a, "-o=") {
			continue
		}
		out = append(out, a)
	}
	return out
}

func printHelp() {
	fmt.Println(`azure-cli-simulator — drop-in az replacement for azfloci testing

Commands:
  account show          Show active subscription
  account set           Switch subscription
  account list          List all subscriptions
  group create          Create a resource group
  group show            Show a resource group
  group list            List resource groups
  group delete          Delete a resource group
  resource list         List resources
  resource show         Show a resource
  resource delete       Delete a resource
  resource tag          Tag a resource`)
}
