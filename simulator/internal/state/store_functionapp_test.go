package state

import (
	"path/filepath"
	"testing"
)

func faStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestCreateFunctionApp(t *testing.T) {
	s := faStore(t)
	fa, err := s.CreateFunctionApp("fa1", "rg1", "eastus", "node", "18", nil)
	if err != nil {
		t.Fatal(err)
	}
	if fa.Name != "fa1" || fa.Runtime != "node" {
		t.Errorf("got %+v", fa)
	}
	if _, err := s.CreateFunctionApp("fa1", "rg1", "eastus", "", "", nil); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestFunctionAppCRUD(t *testing.T) {
	s := faStore(t)
	s.CreateFunctionApp("fa1", "rg1", "eastus", "python", "3.11", nil)
	if got := s.GetFunctionApp("fa1"); got == nil {
		t.Fatal("expected fa1")
	}
	if got := s.ListFunctionApps(""); len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if err := s.DeleteFunctionApp("fa1"); err != nil {
		t.Fatal(err)
	}
	if got := s.GetFunctionApp("fa1"); got != nil {
		t.Fatal("expected nil")
	}
}

func TestFunctionCRUD(t *testing.T) {
	s := faStore(t)
	s.CreateFunctionApp("fa1", "rg1", "eastus", "node", "18", nil)
	f, err := s.CreateFunction("fa1", "hello", "httpTrigger", "node", nil)
	if err != nil {
		t.Fatal(err)
	}
	if f.TriggerType != "httpTrigger" {
		t.Errorf("trigger: %s", f.TriggerType)
	}
	if got := s.GetFunction("fa1", "hello"); got == nil {
		t.Fatal("expected function")
	}
	if got := s.ListFunctions("fa1"); len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if err := s.DeleteFunction("fa1", "hello"); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionInvoke(t *testing.T) {
	s := faStore(t)
	s.CreateFunctionApp("fa1", "rg1", "eastus", "node", "18", nil)
	s.CreateFunction("fa1", "hello", "httpTrigger", "", nil)
	inv, err := s.InvokeFunction("fa1", "hello", `{"key":"val"}`)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Status != "Succeeded" {
		t.Errorf("status: %s", inv.Status)
	}
	if got := s.ListInvocations("fa1"); len(got) != 1 {
		t.Fatalf("expected 1 invocation, got %d", len(got))
	}
}

func TestFunctionAppDeleteCascade(t *testing.T) {
	s := faStore(t)
	s.CreateFunctionApp("fa1", "rg1", "eastus", "node", "18", nil)
	s.CreateFunction("fa1", "f1", "", "", nil)
	s.InvokeFunction("fa1", "f1", "{}")
	if err := s.DeleteFunctionApp("fa1"); err != nil {
		t.Fatal(err)
	}
	if s.GetFunction("fa1", "f1") != nil {
		t.Error("function should be deleted")
	}
	if len(s.ListInvocations("fa1")) != 0 {
		t.Error("invocations should be deleted")
	}
}
