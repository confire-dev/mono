// Plan intent helpers — shared across login and auth callback.
//
// A plan intent is only an expression of what the user wants.
// It never grants entitlement. Paid plan validity is always
// confirmed server-side before a Stripe checkout is created.

const SLUG_RE = /^[a-z0-9_-]+$/
const MAX_LEN = 64

// normalizePlanIntent converts any incoming value to a safe slug.
// Returns "free" for anything missing, malformed, or suspiciously long.
export function normalizePlanIntent(value: unknown): string {
  if (typeof value !== 'string') return 'free'
  const s = value.toLowerCase().trim()
  if (!s || s.length > MAX_LEN || !SLUG_RE.test(s)) return 'free'
  return s
}

export function isFreePlan(planIntent: string): boolean {
  return planIntent === 'free'
}
