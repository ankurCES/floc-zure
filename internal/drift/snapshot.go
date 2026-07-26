// Package drift captures Azure simulator state snapshots and diffs them.
package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ResourceEntry is a single resource in a snapshot.
type ResourceEntry struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Location   string            `json:"location"`
	Properties json.RawMessage   `json:"properties,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// Snapshot is a point-in-time capture of all resources.
type Snapshot struct {
	Timestamp  time.Time       `json:"timestamp"`
	Label      string          `json:"label,omitempty"`
	Resources  []ResourceEntry `json:"resources"`
}

// DriftChange describes one resource difference.
type DriftChange struct {
	Action   string          `json:"action"` // added, removed, modified
	Type     string          `json:"type"`
	Name     string          `json:"name"`
	ID       string          `json:"id"`
	Details  []FieldDiff     `json:"details,omitempty"`
}

// FieldDiff is a single field-level change.
type FieldDiff struct {
	Field    string `json:"field"`
	Before   string `json:"before"`
	After    string `json:"after"`
}

// DriftReport is the result of comparing two snapshots.
type DriftReport struct {
	Before   string        `json:"before_label"`
	After    string        `json:"after_label"`
	Changes  []DriftChange `json:"changes"`
	Summary  Summary       `json:"summary"`
}

// Summary counts changes by action.
type Summary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
	Total    int `json:"total"`
}

// CaptureFromFile reads a simulator state JSON file and extracts all resources.
func CaptureFromFile(statePath, label string) (*Snapshot, error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	snap := &Snapshot{Timestamp: time.Now().UTC(), Label: label}
	extractMap(raw, "resource_groups", "Microsoft.Resources/resourceGroups", snap)
	extractMap(raw, "storage_accounts", "Microsoft.Storage/storageAccounts", snap)
	extractMap(raw, "key_vaults", "Microsoft.KeyVault/vaults", snap)
	extractMap(raw, "vnets", "Microsoft.Network/virtualNetworks", snap)
	extractMap(raw, "nsgs", "Microsoft.Network/networkSecurityGroups", snap)
	extractMap(raw, "public_ips", "Microsoft.Network/publicIPAddresses", snap)
	extractMap(raw, "vms", "Microsoft.Compute/virtualMachines", snap)
	sort.Slice(snap.Resources, func(i, j int) bool { return snap.Resources[i].ID < snap.Resources[j].ID })
	return snap, nil
}

func extractMap(raw map[string]json.RawMessage, key, resType string, snap *Snapshot) {
	data, ok := raw[key]
	if !ok {
		return
	}
	var items map[string]json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return
	}
	for _, v := range items {
		var entry struct {
			ID       string            `json:"id"`
			Name     string            `json:"name"`
			Location string            `json:"location"`
			Tags     map[string]string `json:"tags"`
		}
		json.Unmarshal(v, &entry)
		snap.Resources = append(snap.Resources, ResourceEntry{
			ID:         entry.ID,
			Name:       entry.Name,
			Type:       resType,
			Location:   entry.Location,
			Properties: v,
			Tags:       entry.Tags,
		})
	}
}

// SaveSnapshot writes a snapshot to a JSON file.
func SaveSnapshot(snap *Snapshot, path string) error {
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadSnapshot reads a snapshot from a JSON file.
func LoadSnapshot(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// Compare diffs two snapshots and returns a report.
func Compare(before, after *Snapshot) *DriftReport {
	beforeMap := indexByID(before.Resources)
	afterMap := indexByID(after.Resources)
	report := &DriftReport{Before: before.Label, After: after.Label}

	// Removed
	for id, b := range beforeMap {
		if _, ok := afterMap[id]; !ok {
			report.Changes = append(report.Changes, DriftChange{Action: "removed", Type: b.Type, Name: b.Name, ID: id})
		}
	}
	// Added
	for id, a := range afterMap {
		if _, ok := beforeMap[id]; !ok {
			report.Changes = append(report.Changes, DriftChange{Action: "added", Type: a.Type, Name: a.Name, ID: id})
		}
	}
	// Modified
	for id, b := range beforeMap {
		a, ok := afterMap[id]
		if !ok {
			continue
		}
		diffs := diffEntry(b, a)
		if len(diffs) > 0 {
			report.Changes = append(report.Changes, DriftChange{Action: "modified", Type: a.Type, Name: a.Name, ID: id, Details: diffs})
		}
	}
	sort.Slice(report.Changes, func(i, j int) bool { return report.Changes[i].ID < report.Changes[j].ID })
	for _, c := range report.Changes {
		switch c.Action {
		case "added":
			report.Summary.Added++
		case "removed":
			report.Summary.Removed++
		case "modified":
			report.Summary.Modified++
		}
	}
	report.Summary.Total = report.Summary.Added + report.Summary.Removed + report.Summary.Modified
	return report
}

func indexByID(entries []ResourceEntry) map[string]ResourceEntry {
	m := make(map[string]ResourceEntry, len(entries))
	for _, e := range entries {
		m[e.ID] = e
	}
	return m
}

func diffEntry(b, a ResourceEntry) []FieldDiff {
	var diffs []FieldDiff
	if b.Location != a.Location {
		diffs = append(diffs, FieldDiff{Field: "location", Before: b.Location, After: a.Location})
	}
	if b.Type != a.Type {
		diffs = append(diffs, FieldDiff{Field: "type", Before: b.Type, After: a.Type})
	}
	if b.Name != a.Name {
		diffs = append(diffs, FieldDiff{Field: "name", Before: b.Name, After: a.Name})
	}
	// Deep compare tags
	allTags := make(map[string]bool)
	for k := range b.Tags { allTags[k] = true }
	for k := range a.Tags { allTags[k] = true }
	for k := range allTags {
		bv, av := b.Tags[k], a.Tags[k]
		if bv != av {
			diffs = append(diffs, FieldDiff{Field: "tags." + k, Before: bv, After: av})
		}
	}
	// Deep compare properties JSON
	if string(b.Properties) != string(a.Properties) {
		propDiffs := diffJSON(b.Properties, a.Properties, "properties")
		diffs = append(diffs, propDiffs...)
	}
	return diffs
}

func diffJSON(b, a json.RawMessage, prefix string) []FieldDiff {
	var bm, am map[string]interface{}
	json.Unmarshal(b, &bm)
	json.Unmarshal(a, &am)
	if bm == nil || am == nil {
		if string(b) != string(a) {
			return []FieldDiff{{Field: prefix, Before: string(b), After: string(a)}}
		}
		return nil
	}
	var diffs []FieldDiff
	allKeys := make(map[string]bool)
	for k := range bm { allKeys[k] = true }
	for k := range am { allKeys[k] = true }
	for k := range allKeys {
		bv := fmt.Sprintf("%v", bm[k])
		av := fmt.Sprintf("%v", am[k])
		if bv != av {
			diffs = append(diffs, FieldDiff{Field: prefix + "." + k, Before: bv, After: av})
		}
	}
	return diffs
}

// FormatText renders a human-readable drift report.
func FormatText(r *DriftReport) string {
	if r.Summary.Total == 0 {
		return "No drift detected.\n"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Drift Report: %s → %s\n", r.Before, r.After))
	sb.WriteString(fmt.Sprintf("  %d added, %d removed, %d modified (%d total)\n\n", r.Summary.Added, r.Summary.Removed, r.Summary.Modified, r.Summary.Total))
	for _, c := range r.Changes {
		icon := map[string]string{"added": "+", "removed": "-", "modified": "~"}[c.Action]
		sb.WriteString(fmt.Sprintf("  [%s] %s  %s (%s)\n", icon, c.Action, c.Name, c.Type))
		for _, d := range c.Details {
			sb.WriteString(fmt.Sprintf("      %s: %q → %q\n", d.Field, d.Before, d.After))
		}
	}
	return sb.String()
}
