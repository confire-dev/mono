// canUseRemoteOptimizer — the single entitlement gate for all remote optimizations.
//
// NEVER do: if (user.plan === 'pro') allowFigma()
// ALWAYS do: const check = canUseRemoteOptimizer({...}); if (check.allowed) ...
//
// All capabilities and limits come from the Plan object.
// The Worker fetches the user's plan_id from profiles, looks up PLANS[plan_id],
// then passes it here. No plan-name strings in business logic.

import type { Plan } from './plans.js'

export type EntitlementReason =
  | 'not_logged_in'
  | 'plan_required'           // plan.features.remoteOptimization = false
  | 'optimizer_not_enabled'   // optimizer not in plan.optimizers.remote
  | 'payload_too_large'       // payloadBytes > plan.limits.maxPayloadBytes
  | 'token_limit_per_call'    // rawTokensEstimate > plan.limits.maxRawTokensPerOptimization
  | 'monthly_optimizations_exceeded'
  | 'monthly_tokens_exceeded'

export interface EntitlementCheck {
  allowed: boolean
  reason?: EntitlementReason
  usage?: {
    used: number
    limit: number
    type: 'optimizations' | 'tokens'
  }
  // True when usage is ≥80% of limit — show a nudge but still allow
  approachingLimit?: boolean
  approachingLimitMessage?: string
}

export interface UsageThisPeriod {
  cloudOptimizationsUsed: number
  cloudTokensUsed: number
}

export interface EntitlementArgs {
  plan: Plan
  optimizer: string           // e.g. "figma", "github_pr"
  payloadBytes: number
  rawTokensEstimate: number   // estimate: payloadBytes / 4
  usageThisPeriod: UsageThisPeriod
}

export function canUseRemoteOptimizer(args: EntitlementArgs): EntitlementCheck {
  const { plan, optimizer, payloadBytes, rawTokensEstimate, usageThisPeriod } = args
  const { limits, features, optimizers } = plan

  // ── Capability checks (plan features) ─────────────────────────────────────

  if (!features.remoteOptimization) {
    return { allowed: false, reason: 'plan_required' }
  }

  if (!optimizers.remote.includes(optimizer)) {
    return { allowed: false, reason: 'optimizer_not_enabled' }
  }

  // ── Per-call size checks ───────────────────────────────────────────────────

  if (payloadBytes > limits.maxPayloadBytes) {
    return {
      allowed: false,
      reason: 'payload_too_large',
      usage: { used: payloadBytes, limit: limits.maxPayloadBytes, type: 'optimizations' },
    }
  }

  if (rawTokensEstimate > limits.maxRawTokensPerOptimization) {
    return {
      allowed: false,
      reason: 'token_limit_per_call',
      usage: { used: rawTokensEstimate, limit: limits.maxRawTokensPerOptimization, type: 'tokens' },
    }
  }

  // ── Monthly usage checks ───────────────────────────────────────────────────

  if (usageThisPeriod.cloudOptimizationsUsed >= limits.cloudOptimizationsMonthly) {
    return {
      allowed: false,
      reason: 'monthly_optimizations_exceeded',
      usage: {
        used:  usageThisPeriod.cloudOptimizationsUsed,
        limit: limits.cloudOptimizationsMonthly,
        type:  'optimizations',
      },
    }
  }

  if (usageThisPeriod.cloudTokensUsed >= limits.cloudTokensMonthly) {
    return {
      allowed: false,
      reason: 'monthly_tokens_exceeded',
      usage: {
        used:  usageThisPeriod.cloudTokensUsed,
        limit: limits.cloudTokensMonthly,
        type:  'tokens',
      },
    }
  }

  // ── Allowed — check if approaching limit for nudge ────────────────────────

  const optPct = usageThisPeriod.cloudOptimizationsUsed / limits.cloudOptimizationsMonthly
  if (optPct >= 0.8) {
    const remaining = limits.cloudOptimizationsMonthly - usageThisPeriod.cloudOptimizationsUsed
    return {
      allowed: true,
      approachingLimit: true,
      approachingLimitMessage: `${usageThisPeriod.cloudOptimizationsUsed}/${limits.cloudOptimizationsMonthly} cloud optimizations used · ${remaining} remaining`,
      usage: {
        used:  usageThisPeriod.cloudOptimizationsUsed,
        limit: limits.cloudOptimizationsMonthly,
        type:  'optimizations',
      },
    }
  }

  return { allowed: true }
}

// ── Human-readable messages ────────────────────────────────────────────────

export function entitlementMessage(check: EntitlementCheck, planName: string): string {
  if (check.allowed) return ''

  switch (check.reason) {
    case 'monthly_optimizations_exceeded':
      return [
        `⚠️ Confire: ${check.usage?.used}/${check.usage?.limit} cloud optimizations used this month.`,
        planName === 'Free'
          ? 'Upgrade to Dev ($10/mo) or Pro ($20/mo) at confire.dev/upgrade'
          : 'Local optimization still active.',
      ].join(' ')

    case 'payload_too_large':
      return `⚠️ Payload too large for cloud optimization. Max: ${Math.round((check.usage?.limit ?? 0) / 1024)}KB.`

    case 'token_limit_per_call':
      return `⚠️ Tool response too large for cloud optimization (${check.usage?.used?.toLocaleString()} tokens, max ${check.usage?.limit?.toLocaleString()}).`

    case 'optimizer_not_enabled':
      return `Cloud optimizer not available on ${planName} plan.`

    default:
      return 'Cloud optimization not available.'
  }
}
