//go:build unit

package service

import "testing"

// 按次定价：quality 标签档优先于尺寸档，未命中时回退默认按次价（xinfeng -02 系列：low 一价、其余一价）。
func TestCalculatePerRequestCostQualityTierPrecedesSizeTier(t *testing.T) {
	low, size1k, def := 0.0032, 0.05, 0.008
	resolved := &ResolvedPricing{
		Mode:                   BillingModePerRequest,
		DefaultPerRequestPrice: def,
		RequestTiers: []PricingInterval{
			{TierLabel: "low", PerRequestPrice: &low},
			{TierLabel: "1K", PerRequestPrice: &size1k},
		},
	}
	svc := &BillingService{}
	resolver := &ModelPricingResolver{}
	cases := []struct {
		name, quality, size string
		count               int
		want                float64
	}{
		{"low quality wins over size tier", "low", "1K", 1, low},
		{"quality must match exactly as forwarded upstream", "LOW", "1K", 1, size1k},
		{"whitespace variant is not the low tier", " low", "4K", 1, def},
		{"unknown quality falls back to size tier", "high", "1K", 1, size1k},
		{"unknown quality and size fall back to default", "high", "4K", 2, def * 2},
		{"empty quality keeps size behaviour", "", "1K", 1, size1k},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.calculatePerRequestCost(resolved, CostInput{
				RequestCount: tc.count, SizeTier: tc.size, QualityTier: tc.quality,
				RateMultiplier: 1, Resolver: resolver,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got.TotalCost != tc.want {
				t.Fatalf("total=%v want %v", got.TotalCost, tc.want)
			}
		})
	}
}
