package handlers

import (
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterKeyVaultHandlers wires keyvault/secret/key commands into the router.
func RegisterKeyVaultHandlers(r *router.Router, store *state.Store) {
	// az keyvault ...
	r.Register("keyvault create", kvCreate(store))
	r.Register("keyvault show", kvShow(store))
	r.Register("keyvault list", kvList(store))
	r.Register("keyvault delete", kvDelete(store))

	// az keyvault secret ...
	r.Register("keyvault secret set", secretSet(store))
	r.Register("keyvault secret show", secretShow(store))
	r.Register("keyvault secret list", secretList(store))
	r.Register("keyvault secret delete", secretDelete(store))

	// az keyvault key ...
	r.Register("keyvault key create", keyCreate(store))
	r.Register("keyvault key show", keyShow(store))
	r.Register("keyvault key list", keyList(store))
	r.Register("keyvault key delete", keyDelete(store))
}

// ---------------------------------------------------------------------------
// Key Vault handlers
// ---------------------------------------------------------------------------

func kvCreate(store *state.Store) router.HandlerFunc {
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
		if sku == "" {
			sku = "standard"
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		kv, err := store.CreateKeyVault(name, rg, location, sku, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(kv)
	}
}

func kvShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		kv := store.GetKeyVault(name)
		if kv == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Vault '%s' not found.\n", name)
			return 1
		}
		return writeJSON(kv)
	}
}

func kvList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListKeyVaults(rg))
	}
}

func kvDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteKeyVault(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Secret handlers
// ---------------------------------------------------------------------------

func secretSet(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		value, ok := router.ParseFlag(args, "--value")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --value is required.")
			return 2
		}
		contentType, _ := router.ParseFlag(args, "--content-type")
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		sec, err := store.SetSecret(vault, name, value, contentType, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(sec)
	}
}

func secretShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		sec := store.GetSecret(vault, name)
		if sec == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Secret '%s' not found in vault '%s'.\n", name, vault)
			return 1
		}
		return writeJSON(sec)
	}
}

func secretList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		return writeJSON(store.ListSecrets(vault))
	}
}

func secretDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		if err := store.DeleteSecret(vault, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Key handlers
// ---------------------------------------------------------------------------

func keyCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		kty, _ := router.ParseFlag(args, "--kty")
		if kty == "" {
			kty = "RSA"
		}
		sizeStr, _ := router.ParseFlag(args, "--size")
		keySize := 2048
		if sizeStr != "" {
			fmt.Sscanf(sizeStr, "%d", &keySize)
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		key, err := store.CreateKey(vault, name, kty, keySize, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(key)
	}
}

func keyShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		key := store.GetKey(vault, name)
		if key == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Key '%s' not found in vault '%s'.\n", name, vault)
			return 1
		}
		return writeJSON(key)
	}
}

func keyList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		return writeJSON(store.ListKeys(vault))
	}
}

func keyDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		vault, ok := router.ParseFlag(args, "--vault-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --vault-name is required.")
			return 2
		}
		if err := store.DeleteKey(vault, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}
