package service

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

type cachedGlobalBillingRateMultiplier struct {
	value     float64
	expiresAt int64 // unix nano
}

const globalBillingRateMultiplierCacheTTL = 60 * time.Second
const globalBillingRateMultiplierErrorTTL = 5 * time.Second
const globalBillingRateMultiplierDBTimeout = 5 * time.Second

func normalizeGlobalBillingRateMultiplier(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 1.0
	}
	if value < 0 {
		return 0
	}
	return value
}

func parseGlobalBillingRateMultiplier(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 1.0
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 1.0
	}
	return normalizeGlobalBillingRateMultiplier(value)
}

// GetGlobalBillingRateMultiplier returns the hidden global billing discount multiplier.
// It is intentionally not part of public settings: ordinary users only see already-discounted
// usage log costs and deductions.
func (s *SettingService) GetGlobalBillingRateMultiplier(ctx context.Context) float64 {
	if s == nil || s.settingRepo == nil {
		return 1.0
	}
	if cached, ok := s.globalBillingRateCache.Load().(*cachedGlobalBillingRateMultiplier); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.value
		}
	}

	result, _, _ := s.globalBillingRateSF.Do("global_billing_rate_multiplier", func() (any, error) {
		if cached, ok := s.globalBillingRateCache.Load().(*cachedGlobalBillingRateMultiplier); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return cached.value, nil
			}
		}
		if ctx == nil {
			ctx = context.Background()
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), globalBillingRateMultiplierDBTimeout)
		defer cancel()
		raw, err := s.settingRepo.GetValue(dbCtx, SettingKeyGlobalBillingRateMultiplier)
		if err != nil {
			ttl := globalBillingRateMultiplierErrorTTL
			if errors.Is(err, ErrSettingNotFound) {
				ttl = globalBillingRateMultiplierCacheTTL
			} else {
				slog.Warn("failed to get global billing rate multiplier setting, defaulting to 1.0", "error", err)
			}
			s.globalBillingRateCache.Store(&cachedGlobalBillingRateMultiplier{
				value:     1.0,
				expiresAt: time.Now().Add(ttl).UnixNano(),
			})
			return 1.0, nil
		}
		value := parseGlobalBillingRateMultiplier(raw)
		s.globalBillingRateCache.Store(&cachedGlobalBillingRateMultiplier{
			value:     value,
			expiresAt: time.Now().Add(globalBillingRateMultiplierCacheTTL).UnixNano(),
		})
		return value, nil
	})
	if value, ok := result.(float64); ok {
		return value
	}
	return 1.0
}
