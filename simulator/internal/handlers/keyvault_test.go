package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ankurCES/floc-zure/simulator/internal/router"
	"github.com/ankurCES/floc-zure/simulator/internal/state"
)

func setupKVRouter(t *testing.T) (*router.Router, *state.Store) {
	t.Helper()
	s, err := state.NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	r := router.New()
	RegisterKeyVaultHandlers(r, s)
	return r, s
}

func TestKeyVaultLifecycle(t *testing.T) {
	r, _ := setupKVRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	if c := r.Dispatch([]string{"keyvault", "create", "--name", "v1", "--resource-group", "rg1"}); c != 0 {
		t.Fatalf("create: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "show", "--name", "v1"}); c != 0 {
		t.Fatalf("show: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "list"}); c != 0 {
		t.Fatalf("list: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "delete", "--name", "v1"}); c != 0 {
		t.Fatalf("delete: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "show", "--name", "v1"}); c == 0 {
		t.Fatal("show after delete should fail")
	}
}

func TestSecretLifecycle(t *testing.T) {
	r, _ := setupKVRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	r.Dispatch([]string{"keyvault", "create", "--name", "v1", "--resource-group", "rg1"})
	if c := r.Dispatch([]string{"keyvault", "secret", "set", "--vault-name", "v1", "--name", "s1", "--value", "secret"}); c != 0 {
		t.Fatalf("set: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "secret", "show", "--vault-name", "v1", "--name", "s1"}); c != 0 {
		t.Fatalf("show: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "secret", "list", "--vault-name", "v1"}); c != 0 {
		t.Fatalf("list: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "secret", "delete", "--vault-name", "v1", "--name", "s1"}); c != 0 {
		t.Fatalf("delete: %d", c)
	}
}

func TestKeyLifecycle(t *testing.T) {
	r, _ := setupKVRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	r.Dispatch([]string{"keyvault", "create", "--name", "v1", "--resource-group", "rg1"})
	if c := r.Dispatch([]string{"keyvault", "key", "create", "--vault-name", "v1", "--name", "k1", "--kty", "RSA", "--size", "4096"}); c != 0 {
		t.Fatalf("create: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "key", "show", "--vault-name", "v1", "--name", "k1"}); c != 0 {
		t.Fatalf("show: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "key", "list", "--vault-name", "v1"}); c != 0 {
		t.Fatalf("list: %d", c)
	}
	if c := r.Dispatch([]string{"keyvault", "key", "delete", "--vault-name", "v1", "--name", "k1"}); c != 0 {
		t.Fatalf("delete: %d", c)
	}
}

func TestKVHandlers_MissingArgs(t *testing.T) {
	r, _ := setupKVRouter(t)
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	cases := []struct {
		name string
		args []string
	}{
		{"kv create no name", []string{"keyvault", "create", "--resource-group", "rg1"}},
		{"kv create no rg", []string{"keyvault", "create", "--name", "v1"}},
		{"kv show no name", []string{"keyvault", "show"}},
		{"kv delete no name", []string{"keyvault", "delete"}},
		{"secret set no name", []string{"keyvault", "secret", "set", "--vault-name", "v1", "--value", "x"}},
		{"secret set no vault", []string{"keyvault", "secret", "set", "--name", "s1", "--value", "x"}},
		{"secret set no value", []string{"keyvault", "secret", "set", "--vault-name", "v1", "--name", "s1"}},
		{"secret show no name", []string{"keyvault", "secret", "show", "--vault-name", "v1"}},
		{"secret show no vault", []string{"keyvault", "secret", "show", "--name", "s1"}},
		{"secret list no vault", []string{"keyvault", "secret", "list"}},
		{"secret delete no name", []string{"keyvault", "secret", "delete", "--vault-name", "v1"}},
		{"secret delete no vault", []string{"keyvault", "secret", "delete", "--name", "s1"}},
		{"key create no name", []string{"keyvault", "key", "create", "--vault-name", "v1"}},
		{"key create no vault", []string{"keyvault", "key", "create", "--name", "k1"}},
		{"key show no name", []string{"keyvault", "key", "show", "--vault-name", "v1"}},
		{"key show no vault", []string{"keyvault", "key", "show", "--name", "k1"}},
		{"key list no vault", []string{"keyvault", "key", "list"}},
		{"key delete no name", []string{"keyvault", "key", "delete", "--vault-name", "v1"}},
		{"key delete no vault", []string{"keyvault", "key", "delete", "--name", "k1"}},
	}
	for _, tc := range cases {
		if c := r.Dispatch(tc.args); c == 0 {
			t.Errorf("%s: expected non-zero exit", tc.name)
		}
	}
}
