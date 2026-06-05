package service

import "context"

func resolveGlobalBillingRateMultiplier(ctx context.Context, settingService *SettingService) float64 {
	if settingService == nil {
		return 1.0
	}
	return normalizeGlobalBillingRateMultiplier(settingService.GetGlobalBillingRateMultiplier(ctx))
}

// applyGlobalBillingRateMultiplier folds the hidden global billing discount into
// every customer-facing cost component before the usage log is saved and before
// user balance/subscription/API-key quota deductions are computed.
func applyGlobalBillingRateMultiplier(cost *CostBreakdown, multiplier float64) {
	if cost == nil {
		return
	}
	multiplier = normalizeGlobalBillingRateMultiplier(multiplier)
	if multiplier == 1.0 {
		return
	}
	cost.InputCost *= multiplier
	cost.OutputCost *= multiplier
	cost.ImageOutputCost *= multiplier
	cost.CacheCreationCost *= multiplier
	cost.CacheReadCost *= multiplier
	cost.TotalCost *= multiplier
	cost.ActualCost *= multiplier
}
