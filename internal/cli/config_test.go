package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ankurCES/floc-zure/pkg/models"
)

func TestConfigCmd_HasSubcommands(t *testing.T) {
	root := RootCmd()
	cfgCmd, _, err := root.Find([]string{"config"})
	if err != nil {
		t.Fatalf("config command not found: %v", err)
	}
	names := make(map[string]bool)
	for _, c := range cfgCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"init", "set", "get", "list"} {
		if !names[want] {
			t.Errorf("config missing subcommand %q", want)
		}
	}
}

func TestConfigLookup(t *testing.T) {
	cfg := &models.Config{
		Subscription: "sub-123",
		Location:     "westus2",
		OutputFormat: "table",
		Verbose:      true,
		Defaults: models.ConfigDefaults{
			ResourceGroup: "rg-default",
			Location:      "northeurope",
		},
	}

	tests := []struct {
		key  string
		want string
	}{
		{"subscription", "sub-123"},
		{"location", "westus2"},
		{"output_format", "table"},
		{"verbose", "true"},
		{"defaults.resource_group", "rg-default"},
		{"defaults.location", "northeurope"},
		{"nonexistent", "(unknown key)"},
	}

	for _, tt := range tests {
		got := configLookup(cfg, tt.key)
		if got != tt.want {
			t.Errorf("configLookup(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestConfigListCmd_OutputFormat(t *testing.T) {
	buf := new(bytes.Buffer)
	root := RootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "list"})
	err := root.Execute()
	if err != nil {
		t.Fatalf("config list failed: %v", err)
	}
	out := buf.String()
	// Should contain key names
	if !strings.Contains(out, "location") {
		t.Errorf("expected 'location' in output: %s", out)
	}
	if !strings.Contains(out, "output_format") {
		t.Errorf("expected 'output_format' in output: %s", out)
	}
}

func TestConfigGetCmd_Location(t *testing.T) {
	buf := new(bytes.Buffer)
	root := RootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "get", "location"})
	err := root.Execute()
	if err != nil {
		t.Fatalf("config get failed: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	// Default location is "eastus"
	if out != "eastus" {
		t.Errorf("expected 'eastus', got %q", out)
	}
}

func TestConfigSetCmd_BadFormat(t *testing.T) {
	root := RootCmd()
	root.SetArgs([]string{"config", "set", "no-equals-sign"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for bad KEY=VALUE format")
	}
}

func TestGroupCmd_HasSubcommands(t *testing.T) {
	root := RootCmd()
	grpCmd, _, err := root.Find([]string{"group"})
	if err != nil {
		t.Fatalf("group command not found: %v", err)
	}
	names := make(map[string]bool)
	for _, c := range grpCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"create", "list", "show", "delete"} {
		if !names[want] {
			t.Errorf("group missing subcommand %q", want)
		}
	}
}

func TestResourceCmd_HasSubcommands(t *testing.T) {
	root := RootCmd()
	resCmd, _, err := root.Find([]string{"resource"})
	if err != nil {
		t.Fatalf("resource command not found: %v", err)
	}
	names := make(map[string]bool)
	for _, c := range resCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"list", "show", "delete", "tag"} {
		if !names[want] {
			t.Errorf("resource missing subcommand %q", want)
		}
	}
}

func TestWorkflowCmd_HasSubcommands(t *testing.T) {
	root := RootCmd()
	wfCmd, _, err := root.Find([]string{"workflow"})
	if err != nil {
		t.Fatalf("workflow command not found: %v", err)
	}
	names := make(map[string]bool)
	for _, c := range wfCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"run", "validate"} {
		if !names[want] {
			t.Errorf("workflow missing subcommand %q", want)
		}
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		input []string
		want  map[string]string
	}{
		{nil, nil},
		{[]string{"env=prod", "team=platform"}, map[string]string{"env": "prod", "team": "platform"}},
		{[]string{"noequals"}, map[string]string{}},
		{[]string{"key=val=ue"}, map[string]string{"key": "val=ue"}},
	}
	for _, tt := range tests {
		got := parseTags(tt.input)
		if tt.want == nil && got != nil {
			t.Errorf("parseTags(%v) = %v, want nil", tt.input, got)
			continue
		}
		for k, v := range tt.want {
			if got[k] != v {
				t.Errorf("parseTags(%v)[%s] = %q, want %q", tt.input, k, got[k], v)
			}
		}
	}
}
