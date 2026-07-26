package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func tempStore(t *testing.T) *state.Store {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")
	s, err := state.NewStore(p)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

// captureStdout runs fn() and returns what it wrote to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// --- Account tests ---

func TestAccountShow(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterAccountHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"account", "show"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})

	var acct state.Subscription
	if err := json.Unmarshal([]byte(out), &acct); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, out)
	}
	if acct.Name != "Simulated-Subscription-1" {
		t.Errorf("name: %s", acct.Name)
	}
	if acct.User.Name != "simulator@azfloci.local" {
		t.Errorf("user: %s", acct.User.Name)
	}
}

func TestAccountList(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterAccountHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"account", "list"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})

	var subs []state.Subscription
	if err := json.Unmarshal([]byte(out), &subs); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}
	if len(subs) != 1 {
		t.Errorf("expected 1 sub, got %d", len(subs))
	}
}

func TestAccountSet(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterAccountHandlers(rtr, store)

	// Set to the existing sub (should succeed silently)
	code := rtr.Dispatch([]string{"account", "set", "--subscription", "00000000-0000-0000-0000-000000000001"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
}

func TestAccountSet_Missing(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterAccountHandlers(rtr, store)

	code := rtr.Dispatch([]string{"account", "set", "--subscription", "nonexistent"})
	if code == 0 {
		t.Fatal("expected non-zero exit for nonexistent sub")
	}
}

// --- Group tests ---

func TestGroupCreate(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterGroupHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"group", "create", "--name", "rg1", "--location", "westus2", "--tags", "env=test"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})

	var rg state.ResourceGroup
	if err := json.Unmarshal([]byte(out), &rg); err != nil {
		t.Fatalf("parse: %v\nout: %s", err, out)
	}
	if rg.Name != "rg1" {
		t.Errorf("name: %s", rg.Name)
	}
	if rg.Location != "westus2" {
		t.Errorf("location: %s", rg.Location)
	}
}

func TestGroupShow(t *testing.T) {
	store := tempStore(t)
	store.CreateResourceGroup("rg1", "eastus", nil)
	rtr := router.New()
	RegisterGroupHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"group", "show", "--name", "rg1"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	if !strings.Contains(out, "rg1") {
		t.Errorf("expected rg1 in output: %s", out)
	}
}

func TestGroupShow_NotFound(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterGroupHandlers(rtr, store)

	code := rtr.Dispatch([]string{"group", "show", "--name", "nope"})
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
}

func TestGroupList(t *testing.T) {
	store := tempStore(t)
	store.CreateResourceGroup("rg1", "eastus", nil)
	store.CreateResourceGroup("rg2", "westus", nil)
	rtr := router.New()
	RegisterGroupHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"group", "list"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})

	var groups []state.ResourceGroup
	if err := json.Unmarshal([]byte(out), &groups); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2, got %d", len(groups))
	}
}

func TestGroupDelete(t *testing.T) {
	store := tempStore(t)
	store.CreateResourceGroup("rg1", "eastus", nil)
	rtr := router.New()
	RegisterGroupHandlers(rtr, store)

	code := rtr.Dispatch([]string{"group", "delete", "--name", "rg1", "--yes"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if store.GetResourceGroup("rg1") != nil {
		t.Error("rg should be deleted")
	}
}

func TestGroupDelete_NotFound(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterGroupHandlers(rtr, store)

	code := rtr.Dispatch([]string{"group", "delete", "--name", "nope", "--yes"})
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
}

// --- Resource tests ---

func TestResourceList(t *testing.T) {
	store := tempStore(t)
	store.CreateResourceGroup("rg1", "eastus", nil)
	sub := store.GetActiveSubscription()
	resID := state.GenerateResourceID(sub.ID, "rg1", "Microsoft.Storage", "storageAccounts", "sa1")
	store.AddResource(&state.Resource{
		ID:       resID,
		Name:     "sa1",
		Type:     "Microsoft.Storage/storageAccounts",
		Location: "eastus",
	})

	rtr := router.New()
	RegisterResourceHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"resource", "list", "--resource-group", "rg1"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})

	var resources []state.Resource
	if err := json.Unmarshal([]byte(out), &resources); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("expected 1, got %d", len(resources))
	}
}

func TestResourceShow(t *testing.T) {
	store := tempStore(t)
	sub := store.GetActiveSubscription()
	resID := state.GenerateResourceID(sub.ID, "rg1", "Microsoft.Compute", "virtualMachines", "vm1")
	store.AddResource(&state.Resource{ID: resID, Name: "vm1", Type: "Microsoft.Compute/virtualMachines", Location: "eastus"})

	rtr := router.New()
	RegisterResourceHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"resource", "show", "--ids", resID})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})
	if !strings.Contains(out, "vm1") {
		t.Errorf("expected vm1: %s", out)
	}
}

func TestResourceShow_NotFound(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterResourceHandlers(rtr, store)

	code := rtr.Dispatch([]string{"resource", "show", "--ids", "/nonexistent"})
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
}

func TestResourceDelete(t *testing.T) {
	store := tempStore(t)
	sub := store.GetActiveSubscription()
	resID := state.GenerateResourceID(sub.ID, "rg1", "Microsoft.Storage", "storageAccounts", "sa1")
	store.AddResource(&state.Resource{ID: resID, Name: "sa1"})

	rtr := router.New()
	RegisterResourceHandlers(rtr, store)

	code := rtr.Dispatch([]string{"resource", "delete", "--ids", resID, "--yes"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if store.GetResource(resID) != nil {
		t.Error("should be deleted")
	}
}

func TestResourceTag(t *testing.T) {
	store := tempStore(t)
	sub := store.GetActiveSubscription()
	resID := state.GenerateResourceID(sub.ID, "rg1", "Microsoft.Storage", "storageAccounts", "sa1")
	store.AddResource(&state.Resource{ID: resID, Name: "sa1"})

	rtr := router.New()
	RegisterResourceHandlers(rtr, store)

	out := captureStdout(t, func() {
		code := rtr.Dispatch([]string{"resource", "tag", "--ids", resID, "--tags", "env=prod", "team=dev"})
		if code != 0 {
			t.Fatalf("exit %d", code)
		}
	})

	var res state.Resource
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if res.Tags["env"] != "prod" {
		t.Errorf("env tag: %v", res.Tags)
	}
	if res.Tags["team"] != "dev" {
		t.Errorf("team tag: %v", res.Tags)
	}
}

func TestResourceTag_NotFound(t *testing.T) {
	store := tempStore(t)
	rtr := router.New()
	RegisterResourceHandlers(rtr, store)

	code := rtr.Dispatch([]string{"resource", "tag", "--ids", "/nonexistent", "--tags", "k=v"})
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
}
