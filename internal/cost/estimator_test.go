package cost

import (
	"strings"
	"testing"
)

func TestEstimate_storageAccount(t *testing.T) {
	e := NewEstimator()
	report := e.Estimate([]ResourceInput{
		{Name: "sa1", Type: "Microsoft.Storage/storageAccounts", SKU: "Standard_LRS", Location: "eastus"},
	})
	if len(report.Resources) != 1 {
		t.Fatalf("expected 1, got %d", len(report.Resources))
	}
	if report.Resources[0].MonthlyCost != 21.84 {
		t.Errorf("expected 21.84, got %.2f", report.Resources[0].MonthlyCost)
	}
	if report.TotalMonthly != 21.84 {
		t.Errorf("expected total 21.84, got %.2f", report.TotalMonthly)
	}
}

func TestEstimate_VM(t *testing.T) {
	e := NewEstimator()
	report := e.Estimate([]ResourceInput{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", SKU: "Standard_D4s_v3", Location: "eastus"},
	})
	if report.Resources[0].MonthlyCost != 140.16 {
		t.Errorf("expected 140.16, got %.2f", report.Resources[0].MonthlyCost)
	}
}

func TestEstimate_unknownType(t *testing.T) {
	e := NewEstimator()
	report := e.Estimate([]ResourceInput{
		{Name: "x", Type: "Microsoft.Foo/bars", Location: "eastus"},
	})
	if report.Resources[0].MonthlyCost != 0 {
		t.Errorf("expected 0 for unknown, got %.2f", report.Resources[0].MonthlyCost)
	}
	if !strings.Contains(report.Resources[0].Notes, "No pricing data") {
		t.Errorf("expected note about missing pricing")
	}
}

func TestEstimate_fallbackSKU(t *testing.T) {
	e := NewEstimator()
	report := e.Estimate([]ResourceInput{
		{Name: "sa1", Type: "Microsoft.Storage/storageAccounts", SKU: "Unknown_SKU", Location: "eastus"},
	})
	if report.Resources[0].MonthlyCost == 0 {
		t.Error("expected fallback to default tier")
	}
	if !strings.Contains(report.Resources[0].Notes, "not found") {
		t.Error("expected fallback note")
	}
}

func TestEstimate_multipleResources(t *testing.T) {
	e := NewEstimator()
	report := e.Estimate([]ResourceInput{
		{Name: "sa1", Type: "Microsoft.Storage/storageAccounts", SKU: "Standard_LRS"},
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", SKU: "Standard_B2s"},
		{Name: "kv1", Type: "Microsoft.KeyVault/vaults", SKU: "standard"},
	})
	if len(report.Resources) != 3 {
		t.Fatalf("expected 3, got %d", len(report.Resources))
	}
	expected := 21.84 + 30.37 + 0.03
	if report.TotalMonthly < expected-0.01 || report.TotalMonthly > expected+0.01 {
		t.Errorf("expected ~%.2f, got %.2f", expected, report.TotalMonthly)
	}
	// Sorted descending by cost
	if report.Resources[0].MonthlyCost < report.Resources[1].MonthlyCost {
		t.Error("expected descending sort")
	}
}

func TestEstimate_freeResources(t *testing.T) {
	e := NewEstimator()
	report := e.Estimate([]ResourceInput{
		{Name: "vnet1", Type: "Microsoft.Network/virtualNetworks", SKU: "default"},
		{Name: "nsg1", Type: "Microsoft.Network/networkSecurityGroups", SKU: "default"},
	})
	if report.TotalMonthly != 0 {
		t.Errorf("expected 0, got %.2f", report.TotalMonthly)
	}
}

func TestEstimateFromJSON(t *testing.T) {
	e := NewEstimator()
	data := []byte(`[{"name":"sa1","type":"Microsoft.Storage/storageAccounts","sku":"Standard_LRS","location":"eastus"}]`)
	report, err := e.EstimateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalMonthly != 21.84 {
		t.Errorf("expected 21.84, got %.2f", report.TotalMonthly)
	}
}

func TestEstimateFromJSON_invalid(t *testing.T) {
	e := NewEstimator()
	_, err := e.EstimateFromJSON([]byte(`{bad}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatText(t *testing.T) {
	report := &CostReport{
		Currency:     "USD",
		TotalMonthly: 52.24,
		Disclaimer:   "Test disclaimer",
		Resources: []CostEstimate{
			{ResourceName: "vm1", ResourceType: "Microsoft.Compute/virtualMachines", SKU: "Standard_B2s", MonthlyCost: 30.37},
			{ResourceName: "sa1", ResourceType: "Microsoft.Storage/storageAccounts", SKU: "Standard_LRS", MonthlyCost: 21.84},
		},
	}
	text := FormatText(report)
	if !strings.Contains(text, "vm1") {
		t.Error("expected vm1 in output")
	}
	if !strings.Contains(text, "$52.24") {
		t.Error("expected total in output")
	}
	if !strings.Contains(text, "TOTAL") {
		t.Error("expected TOTAL label")
	}
}

func TestGetPricing(t *testing.T) {
	e := NewEstimator()
	p, ok := e.GetPricing("Microsoft.Storage/storageAccounts")
	if !ok {
		t.Fatal("expected pricing data")
	}
	if len(p.Tiers) == 0 {
		t.Error("expected tiers")
	}
}

func TestSizeAsSKU(t *testing.T) {
	e := NewEstimator()
	report := e.Estimate([]ResourceInput{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", Size: "Standard_B1s"},
	})
	if report.Resources[0].MonthlyCost != 7.59 {
		t.Errorf("expected 7.59, got %.2f", report.Resources[0].MonthlyCost)
	}
}
