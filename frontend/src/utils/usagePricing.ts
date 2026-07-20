export const TOKENS_PER_MILLION = 1_000_000

interface TokenPriceFormatOptions {
  fractionDigits?: number
  withCurrencySymbol?: boolean
  emptyValue?: string
}

interface UsageCostDiscountContext {
  input_cost?: number | null
  output_cost?: number | null
  cache_creation_cost?: number | null
  cache_read_cost?: number | null
  image_input_cost?: number | null
  image_output_cost?: number | null
  total_cost?: number | null
}

function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

export function calculateTokenUnitPrice(
  cost: number | null | undefined,
  tokens: number | null | undefined
): number | null {
  if (!isFiniteNumber(cost) || !isFiniteNumber(tokens) || tokens <= 0) {
    return null
  }

  return cost / tokens
}

export function calculateTokenPricePerMillion(
  cost: number | null | undefined,
  tokens: number | null | undefined
): number | null {
  const unitPrice = calculateTokenUnitPrice(cost, tokens)
  if (unitPrice == null) {
    return null
  }

  return unitPrice * TOKENS_PER_MILLION
}

export function formatTokenPricePerMillion(
  cost: number | null | undefined,
  tokens: number | null | undefined,
  options: TokenPriceFormatOptions = {}
): string {
  const pricePerMillion = calculateTokenPricePerMillion(cost, tokens)
  if (pricePerMillion == null) {
    return options.emptyValue ?? '-'
  }

  const fractionDigits = options.fractionDigits ?? 4
  const formatted = pricePerMillion.toFixed(fractionDigits)
  return options.withCurrencySymbol == false ? formatted : `$${formatted}`
}

export function resolveUsageCostDiscountMultiplier(context: UsageCostDiscountContext | null | undefined): number {
  if (!context || !isFiniteNumber(context.total_cost)) {
    return 1
  }

  const componentTotal = [
    context.input_cost,
    context.output_cost,
    context.cache_creation_cost,
    context.cache_read_cost,
    context.image_input_cost,
    context.image_output_cost,
  ]
    .filter(isFiniteNumber)
    .reduce((sum, value) => sum + value, 0)

  if (componentTotal <= 0) {
    return 1
  }

  const multiplier = context.total_cost / componentTotal
  return Number.isFinite(multiplier) && multiplier >= 0 ? multiplier : 1
}

export function calculateDiscountedUsageCost(
  cost: number | null | undefined,
  context: UsageCostDiscountContext | null | undefined
): number {
  if (!isFiniteNumber(cost)) {
    return 0
  }

  return cost * resolveUsageCostDiscountMultiplier(context)
}
