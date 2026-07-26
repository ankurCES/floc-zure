package handlers

import (
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterStorageHandlers wires storage account/container/blob commands into the router.
func RegisterStorageHandlers(r *router.Router, store *state.Store) {
	// az storage account ...
	r.Register("storage account create", storageAccountCreate(store))
	r.Register("storage account show", storageAccountShow(store))
	r.Register("storage account list", storageAccountList(store))
	r.Register("storage account delete", storageAccountDelete(store))

	// az storage container ...
	r.Register("storage container create", containerCreate(store))
	r.Register("storage container show", containerShow(store))
	r.Register("storage container list", containerList(store))
	r.Register("storage container delete", containerDelete(store))

	// az storage blob ...
	r.Register("storage blob upload", blobUpload(store))
	r.Register("storage blob show", blobShow(store))
	r.Register("storage blob list", blobList(store))
	r.Register("storage blob delete", blobDelete(store))
}

// ---------------------------------------------------------------------------
// Storage Account handlers
// ---------------------------------------------------------------------------

func storageAccountCreate(store *state.Store) router.HandlerFunc {
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
		kind, _ := router.ParseFlag(args, "--kind")
		if kind == "" {
			kind = "StorageV2"
		}
		sku, _ := router.ParseFlag(args, "--sku")
		if sku == "" {
			sku = "Standard_LRS"
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			tags = nil
		}
		sa, err := store.CreateStorageAccount(name, rg, location, kind, sku, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(sa)
	}
}

func storageAccountShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		sa := store.GetStorageAccount(name)
		if sa == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Storage account '%s' not found.\n", name)
			return 1
		}
		return writeJSON(sa)
	}
}

func storageAccountList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		return writeJSON(store.ListStorageAccounts(rg))
	}
}

func storageAccountDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		if err := store.DeleteStorageAccount(name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Container handlers
// ---------------------------------------------------------------------------

func containerCreate(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		c, err := store.CreateContainer(acct, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(c)
	}
}

func containerShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		c := store.GetContainer(acct, name)
		if c == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Container '%s' not found in account '%s'.\n", name, acct)
			return 1
		}
		return writeJSON(c)
	}
}

func containerList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		return writeJSON(store.ListContainers(acct))
	}
}

func containerDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		if err := store.DeleteContainer(acct, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

// ---------------------------------------------------------------------------
// Blob handlers
// ---------------------------------------------------------------------------

func blobUpload(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		container, ok := router.ParseFlag(args, "--container-name")
		if !ok {
			container, ok = router.ParseFlag(args, "-c")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --container-name is required.")
			return 2
		}
		filePath, _ := router.ParseFlag(args, "--file")
		if filePath == "" {
			filePath, _ = router.ParseFlag(args, "-f")
		}
		contentType, _ := router.ParseFlag(args, "--content-type")

		// Get file size if file exists
		var size int64
		if filePath != "" {
			info, err := os.Stat(filePath)
			if err == nil {
				size = info.Size()
			}
		}

		b, err := store.CreateBlob(acct, container, name, contentType, size, filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(b)
	}
}

func blobShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		container, ok := router.ParseFlag(args, "--container-name")
		if !ok {
			container, ok = router.ParseFlag(args, "-c")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --container-name is required.")
			return 2
		}
		b := store.GetBlob(acct, container, name)
		if b == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Blob '%s' not found.\n", name)
			return 1
		}
		return writeJSON(b)
	}
}

func blobList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		container, ok := router.ParseFlag(args, "--container-name")
		if !ok {
			container, ok = router.ParseFlag(args, "-c")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --container-name is required.")
			return 2
		}
		return writeJSON(store.ListBlobs(acct, container))
	}
}

func blobDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		name, ok := router.ParseFlag(args, "--name")
		if !ok {
			name, ok = router.ParseFlag(args, "-n")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --name is required.")
			return 2
		}
		acct, ok := router.ParseFlag(args, "--account-name")
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --account-name is required.")
			return 2
		}
		container, ok := router.ParseFlag(args, "--container-name")
		if !ok {
			container, ok = router.ParseFlag(args, "-c")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --container-name is required.")
			return 2
		}
		if err := store.DeleteBlob(acct, container, name); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}
