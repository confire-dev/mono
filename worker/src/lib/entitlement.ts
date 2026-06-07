// canUseFirewall — the single entitlement gate for cloud firewall features.
//
// All capabilities come from the Plan object.
// The Worker fetches the user's plan_id from profiles, looks up PLANS[plan_id],
// then passes it here. No plan-name strings in business logic.

import type { Plan } from './plans.js'

export type EntitlementReason =
  | 'not_logged_in'
  | 'plan_required'
  | 'payload_too_large'

export interface EntitlementCheck {
  allowed: boolean
  reason?: EntitlementReason
  approachingLimit?: boolean
  approachingLimitMessage?: string
}

export interface EntitlementArgs {
  plan: Plan
  payloadBytes: number
}

export function canUseFirewall(args: EntitlementArgs): EntitlementCheck {
  const { plan, payloadBytes } = args

  if (!plan.features.firewallEnabled) {
    return { allowed: false, reason: 'plan_required' }
  }

  if (payloadBytes > plan.limits.maxPayloadBytes) {
    return { allowed: false, reason: 'payload_too_large' }
  }

  return { allowed: true }
}

// ── Human-readable messages ────────────────────────────────────────────────

export function entitlementMessage(check: EntitlementCheck, planName: string): string {
  if (check.allowed) return ''

  switch (check.reason) {
    case 'payload_too_large':
      return `⚠️ Payload too large for cloud firewall sync.`

    case 'plan_required':
      return `Cloud firewall features not available on ${planName} plan.`

    default:
      return 'Cloud firewall not available.'
  }
}
