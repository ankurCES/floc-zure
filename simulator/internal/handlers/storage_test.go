package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func setupStorageRouter(t *testing.T) (*router.Router, *state.Store) {
	t.Helper()
	s, err := state.NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	r := router.New()
	RegisterStorageHandlers(r, s)
	RegisterAccountHandlers(r, s)
	RegisterGroupHandlers(r, s)
	return r, s
}

func TestStorageAccountLifecycle(t *testing.T) {
	r, _ := setupStorageRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	if c := r.Dispatch([]string{"storage", "account", "create", "--name", "sa1", "--resource-group", "rg1", "-l", "westus2"}); c != 0 {
		t.Fatalf("create: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "account", "show", "--name", "sa1"}); c != 0 {
		t.Fatalf("show: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "account", "list"}); c != 0 {
		t.Fatalf("list: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "account", "delete", "--name", "sa1"}); c != 0 {
		t.Fatalf("delete: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "account", "show", "--name", "sa1"}); c == 0 {
		t.Fatal("show after delete should fail")
	}
}

func TestContainerLifecycle(t *testing.T) {
	r, _ := setupStorageRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	r.Dispatch([]string{"storage", "account", "create", "--name", "sa1", "--resource-group", "rg1"})
	if c := r.Dispatch([]string{"storage", "container", "create", "--name", "c1", "--account-name", "sa1"}); c != 0 {
		t.Fatalf("create: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "container", "show", "--name", "c1", "--account-name", "sa1"}); c != 0 {
		t.Fatalf("show: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "container", "list", "--account-name", "sa1"}); c != 0 {
		t.Fatalf("list: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "container", "delete", "--name", "c1", "--account-name", "sa1"}); c != 0 {
		t.Fatalf("delete: %d", c)
	}
}

func TestBlobLifecycle(t *testing.T) {
	r, _ := setupStorageRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	r.Dispatch([]string{"storage", "account", "create", "--name", "sa1", "--resource-group", "rg1"})
	r.Dispatch([]string{"storage", "container", "create", "--name", "c1", "--account-name", "sa1"})

	if c := r.Dispatch([]string{"storage", "blob", "upload", "--name", "b1.txt", "--account-name", "sa1", "--container-name", "c1"}); c != 0 {
		t.Fatalf("upload: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "blob", "show", "--name", "b1.txt", "--account-name", "sa1", "--container-name", "c1"}); c != 0 {
		t.Fatalf("show: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "blob", "list", "--account-name", "sa1", "--container-name", "c1"}); c != 0 {
		t.Fatalf("list: %d", c)
	}
	if c := r.Dispatch([]string{"storage", "blob", "delete", "--name", "b1.txt", "--account-name", "sa1", "--container-name", "c1"}); c != 0 {
		t.Fatalf("delete: %d", c)
	}
}

func TestStorageHandlers_MissingArgs(t *testing.T) {
	r, _ := setupStorageRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	cases := []struct {
		name string
		args []string
	}{
		{"sa create no name", []string{"storage", "account", "create", "--resource-group", "rg1"}},
		{"sa create no rg", []string{"storage", "account", "create", "--name", "sa1"}},
		{"sa show no name", []string{"storage", "account", "show"}},
		{"sa delete no name", []string{"storage", "account", "delete"}},
		{"container create no name", []string{"storage", "container", "create", "--account-name", "sa1"}},
		{"container create no acct", []string{"storage", "container", "create", "--name", "c1"}},
		{"container show no name", []string{"storage", "container", "show", "--account-name", "sa1"}},
		{"container show no acct", []string{"storage", "container", "show", "--name", "c1"}},
		{"container list no acct", []string{"storage", "container", "list"}},
		{"container delete no name", []string{"storage", "container", "delete", "--account-name", "sa1"}},
		{"blob upload no name", []string{"storage", "blob", "upload", "--account-name", "sa1", "--container-name", "c1"}},
		{"blob upload no acct", []string{"storage", "blob", "upload", "--name", "b1", "--container-name", "c1"}},
		{"blob upload no container", []string{"storage", "blob", "upload", "--name", "b1", "--account-name", "sa1"}},
		{"blob show no name", []string{"storage", "blob", "show", "--account-name", "sa1", "--container-name", "c1"}},
		{"blob list no acct", []string{"storage", "blob", "list", "--container-name", "c1"}},
		{"blob list no container", []string{"storage", "blob", "list", "--account-name", "sa1"}},
		{"blob delete no name", []string{"storage", "blob", "delete", "--account-name", "sa1", "--container-name", "c1"}},
	}
	for _, tc := range cases {
		if c := r.Dispatch(tc.args); c == 0 {
			t.Errorf("%s: expected non-zero exit", tc.name)
		}
	}
}
