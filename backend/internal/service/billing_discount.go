package service

import "context"

func resolveGlobalBillingRateMultiplier(ctx context.Context, settingService *SettingService) float64 {
	if settingService == nil {
		return 1.0
	}
	return normalizeGlobalBillingRateMultiplier(settingService.GetGlobalBillingRateMultiplier(ctx))
}

// applyGlobalBillingRateMultiplier folds the hidden global billing discount into
// total and billed costs before the usage log is saved and before user
// balance/subscription/API-key quota deductions are computed.
//
// Per-token cost components intentionally stay at their pre-global-discount
// values. The UI derives the row-level global discount from total_cost and
// actual_cost when it needs to render discounted component costs; unit-price
// display can then remain based on the stable pre-discount component costs.
//
// Fixed per-unit pricing — image generation, video generation, and per-request
// billing — is charged at face value and is NOT discounted: those modes price
// by size/count (e.g. 1k vs 2k image) and must bill exactly the configured
// amount. The hidden global discount applies only to token-based billing
// (subscription-quota resale).
func applyGlobalBillingRateMultiplier(cost *CostBreakdown, multiplier float64) {
	if cost == nil {
		return
	}
	switch cost.BillingMode {
	case string(BillingModeImage), string(BillingModeVideo), string(BillingModePerRequest):
		return
	}
	multiplier = normalizeGlobalBillingRateMultiplier(multiplier)
	if multiplier == 1.0 {
		return
	}
	cost.TotalCost *= multiplier
	cost.ActualCost *= multiplier
}
