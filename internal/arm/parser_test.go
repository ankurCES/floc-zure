package arm

import (
	"testing"
)

func TestParseTemplateBytes_valid(t *testing.T) {
	data := []byte(`{
		"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": {},
		"resources": [
			{"type": "Microsoft.Storage/storageAccounts", "apiVersion": "2023-01-01", "name": "test", "location": "eastus"}
		]
	}`)
	tmpl, err := ParseTemplateBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tmpl.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(tmpl.Resources))
	}
	if tmpl.Resources[0].Name != "test" {
		t.Errorf("expected name 'test', got %q", tmpl.Resources[0].Name)
	}
}

func TestParseTemplateBytes_noResources(t *testing.T) {
	data := []byte(`{"$schema":"x","contentVersion":"1.0.0.0","resources":[]}`)
	_, err := ParseTemplateBytes(data)
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}

func TestParseTemplateBytes_invalidJSON(t *testing.T) {
	_, err := ParseTemplateBytes([]byte(`{bad json}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseParameterBytes(t *testing.T) {
	data := []byte(`{
		"parameters": {
			"name": {"value": "hello"},
			"count": {"value": 42}
		}
	}`)
	params, err := ParseParameterBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params["name"] != "hello" {
		t.Errorf("expected 'hello', got %v", params["name"])
	}
	if params["count"] != float64(42) {
		t.Errorf("expected 42, got %v", params["count"])
	}
}

func TestResolve_parametersAndVariables(t *testing.T) {
	tmpl := &Template{
		Parameters: map[string]ParameterDef{
			"env":      {Type: "string", DefaultValue: "dev"},
			"location": {Type: "string"},
		},
		Variables: map[string]interface{}{
			"storageName": "[concat(parameters('env'), 'storage')]",
		},
		Resources: []ResourceDef{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "[variables('storageName')]",
				Location: "[parameters('location')]",
			},
		},
	}
	result, err := Resolve(tmpl, map[string]interface{}{"location": "westus2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Parameters["env"] != "dev" {
		t.Errorf("expected default 'dev', got %v", result.Parameters["env"])
	}
	if result.Parameters["location"] != "westus2" {
		t.Errorf("expected 'westus2', got %v", result.Parameters["location"])
	}
	if result.Resources[0].Name != "devstorage" {
		t.Errorf("expected 'devstorage', got %q", result.Resources[0].Name)
	}
	if result.Resources[0].Location != "westus2" {
		t.Errorf("expected location 'westus2', got %q", result.Resources[0].Location)
	}
}

func TestResolve_missingRequired(t *testing.T) {
	tmpl := &Template{
		Parameters: map[string]ParameterDef{
			"name": {Type: "string"}, // no default
		},
		Resources: []ResourceDef{
			{Type: "Microsoft.Storage/storageAccounts", Name: "x", Location: "eastus"},
		},
	}
	_, err := Resolve(tmpl, nil)
	if err == nil {
		t.Fatal("expected error for missing required parameter")
	}
}

func TestResolve_concat(t *testing.T) {
	tmpl := &Template{
		Parameters: map[string]ParameterDef{
			"a": {Type: "string", DefaultValue: "hello"},
			"b": {Type: "string", DefaultValue: "world"},
		},
		Resources: []ResourceDef{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "[concat(parameters('a'), '-', parameters('b'))]",
				Location: "eastus",
			},
		},
	}
	result, err := Resolve(tmpl, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resources[0].Name != "hello-world" {
		t.Errorf("expected 'hello-world', got %q", result.Resources[0].Name)
	}
}

func TestResolve_toLower(t *testing.T) {
	tmpl := &Template{
		Parameters: map[string]ParameterDef{
			"name": {Type: "string", DefaultValue: "MyStorage"},
		},
		Resources: []ResourceDef{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "[toLower(parameters('name'))]",
				Location: "eastus",
			},
		},
	}
	result, err := Resolve(tmpl, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resources[0].Name != "mystorage" {
		t.Errorf("expected 'mystorage', got %q", result.Resources[0].Name)
	}
}

func TestResolve_resourceGroupLocation(t *testing.T) {
	tmpl := &Template{
		Parameters: map[string]ParameterDef{
			"location": {Type: "string", DefaultValue: "centralus"},
		},
		Resources: []ResourceDef{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "test",
				Location: "[resourceGroup().location]",
			},
		},
	}
	result, err := Resolve(tmpl, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// resourceGroup().location falls back to parameters.location
	if result.Resources[0].Location != "centralus" {
		t.Errorf("expected 'centralus', got %q", result.Resources[0].Location)
	}
}

func TestResolve_tags(t *testing.T) {
	tmpl := &Template{
		Parameters: map[string]ParameterDef{
			"env": {Type: "string", DefaultValue: "prod"},
		},
		Resources: []ResourceDef{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "test",
				Location: "eastus",
				Tags:     map[string]string{"environment": "[parameters('env')]"},
			},
		},
	}
	result, err := Resolve(tmpl, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resources[0].Tags["environment"] != "prod" {
		t.Errorf("expected tag 'prod', got %q", result.Resources[0].Tags["environment"])
	}
}

func TestResolve_uniqueString(t *testing.T) {
	tmpl := &Template{
		Resources: []ResourceDef{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "[uniqueString('test')]",
				Location: "eastus",
			},
		},
	}
	result, err := Resolve(tmpl, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resources[0].Name == "" || result.Resources[0].Name == "uniqueString('test')" {
		t.Errorf("expected resolved uniqueString, got %q", result.Resources[0].Name)
	}
	// Should be deterministic
	result2, _ := Resolve(tmpl, nil)
	if result.Resources[0].Name != result2.Resources[0].Name {
		t.Error("uniqueString should be deterministic")
	}
}

func TestResolve_format(t *testing.T) {
	tmpl := &Template{
		Parameters: map[string]ParameterDef{
			"prefix": {Type: "string", DefaultValue: "app"},
		},
		Resources: []ResourceDef{
			{
				Type:     "Microsoft.Storage/storageAccounts",
				Name:     "[format('{0}-storage', parameters('prefix'))]",
				Location: "eastus",
			},
		},
	}
	result, err := Resolve(tmpl, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resources[0].Name != "app-storage" {
		t.Errorf("expected 'app-storage', got %q", result.Resources[0].Name)
	}
}

func TestExtractFuncArgs(t *testing.T) {
	args := extractFuncArgs("concat('a', 'b', 'c')", "concat")
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	if args[0] != "'a'" || args[1] != "'b'" || args[2] != "'c'" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestExtractFuncArgs_nested(t *testing.T) {
	args := extractFuncArgs("concat(parameters('a'), '-', variables('b'))", "concat")
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	if args[0] != "parameters('a')" {
		t.Errorf("expected parameters('a'), got %q", args[0])
	}
}

func TestResourceTypeToCommand(t *testing.T) {
	tests := []struct {
		resourceType string
		wantOK       bool
		wantPrefix0  string
	}{
		{"Microsoft.Storage/storageAccounts", true, "storage"},
		{"Microsoft.KeyVault/vaults", true, "keyvault"},
		{"Microsoft.Network/virtualNetworks", true, "network"},
		{"Microsoft.Compute/virtualMachines", true, "vm"},
		{"Microsoft.Foo/bars", false, ""},
	}
	for _, tc := range tests {
		cm, ok := resourceTypeToCommand(tc.resourceType)
		if ok != tc.wantOK {
			t.Errorf("%s: expected ok=%v, got %v", tc.resourceType, tc.wantOK, ok)
			continue
		}
		if ok && cm.prefix[0] != tc.wantPrefix0 {
			t.Errorf("%s: expected prefix[0]=%q, got %q", tc.resourceType, tc.wantPrefix0, cm.prefix[0])
		}
	}
}
