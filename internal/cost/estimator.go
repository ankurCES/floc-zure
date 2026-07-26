// Package cost provides Azure resource cost estimation based on simulated resources.
// It uses embedded pricing data for common resource types and calculates
// monthly cost estimates without requiring an Azure subscription.
package cost

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// PricingTier maps SKU/size names to monthly USD prices.
type PricingTier struct {
	Name         string  `json:"name"`
	MonthlyCost  float64 `json:"monthlyCost"`
	HourlyCost   float64 `json:"hourlyCost,omitempty"`
	Unit         string  `json:"unit"` // "month", "hour", "GB/month", etc.
	Description  string  `json:"description,omitempty"`
}

// ResourcePricing holds pricing tiers for a resource type.
type ResourcePricing struct {
	ResourceType string        `json:"resourceType"`
	Provider     string        `json:"provider"`
	Tiers        []PricingTier `json:"tiers"`
}

// CostEstimate represents the estimated cost for a single resource.
type CostEstimate struct {
	ResourceName string  `json:"resourceName"`
	ResourceType string  `json:"resourceType"`
	SKU          string  `json:"sku"`
	Location     string  `json:"location"`
	MonthlyCost  float64 `json:"monthlyCost"`
	Currency     string  `json:"currency"`
	Notes        string  `json:"notes,omitempty"`
}

// CostReport is the full cost estimation report for all resources.
type CostReport struct {
	Resources    []CostEstimate `json:"resources"`
	TotalMonthly float64        `json:"totalMonthlyCost"`
	Currency     string         `json:"currency"`
	Disclaimer   string         `json:"disclaimer"`
}

// ResourceInput describes a resource for cost estimation.
type ResourceInput struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Location     string                 `json:"location"`
	SKU          string                 `json:"sku,omitempty"`
	Size         string                 `json:"size,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// Estimator calculates cost estimates for Azure resources.
type Estimator struct {
	pricing map[string]*ResourcePricing // lowercase resource type -> pricing
}

// NewEstimator creates a cost estimator with built-in pricing data.
func NewEstimator() *Estimator {
	e := &Estimator{
		pricing: make(map[string]*ResourcePricing),
	}
	e.loadBuiltinPricing()
	return e
}

// Estimate calculates costs for a list of resources.
func (e *Estimator) Estimate(resources []ResourceInput) *CostReport {
	report := &CostReport{
		Currency:   "USD",
		Disclaimer: "Estimates are approximate and based on list prices as of 2024. Actual costs may vary based on region, discounts, reserved instances, and usage patterns.",
	}

	for _, res := range resources {
		est := e.estimateResource(res)
		report.Resources = append(report.Resources, est)
		report.TotalMonthly += est.MonthlyCost
	}

	// Sort by cost descending
	sort.Slice(report.Resources, func(i, j int) bool {
		return report.Resources[i].MonthlyCost > report.Resources[j].MonthlyCost
	})

	return report
}

// EstimateFromJSON estimates costs from a JSON array of ResourceInput.
func (e *Estimator) EstimateFromJSON(data []byte) (*CostReport, error) {
	var resources []ResourceInput
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("parse resource list: %w", err)
	}
	return e.Estimate(resources), nil
}

// GetPricing returns the pricing data for a resource type.
func (e *Estimator) GetPricing(resourceType string) (*ResourcePricing, bool) {
	p, ok := e.pricing[strings.ToLower(resourceType)]
	return p, ok
}

// estimateResource calculates the cost for a single resource.
func (e *Estimator) estimateResource(res ResourceInput) CostEstimate {
	est := CostEstimate{
		ResourceName: res.Name,
		ResourceType: res.Type,
		SKU:          res.SKU,
		Location:     res.Location,
		Currency:     "USD",
	}

	// Determine the SKU to look up
	sku := res.SKU
	if sku == "" {
		sku = res.Size
	}
	est.SKU = sku

	pricing, ok := e.pricing[strings.ToLower(res.Type)]
	if !ok {
		est.Notes = "No pricing data available for this resource type"
		return est
	}

	// Find matching tier
	tier := e.findTier(pricing, sku)
	if tier == nil {
		// Fall back to first tier (default/basic)
		if len(pricing.Tiers) > 0 {
			tier = &pricing.Tiers[0]
			est.Notes = fmt.Sprintf("SKU %q not found, using default tier %q", sku, tier.Name)
		} else {
			est.Notes = "No pricing tiers available"
			return est
		}
	}

	est.MonthlyCost = tier.MonthlyCost
	if est.Notes == "" && tier.Description != "" {
		est.Notes = tier.Description
	}

	return est
}

// findTier matches a SKU name against available pricing tiers.
func (e *Estimator) findTier(pricing *ResourcePricing, sku string) *PricingTier {
	if sku == "" {
		return nil
	}
	skuLower := strings.ToLower(sku)
	for i := range pricing.Tiers {
		if strings.ToLower(pricing.Tiers[i].Name) == skuLower {
			return &pricing.Tiers[i]
		}
	}
	// Partial match
	for i := range pricing.Tiers {
		if strings.Contains(strings.ToLower(pricing.Tiers[i].Name), skuLower) ||
			strings.Contains(skuLower, strings.ToLower(pricing.Tiers[i].Name)) {
			return &pricing.Tiers[i]
		}
	}
	return nil
}

// FormatText returns a human-readable cost report.
func FormatText(report *CostReport) string {
	var sb strings.Builder
	sb.WriteString("Azure Cost Estimation Report\n")
	sb.WriteString("============================\n\n")

	if len(report.Resources) == 0 {
		sb.WriteString("No resources to estimate.\n")
		return sb.String()
	}

	// Column widths
	maxName := 8
	maxType := 4
	maxSKU := 3
	for _, r := range report.Resources {
		if len(r.ResourceName) > maxName {
			maxName = len(r.ResourceName)
		}
		shortType := shortResourceType(r.ResourceType)
		if len(shortType) > maxType {
			maxType = len(shortType)
		}
		if len(r.SKU) > maxSKU {
			maxSKU = len(r.SKU)
		}
	}

	// Header
	fmt.Fprintf(&sb, "%-*s  %-*s  %-*s  %12s\n", maxName, "Resource", maxType, "Type", maxSKU, "SKU", "Monthly Cost")
	sb.WriteString(strings.Repeat("-", maxName+maxType+maxSKU+18) + "\n")

	for _, r := range report.Resources {
		fmt.Fprintf(&sb, "%-*s  %-*s  %-*s  %11s\n",
			maxName, r.ResourceName,
			maxType, shortResourceType(r.ResourceType),
			maxSKU, r.SKU,
			formatUSD(r.MonthlyCost))
		if r.Notes != "" {
			fmt.Fprintf(&sb, "%s  ↳ %s\n", strings.Repeat(" ", maxName), r.Notes)
		}
	}

	sb.WriteString(strings.Repeat("-", maxName+maxType+maxSKU+18) + "\n")
	fmt.Fprintf(&sb, "%-*s  %*s  %11s\n", maxName, "TOTAL", maxType+maxSKU+2, "", formatUSD(report.TotalMonthly))
	sb.WriteString("\n")
	sb.WriteString(report.Disclaimer + "\n")

	return sb.String()
}

// shortResourceType extracts the last part of a resource type (e.g. "storageAccounts").
func shortResourceType(rt string) string {
	parts := strings.Split(rt, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return rt
}

func formatUSD(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}
