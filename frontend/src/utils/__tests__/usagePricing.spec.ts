import { describe, expect, it } from 'vitest'

import {
  calculateDiscountedUsageCost,
  calculateTokenPricePerMillion,
  calculateTokenUnitPrice,
  formatTokenPricePerMillion,
  resolveUsageCostDiscountMultiplier,
} from '@/utils/usagePricing'

describe('usagePricing', () => {
  it('keeps legacy behavior when no global multiplier context is provided', () => {
    expect(calculateTokenUnitPrice(0.5, 2)).toBeCloseTo(0.25)
    expect(calculateTokenPricePerMillion(0.5, 2)).toBeCloseTo(250000)
    expect(formatTokenPricePerMillion(0.5, 2)).toBe('$250000.0000')
  })

  it('keeps unit price stable while component costs are discounted for display', () => {
    const usage = {
      input_cost: 10,
      output_cost: 30,
      cache_creation_cost: 5,
      cache_read_cost: 5,
      total_cost: 25,
    }

    expect(resolveUsageCostDiscountMultiplier(usage)).toBeCloseTo(0.5)
    expect(calculateDiscountedUsageCost(usage.input_cost, usage)).toBeCloseTo(5)
    expect(calculateTokenPricePerMillion(usage.input_cost, 2)).toBeCloseTo(5000000)
  })

  it('includes image token costs when deriving the usage discount multiplier', () => {
    const usage = {
      input_cost: 10,
      output_cost: 20,
      cache_creation_cost: 5,
      cache_read_cost: 5,
      image_input_cost: 15,
      image_output_cost: 5,
      total_cost: 30,
    }

    expect(resolveUsageCostDiscountMultiplier(usage)).toBeCloseTo(0.5)
    expect(calculateDiscountedUsageCost(usage.input_cost, usage)).toBeCloseTo(5)
    expect(calculateDiscountedUsageCost(usage.image_input_cost, usage)).toBeCloseTo(7.5)
    expect(calculateDiscountedUsageCost(usage.image_output_cost, usage)).toBeCloseTo(2.5)
  })

  it('falls back to cost divided by tokens when no unit price snapshot is provided', () => {
    expect(
      calculateTokenUnitPrice(0.5, 2)
    ).toBeCloseTo(0.25)
  })
})
