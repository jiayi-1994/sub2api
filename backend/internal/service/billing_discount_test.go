//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyGlobalBillingRateMultiplier_DiscountsTokenBillingOnly(t *testing.T) {
	const multiplier = 0.5

	t.Run("token mode is discounted", func(t *testing.T) {
		cost := &CostBreakdown{BillingMode: string(BillingModeToken), TotalCost: 10, ActualCost: 8}
		applyGlobalBillingRateMultiplier(cost, multiplier)
		require.InDelta(t, 5, cost.TotalCost, 1e-12)
		require.InDelta(t, 4, cost.ActualCost, 1e-12)
	})

	t.Run("empty mode defaults to token and is discounted", func(t *testing.T) {
		cost := &CostBreakdown{BillingMode: "", TotalCost: 10, ActualCost: 8}
		applyGlobalBillingRateMultiplier(cost, multiplier)
		require.InDelta(t, 5, cost.TotalCost, 1e-12)
		require.InDelta(t, 4, cost.ActualCost, 1e-12)
	})

	// Fixed per-unit pricing bills exactly the configured amount, never discounted.
	for _, mode := range []BillingMode{BillingModeImage, BillingModeVideo, BillingModePerRequest} {
		t.Run(string(mode)+" mode keeps face value", func(t *testing.T) {
			cost := &CostBreakdown{BillingMode: string(mode), TotalCost: 10, ActualCost: 8}
			applyGlobalBillingRateMultiplier(cost, multiplier)
			require.InDelta(t, 10, cost.TotalCost, 1e-12)
			require.InDelta(t, 8, cost.ActualCost, 1e-12)
		})
	}

	t.Run("nil cost is safe", func(t *testing.T) {
		require.NotPanics(t, func() { applyGlobalBillingRateMultiplier(nil, multiplier) })
	})
}
