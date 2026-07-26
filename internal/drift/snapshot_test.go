package drift

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureFromFile(t *testing.T) {
	state := map[string]interface{}{
		"active_subscription": "sub1",
		"subscriptions": []interface{}{},
		"resource_groups": map[string]interface{}{
			"rg1": map[string]interface{}{"id": "/sub/rg1", "name": "rg1", "location": "eastus"},
		},
		"vnets": map[string]interface{}{
			"vnet1": map[string]interface{}{"id": "/sub/vnet1", "name": "vnet1", "location": "eastus"},
		},
		"vms": map[string]interface{}{
			"vm1": map[string]interface{}{"id": "/sub/vm1", "name": "vm1", "location": "westus"},
		},
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")
	data, _ := json.Marshal(state)
	os.WriteFile(p, data, 0644)
	snap, err := CaptureFromFile(p, "test")
	if err != nil { t.Fatal(err) }
	if len(snap.Resources) != 3 { t.Fatalf("expected 3, got %d", len(snap.Resources)) }
}

func TestSaveAndLoad(t *testing.T) {
	snap := &Snapshot{Label: "v1", Resources: []ResourceEntry{{ID: "/a", Name: "a", Type: "t"}}}
	dir := t.TempDir()
	p := filepath.Join(dir, "snap.json")
	if err := SaveSnapshot(snap, p); err != nil { t.Fatal(err) }
	loaded, err := LoadSnapshot(p)
	if err != nil { t.Fatal(err) }
	if len(loaded.Resources) != 1 || loaded.Label != "v1" { t.Errorf("mismatch: %+v", loaded) }
}

func TestCompare_NoDrift(t *testing.T) {
	s := &Snapshot{Label: "a", Resources: []ResourceEntry{{ID: "/a", Name: "a", Type: "t", Location: "east"}}}
	r := Compare(s, s)
	if r.Summary.Total != 0 { t.Errorf("expected 0, got %d", r.Summary.Total) }
}

func TestCompare_Added(t *testing.T) {
	before := &Snapshot{Label: "before", Resources: []ResourceEntry{}}
	after := &Snapshot{Label: "after", Resources: []ResourceEntry{{ID: "/a", Name: "a", Type: "t"}}}
	r := Compare(before, after)
	if r.Summary.Added != 1 { t.Errorf("added: %d", r.Summary.Added) }
}

func TestCompare_Removed(t *testing.T) {
	before := &Snapshot{Label: "before", Resources: []ResourceEntry{{ID: "/a", Name: "a", Type: "t"}}}
	after := &Snapshot{Label: "after", Resources: []ResourceEntry{}}
	r := Compare(before, after)
	if r.Summary.Removed != 1 { t.Errorf("removed: %d", r.Summary.Removed) }
}

func TestCompare_Modified(t *testing.T) {
	before := &Snapshot{Label: "b", Resources: []ResourceEntry{{ID: "/a", Name: "a", Type: "t", Location: "east"}}}
	after := &Snapshot{Label: "a", Resources: []ResourceEntry{{ID: "/a", Name: "a", Type: "t", Location: "west"}}}
	r := Compare(before, after)
	if r.Summary.Modified != 1 { t.Errorf("modified: %d", r.Summary.Modified) }
}

func TestFormatText_NoDrift(t *testing.T) {
	r := &DriftReport{Summary: Summary{}}
	out := FormatText(r)
	if out != "No drift detected.\n" { t.Errorf("got: %s", out) }
}
