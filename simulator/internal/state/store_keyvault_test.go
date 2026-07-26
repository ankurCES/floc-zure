package state

import (
	"testing"
)

func TestKeyVault_CRUD(t *testing.T) {
	s := tempStore(t)

	// Create
	kv, err := s.CreateKeyVault("myvault", "rg1", "eastus", "standard", map[string]string{"env": "dev"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if kv.Name != "myvault" {
		t.Errorf("name: %s", kv.Name)
	}
	if kv.Location != "eastus" {
		t.Errorf("location: %s", kv.Location)
	}
	if kv.Properties.SKU.Name != "standard" {
		t.Errorf("sku: %s", kv.Properties.SKU.Name)
	}
	if kv.Properties.VaultURI != "https://myvault.vault.azure.net/" {
		t.Errorf("vaultUri: %s", kv.Properties.VaultURI)
	}
	if kv.Properties.ProvisioningState != "Succeeded" {
		t.Errorf("state: %s", kv.Properties.ProvisioningState)
	}
	if kv.Tags["env"] != "dev" {
		t.Errorf("tags: %v", kv.Tags)
	}

	// Get
	got := s.GetKeyVault("myvault")
	if got == nil {
		t.Fatal("nil")
	}

	// List
	all := s.ListKeyVaults("")
	if len(all) != 1 {
		t.Fatalf("list all: %d", len(all))
	}
	filtered := s.ListKeyVaults("rg1")
	if len(filtered) != 1 {
		t.Fatalf("list rg1: %d", len(filtered))
	}
	empty := s.ListKeyVaults("other-rg")
	if len(empty) != 0 {
		t.Fatalf("list other: %d", len(empty))
	}

	// Delete
	if err := s.DeleteKeyVault("myvault"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if s.GetKeyVault("myvault") != nil {
		t.Error("should be deleted")
	}
}

func TestKeyVault_Duplicate(t *testing.T) {
	s := tempStore(t)
	_, err := s.CreateKeyVault("v1", "rg1", "eastus", "standard", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateKeyVault("v1", "rg1", "eastus", "standard", nil)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestKeyVault_DeleteNotFound(t *testing.T) {
	s := tempStore(t)
	if err := s.DeleteKeyVault("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSecret_CRUD(t *testing.T) {
	s := tempStore(t)
	s.CreateKeyVault("v1", "rg1", "eastus", "standard", nil)

	// Set
	sec, err := s.SetSecret("v1", "dbpw", "s3cr3t!", "text/plain", map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if sec.Name != "dbpw" {
		t.Errorf("name: %s", sec.Name)
	}
	if sec.Value != "s3cr3t!" {
		t.Errorf("value: %s", sec.Value)
	}
	if sec.VaultName != "v1" {
		t.Errorf("vault: %s", sec.VaultName)
	}
	if !sec.Enabled {
		t.Error("should be enabled")
	}
	if sec.ContentType != "text/plain" {
		t.Errorf("contentType: %s", sec.ContentType)
	}
	if sec.Tags["env"] != "prod" {
		t.Errorf("tags: %v", sec.Tags)
	}
	if sec.Version == "" {
		t.Error("version should be set")
	}

	// Get
	got := s.GetSecret("v1", "dbpw")
	if got == nil {
		t.Fatal("nil")
	}

	// List
	all := s.ListSecrets("v1")
	if len(all) != 1 {
		t.Fatalf("list: %d", len(all))
	}

	// Update (set again)
	sec2, err := s.SetSecret("v1", "dbpw", "new-value", "", nil)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if sec2.Value != "new-value" {
		t.Errorf("updated value: %s", sec2.Value)
	}
	// Version should differ
	if sec2.Version == sec.Version {
		t.Error("version should change on update")
	}

	// Delete
	if err := s.DeleteSecret("v1", "dbpw"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if s.GetSecret("v1", "dbpw") != nil {
		t.Error("should be deleted")
	}
}

func TestSecret_NoVault(t *testing.T) {
	s := tempStore(t)
	_, err := s.SetSecret("nope", "s1", "v", "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSecret_DeleteNotFound(t *testing.T) {
	s := tempStore(t)
	if err := s.DeleteSecret("nope", "s1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestKey_CRUD(t *testing.T) {
	s := tempStore(t)
	s.CreateKeyVault("v1", "rg1", "eastus", "standard", nil)

	// Create
	key, err := s.CreateKey("v1", "mykey", "RSA", 4096, map[string]string{"use": "signing"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if key.Name != "mykey" {
		t.Errorf("name: %s", key.Name)
	}
	if key.KeyType != "RSA" {
		t.Errorf("kty: %s", key.KeyType)
	}
	if key.KeySize != 4096 {
		t.Errorf("size: %d", key.KeySize)
	}
	if !key.Enabled {
		t.Error("should be enabled")
	}
	if len(key.KeyOps) == 0 {
		t.Error("key_ops should be populated")
	}
	if key.Tags["use"] != "signing" {
		t.Errorf("tags: %v", key.Tags)
	}

	// Get
	got := s.GetKey("v1", "mykey")
	if got == nil {
		t.Fatal("nil")
	}

	// List
	all := s.ListKeys("v1")
	if len(all) != 1 {
		t.Fatalf("list: %d", len(all))
	}

	// Delete
	if err := s.DeleteKey("v1", "mykey"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if s.GetKey("v1", "mykey") != nil {
		t.Error("should be deleted")
	}
}

func TestKey_NoVault(t *testing.T) {
	s := tempStore(t)
	_, err := s.CreateKey("nope", "k1", "RSA", 2048, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestKey_DeleteNotFound(t *testing.T) {
	s := tempStore(t)
	if err := s.DeleteKey("nope", "k1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteKeyVault_CascadesSecretsAndKeys(t *testing.T) {
	s := tempStore(t)
	s.CreateKeyVault("v1", "rg1", "eastus", "standard", nil)
	s.SetSecret("v1", "s1", "val", "", nil)
	s.CreateKey("v1", "k1", "RSA", 2048, nil)

	if err := s.DeleteKeyVault("v1"); err != nil {
		t.Fatal(err)
	}
	if len(s.ListSecrets("v1")) != 0 {
		t.Error("secrets should be cascade-deleted")
	}
	if len(s.ListKeys("v1")) != 0 {
		t.Error("keys should be cascade-deleted")
	}
}

func TestGetSecret_NotFound(t *testing.T) {
	s := tempStore(t)
	if s.GetSecret("nope", "s1") != nil {
		t.Error("expected nil")
	}
}

func TestGetKey_NotFound(t *testing.T) {
	s := tempStore(t)
	if s.GetKey("nope", "k1") != nil {
		t.Error("expected nil")
	}
}

func TestGetStorageAccount_NotFound(t *testing.T) {
	s := tempStore(t)
	if s.GetStorageAccount("nope") != nil {
		t.Error("expected nil")
	}
}

func TestGetContainer_NotFound(t *testing.T) {
	s := tempStore(t)
	if s.GetContainer("nope", "c1") != nil {
		t.Error("expected nil")
	}
}

func TestGetBlob_NotFound(t *testing.T) {
	s := tempStore(t)
	if s.GetBlob("nope", "c1", "b1") != nil {
		t.Error("expected nil")
	}
}
