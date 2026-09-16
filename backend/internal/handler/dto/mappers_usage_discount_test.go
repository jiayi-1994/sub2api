package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// Components are stored at list price; total_cost carries the hidden global
// discount. The user DTO must not expose that gap.
func TestUsageLogFromService_FoldsGlobalDiscountIntoComponents(t *testing.T) {
	t.Parallel()

	// list components sum to 0.040, total discounted by 0.8 -> 0.032
	l := &service.UsageLog{
		InputCost:      0.010,
		OutputCost:     0.030,
		TotalCost:      0.032,
		ActualCost:     0.048, // 0.032 * rate 1.5
		RateMultiplier: 1.5,
	}

	u := UsageLogFromService(l)
	sum := u.InputCost + u.OutputCost + u.CacheCreationCost + u.CacheReadCost +
		u.ImageInputCost + u.ImageOutputCost
	require.InDelta(t, u.TotalCost, sum, 1e-12, "user components must sum to total_cost")
	require.InDelta(t, 0.008, u.InputCost, 1e-12)
	require.InDelta(t, 0.024, u.OutputCost, 1e-12)
	require.Equal(t, 0.032, u.TotalCost, "total_cost untouched")
	require.Equal(t, 0.048, u.ActualCost, "actual_cost untouched")

	a := UsageLogFromServiceAdmin(l)
	require.Equal(t, 0.010, a.InputCost, "admin keeps list-price components")
	require.Equal(t, 0.030, a.OutputCost)
	require.Equal(t, 0.032, a.TotalCost)
}

func TestUsageLogFromService_LeavesUndiscountedRowsAlone(t *testing.T) {
	t.Parallel()

	// image billing: single component already equals total -> no-op
	img := UsageLogFromService(&service.UsageLog{ImageOutputCost: 0.5, TotalCost: 0.5})
	require.Equal(t, 0.5, img.ImageOutputCost)

	// no component breakdown at all -> no-op, no divide by zero
	bare := UsageLogFromService(&service.UsageLog{TotalCost: 0.75})
	require.Equal(t, 0.75, bare.TotalCost)
	require.Zero(t, bare.InputCost)
}
