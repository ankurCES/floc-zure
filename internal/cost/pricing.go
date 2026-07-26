package cost

// loadBuiltinPricing populates the estimator with approximate Azure list prices.
func (e *Estimator) loadBuiltinPricing() {
	data := []ResourcePricing{
		{
			ResourceType: "microsoft.storage/storageaccounts",
			Provider:     "Microsoft.Storage",
			Tiers: []PricingTier{
				{Name: "Standard_LRS", MonthlyCost: 21.84, Unit: "GB/month", Description: "Locally redundant, ~1TB"},
				{Name: "Standard_GRS", MonthlyCost: 43.01, Unit: "GB/month", Description: "Geo-redundant, ~1TB"},
				{Name: "Standard_ZRS", MonthlyCost: 27.26, Unit: "GB/month", Description: "Zone-redundant, ~1TB"},
				{Name: "Premium_LRS", MonthlyCost: 150.00, Unit: "GB/month", Description: "Premium SSD, ~1TB"},
			},
		},
		{
			ResourceType: "microsoft.keyvault/vaults",
			Provider:     "Microsoft.KeyVault",
			Tiers: []PricingTier{
				{Name: "standard", MonthlyCost: 0.03, Unit: "per 10k operations", Description: "Standard tier"},
				{Name: "premium", MonthlyCost: 5.00, Unit: "month", Description: "Premium tier with HSM"},
			},
		},
		{
			ResourceType: "microsoft.compute/virtualmachines",
			Provider:     "Microsoft.Compute",
			Tiers: []PricingTier{
				{Name: "Standard_B1s", MonthlyCost: 7.59, HourlyCost: 0.0104, Unit: "hour", Description: "1 vCPU, 1 GiB RAM"},
				{Name: "Standard_B2s", MonthlyCost: 30.37, HourlyCost: 0.0416, Unit: "hour", Description: "2 vCPU, 4 GiB RAM"},
				{Name: "Standard_B2ms", MonthlyCost: 60.74, HourlyCost: 0.0832, Unit: "hour", Description: "2 vCPU, 8 GiB RAM"},
				{Name: "Standard_D2s_v3", MonthlyCost: 70.08, HourlyCost: 0.096, Unit: "hour", Description: "2 vCPU, 8 GiB RAM"},
				{Name: "Standard_D4s_v3", MonthlyCost: 140.16, HourlyCost: 0.192, Unit: "hour", Description: "4 vCPU, 16 GiB RAM"},
				{Name: "Standard_D8s_v3", MonthlyCost: 280.32, HourlyCost: 0.384, Unit: "hour", Description: "8 vCPU, 32 GiB RAM"},
				{Name: "Standard_E2s_v3", MonthlyCost: 91.98, HourlyCost: 0.126, Unit: "hour", Description: "2 vCPU, 16 GiB RAM"},
				{Name: "Standard_F2s_v2", MonthlyCost: 61.32, HourlyCost: 0.084, Unit: "hour", Description: "2 vCPU, 4 GiB RAM"},
			},
		},
		{
			ResourceType: "microsoft.network/virtualnetworks",
			Provider:     "Microsoft.Network",
			Tiers: []PricingTier{
				{Name: "default", MonthlyCost: 0.00, Unit: "month", Description: "VNets are free; peering charged separately"},
			},
		},
		{
			ResourceType: "microsoft.network/publicipaddresses",
			Provider:     "Microsoft.Network",
			Tiers: []PricingTier{
				{Name: "Static", MonthlyCost: 3.65, Unit: "month", Description: "Static public IP"},
				{Name: "Dynamic", MonthlyCost: 2.63, Unit: "month", Description: "Dynamic public IP"},
				{Name: "default", MonthlyCost: 3.65, Unit: "month", Description: "Public IP (default static)"},
			},
		},
		{
			ResourceType: "microsoft.network/networksecuritygroups",
			Provider:     "Microsoft.Network",
			Tiers: []PricingTier{
				{Name: "default", MonthlyCost: 0.00, Unit: "month", Description: "NSGs are free"},
			},
		},
	}
	for i := range data {
		e.pricing[data[i].ResourceType] = &data[i]
	}
}
