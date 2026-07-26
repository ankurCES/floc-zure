package state

import (
	"testing"
)

func TestStorageAccount_CRUD(t *testing.T) {
	s := tempStore(t)

	// Create
	sa, err := s.CreateStorageAccount("mysa", "rg1", "westus2", "StorageV2", "Standard_LRS", map[string]string{"env": "test"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if sa.Name != "mysa" {
		t.Errorf("name: %s", sa.Name)
	}
	if sa.Location != "westus2" {
		t.Errorf("location: %s", sa.Location)
	}
	if sa.Kind != "StorageV2" {
		t.Errorf("kind: %s", sa.Kind)
	}
	if sa.SKU.Name != "Standard_LRS" {
		t.Errorf("sku: %s", sa.SKU.Name)
	}
	if sa.SKU.Tier != "Standard" {
		t.Errorf("tier: %s", sa.SKU.Tier)
	}
	if sa.Tags["env"] != "test" {
		t.Errorf("tags: %v", sa.Tags)
	}
	if sa.ProvisioningState != "Succeeded" {
		t.Errorf("state: %s", sa.ProvisioningState)
	}
	if sa.PrimaryEndpoints.Blob != "https://mysa.blob.core.windows.net/" {
		t.Errorf("blob endpoint: %s", sa.PrimaryEndpoints.Blob)
	}

	// Get
	got := s.GetStorageAccount("mysa")
	if got == nil {
		t.Fatal("GetStorageAccount returned nil")
	}
	if got.Name != "mysa" {
		t.Errorf("get name: %s", got.Name)
	}

	// List
	all := s.ListStorageAccounts("")
	if len(all) != 1 {
		t.Fatalf("list all: %d", len(all))
	}
	filtered := s.ListStorageAccounts("rg1")
	if len(filtered) != 1 {
		t.Fatalf("list rg1: %d", len(filtered))
	}
	empty := s.ListStorageAccounts("other-rg")
	if len(empty) != 0 {
		t.Fatalf("list other-rg: %d", len(empty))
	}

	// Delete
	if err := s.DeleteStorageAccount("mysa"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if s.GetStorageAccount("mysa") != nil {
		t.Error("should be deleted")
	}
}

func TestStorageAccount_Duplicate(t *testing.T) {
	s := tempStore(t)
	_, err := s.CreateStorageAccount("sa1", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateStorageAccount("sa1", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestStorageAccount_DeleteNotFound(t *testing.T) {
	s := tempStore(t)
	if err := s.DeleteStorageAccount("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestStorageAccount_PremiumTier(t *testing.T) {
	s := tempStore(t)
	sa, err := s.CreateStorageAccount("psa", "rg1", "eastus", "StorageV2", "Premium_LRS", nil)
	if err != nil {
		t.Fatal(err)
	}
	if sa.SKU.Tier != "Premium" {
		t.Errorf("expected Premium, got %s", sa.SKU.Tier)
	}
}

func TestContainer_CRUD(t *testing.T) {
	s := tempStore(t)
	s.CreateStorageAccount("mysa", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)

	// Create
	c, err := s.CreateContainer("mysa", "mycontainer")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.Name != "mycontainer" {
		t.Errorf("name: %s", c.Name)
	}
	if c.AccountName != "mysa" {
		t.Errorf("account: %s", c.AccountName)
	}
	if c.PublicAccess != "off" {
		t.Errorf("publicAccess: %s", c.PublicAccess)
	}

	// Get
	got := s.GetContainer("mysa", "mycontainer")
	if got == nil {
		t.Fatal("nil")
	}

	// List
	all := s.ListContainers("mysa")
	if len(all) != 1 {
		t.Fatalf("list: %d", len(all))
	}

	// Delete
	if err := s.DeleteContainer("mysa", "mycontainer"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if s.GetContainer("mysa", "mycontainer") != nil {
		t.Error("should be deleted")
	}
}

func TestContainer_NoAccount(t *testing.T) {
	s := tempStore(t)
	_, err := s.CreateContainer("nope", "c1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContainer_Duplicate(t *testing.T) {
	s := tempStore(t)
	s.CreateStorageAccount("sa1", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)
	_, err := s.CreateContainer("sa1", "c1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateContainer("sa1", "c1")
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestContainer_DeleteNotFound(t *testing.T) {
	s := tempStore(t)
	if err := s.DeleteContainer("nope", "c1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestBlob_CRUD(t *testing.T) {
	s := tempStore(t)
	s.CreateStorageAccount("sa1", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)
	s.CreateContainer("sa1", "c1")

	// Create
	b, err := s.CreateBlob("sa1", "c1", "test.txt", "text/plain", 42, "/tmp/test.txt")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if b.Name != "test.txt" {
		t.Errorf("name: %s", b.Name)
	}
	if b.ContentType != "text/plain" {
		t.Errorf("content-type: %s", b.ContentType)
	}
	if b.ContentLen != 42 {
		t.Errorf("size: %d", b.ContentLen)
	}
	if b.BlobType != "BlockBlob" {
		t.Errorf("blobType: %s", b.BlobType)
	}

	// Get
	got := s.GetBlob("sa1", "c1", "test.txt")
	if got == nil {
		t.Fatal("nil")
	}

	// List
	all := s.ListBlobs("sa1", "c1")
	if len(all) != 1 {
		t.Fatalf("list: %d", len(all))
	}

	// Delete
	if err := s.DeleteBlob("sa1", "c1", "test.txt"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if s.GetBlob("sa1", "c1", "test.txt") != nil {
		t.Error("should be deleted")
	}
}

func TestBlob_DefaultContentType(t *testing.T) {
	s := tempStore(t)
	s.CreateStorageAccount("sa1", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)
	s.CreateContainer("sa1", "c1")

	b, err := s.CreateBlob("sa1", "c1", "data.bin", "", 100, "")
	if err != nil {
		t.Fatal(err)
	}
	if b.ContentType != "application/octet-stream" {
		t.Errorf("expected default content type, got %s", b.ContentType)
	}
}

func TestBlob_NoContainer(t *testing.T) {
	s := tempStore(t)
	_, err := s.CreateBlob("nope", "c1", "b1", "", 0, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBlob_DeleteNotFound(t *testing.T) {
	s := tempStore(t)
	if err := s.DeleteBlob("nope", "c1", "b1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteStorageAccount_CascadesContainersAndBlobs(t *testing.T) {
	s := tempStore(t)
	s.CreateStorageAccount("sa1", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)
	s.CreateContainer("sa1", "c1")
	s.CreateBlob("sa1", "c1", "b1", "", 0, "")

	if err := s.DeleteStorageAccount("sa1"); err != nil {
		t.Fatal(err)
	}
	if len(s.ListContainers("sa1")) != 0 {
		t.Error("containers should be cascade-deleted")
	}
	if len(s.ListBlobs("sa1", "c1")) != 0 {
		t.Error("blobs should be cascade-deleted")
	}
}

func TestDeleteContainer_CascadesBlobs(t *testing.T) {
	s := tempStore(t)
	s.CreateStorageAccount("sa1", "rg1", "eastus", "StorageV2", "Standard_LRS", nil)
	s.CreateContainer("sa1", "c1")
	s.CreateBlob("sa1", "c1", "b1", "", 0, "")
	s.CreateBlob("sa1", "c1", "b2", "", 0, "")

	if err := s.DeleteContainer("sa1", "c1"); err != nil {
		t.Fatal(err)
	}
	if len(s.ListBlobs("sa1", "c1")) != 0 {
		t.Error("blobs should be cascade-deleted")
	}
}
