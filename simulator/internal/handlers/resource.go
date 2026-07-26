package handlers

import (
	"fmt"
	"os"
	"strings"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

// RegisterResourceHandlers wires resource commands into the router.
func RegisterResourceHandlers(r *router.Router, store *state.Store) {
	r.Register("resource list", resourceList(store))
	r.Register("resource show", resourceShow(store))
	r.Register("resource delete", resourceDelete(store))
	r.Register("resource tag", resourceTag(store))
}

func resourceList(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		rg, _ := router.ParseFlag(args, "--resource-group")
		if rg == "" {
			rg, _ = router.ParseFlag(args, "-g")
		}
		resources := store.ListResources(rg)
		return writeJSON(resources)
	}
}

func resourceShow(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		id, ok := router.ParseFlag(args, "--ids")
		if !ok {
			id, ok = router.ParseFlag(args, "--id")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --ids is required.")
			return 2
		}
		res := store.GetResource(id)
		if res == nil {
			fmt.Fprintf(os.Stderr, "ERROR: Resource '%s' could not be found.\n", id)
			return 1
		}
		return writeJSON(res)
	}
}

func resourceDelete(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		id, ok := router.ParseFlag(args, "--ids")
		if !ok {
			id, ok = router.ParseFlag(args, "--id")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --ids is required.")
			return 2
		}
		if err := store.DeleteResource(id); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	}
}

func resourceTag(store *state.Store) router.HandlerFunc {
	return func(args []string) int {
		id, ok := router.ParseFlag(args, "--ids")
		if !ok {
			id, ok = router.ParseFlag(args, "--id")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "ERROR: --ids is required.")
			return 2
		}
		tags := router.ParseTags(args)
		if len(tags) == 0 {
			// Also try inline --tags key=val (single pair without ParseTags finding it)
			tagsRaw, hasRaw := router.ParseFlag(args, "--tags")
			if hasRaw {
				tags = parseInlineTags(tagsRaw)
			}
		}
		if len(tags) == 0 {
			fmt.Fprintln(os.Stderr, "ERROR: --tags is required.")
			return 2
		}
		res, err := store.TagResource(id, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return writeJSON(res)
	}
}

// parseInlineTags handles "--tags k1=v1 k2=v2" passed as a single string.
func parseInlineTags(raw string) map[string]string {
	tags := make(map[string]string)
	for _, pair := range strings.Fields(raw) {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			tags[parts[0]] = parts[1]
		}
	}
	return tags
}
