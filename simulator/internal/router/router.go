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

	// Try longest prefix first (5 words, then 4, then 3, then 2, then 1).
	if len(cleaned) >= 5 {
		key := cleaned[0] + " " + cleaned[1] + " " + cleaned[2] + " " + cleaned[3] + " " + cleaned[4]
		if fn, ok := r.routes[key]; ok {
			return fn(cleaned[5:])
		}
	}
	if len(cleaned) >= 4 {
		key := cleaned[0] + " " + cleaned[1] + " " + cleaned[2] + " " + cleaned[3]
		if fn, ok := r.routes[key]; ok {
			return fn(cleaned[4:])
		}
	}
	if len(cleaned) >= 3 {
		key := cleaned[0] + " " + cleaned[1] + " " + cleaned[2]
		if fn, ok := r.routes[key]; ok {
			return fn(cleaned[3:])
		}
	}
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
  account show               Show active subscription
  account set                Switch subscription
  account list               List all subscriptions

  group create               Create a resource group
  group show                 Show a resource group
  group list                 List resource groups
  group delete               Delete a resource group

  resource list              List resources
  resource show              Show a resource
  resource delete            Delete a resource
  resource tag               Tag a resource

  storage account create     Create a storage account
  storage account show       Show a storage account
  storage account list       List storage accounts
  storage account delete     Delete a storage account
  storage container create   Create a blob container
  storage container show     Show a blob container
  storage container list     List blob containers
  storage container delete   Delete a blob container
  storage blob upload        Upload a blob
  storage blob show          Show blob properties
  storage blob list          List blobs
  storage blob delete        Delete a blob

  keyvault create            Create a key vault
  keyvault show              Show a key vault
  keyvault list              List key vaults
  keyvault delete            Delete a key vault
  keyvault secret set        Set a secret
  keyvault secret show       Show a secret
  keyvault secret list       List secrets
  keyvault secret delete     Delete a secret
  keyvault key create        Create a key
  keyvault key show          Show a key
  keyvault key list          List keys
  keyvault key delete        Delete a key

  network vnet create        Create a virtual network
  network vnet show          Show a virtual network
  network vnet list          List virtual networks
  network vnet delete        Delete a virtual network
  network vnet subnet create Create a subnet
  network vnet subnet show   Show a subnet
  network vnet subnet list   List subnets
  network vnet subnet delete Delete a subnet
  network nsg create         Create a network security group
  network nsg show           Show a network security group
  network nsg list           List network security groups
  network nsg delete         Delete a network security group
  network nsg rule create    Create an NSG rule
  network nsg rule delete    Delete an NSG rule
  network public-ip create   Create a public IP address
  network public-ip show     Show a public IP address
  network public-ip list     List public IP addresses
  network public-ip delete   Delete a public IP address

  vm create                  Create a virtual machine
  vm show                    Show a virtual machine
  vm list                    List virtual machines
  vm delete                  Delete a virtual machine
  vm start                   Start a virtual machine
  vm stop                    Stop a virtual machine
  vm restart                 Restart a virtual machine
  vm deallocate              Deallocate a virtual machine

  servicebus namespace create       Create a Service Bus namespace
  servicebus namespace show         Show a namespace
  servicebus namespace list         List namespaces
  servicebus namespace delete       Delete a namespace
  servicebus queue create           Create a queue
  servicebus queue show             Show a queue
  servicebus queue list             List queues
  servicebus queue delete           Delete a queue
  servicebus topic create           Create a topic
  servicebus topic show             Show a topic
  servicebus topic list             List topics
  servicebus topic delete           Delete a topic
  servicebus topic subscription create  Create a topic subscription
  servicebus topic subscription show    Show a subscription
  servicebus topic subscription list    List subscriptions
  servicebus topic subscription delete  Delete a subscription
  servicebus queue message send     Send a message to a queue
  servicebus queue message receive  Receive (dequeue) a message
  servicebus queue message peek     Peek at next message

  functionapp create                Create a function app
  functionapp show                  Show a function app
  functionapp list                  List function apps
  functionapp delete                Delete a function app
  functionapp function create       Create a function
  functionapp function show         Show a function
  functionapp function list         List functions in an app
  functionapp function delete       Delete a function
  functionapp function invoke       Invoke a function
  functionapp function invocations  List invocation history`)
}
