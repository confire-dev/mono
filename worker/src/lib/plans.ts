// Plans are loaded from Cloudflare KV — immediately available to every Worker instance,
// globally replicated, ~1ms read latency.
//
// Source of truth: Supabase `plans` table (edit there via dashboard or admin API).
// KV is the runtime store. A sync writes from Supabase → KV when plans change.
//
// Sync triggers:
//   POST /admin/plans/sync  → reads all plans from Supabase, writes to KV
//   Called by: wrangler deploy script, admin UI, or whenever you update a plan in DB.
//
// Module-level cache:
//   Within a warm isolate, plans are cached in memory after the first KV read.
//   Subsequent requests on the same isolate pay 0ms. KV pays ~1ms.

import type { Env } from '../types.js'

export type PlanId = 'free' | 'dev' | 'dev_annual' | 'pro' | 'pro_annual' | 'enterprise'
export type BillingInterval = 'monthly' | 'annual'

export interface Plan {
  id: PlanId
  name: string
  tagline: string
  status: 'active' | 'hidden' | 'deprecated'
  audience: 'individual' | 'team' | 'enterprise'
  billingMode: 'free' | 'subscription' | 'enterprise'
  interval: BillingInterval | null

  stripe: {
    productId: string | null
    priceId: string | null  // monthly price (backward compat)
    prices?: {
      monthly?: string | null
      annual?: string | null
    }
    checkoutMode: 'subscription' | 'payment' | null
  }

  pricing: {
    amountCents: number
    currency: 'usd'
    displayPrice: string
  }

  limits: {
    cloudOptimizationsMonthly: number
    cloudTokensMonthly: number
    maxRawTokensPerOptimization: number
    maxPayloadBytes: number
    retainedHistoryDays: number
    cliSessions: number
  }

  credits: {
    includedMonthly: number  // monthly included credits (also used by seed trigger)
    annual?: number           // total credits for annual subscribers per period
    rollover: boolean
    allowManualGrants: boolean
    allowPurchases: boolean
  }

  features: {
    localOptimization: boolean
    remoteOptimization: boolean
    usageDashboard: boolean
    advancedUsageDashboard: boolean
    cliSessionManagement: boolean
    payloadCapture: boolean
    exportData: boolean
    priorityOptimizerUpdates: boolean
    customOptimizers: boolean
    ssoSaml: boolean
    optimizationHistory: boolean
    earlyAccessAdapters: boolean
    sessionMemoryGuard: boolean
    preCompactOptimizer: boolean
    localMemoryPacks: boolean
    // firewallGroupToggles: paid users can customise which MCP firewall rule groups are active.
    // Value comes from the `plans` table in Supabase — not hardcoded here.
    firewallGroupToggles: boolean
  }

  optimizers: {
    local:  string[]
    remote: string[]
  }

  telemetry: {
    requiredUsageMetering: boolean
    optionalProductAnalyticsDefault: boolean
  }
}

// ── KV key ────────────────────────────────────────────────────────────────────

const KV_KEY = 'plans_v1'

// ── Module-level cache (per isolate, ~0ms after first read) ──────────────────

let _cache: Record<string, Plan> | null = null

// ── Public API ────────────────────────────────────────────────────────────────

// getPlans returns all plans.
// Warm isolate:  ~0ms  (module memory)
// Cold isolate:  ~1ms  (KV hit)
// First ever:    ~50ms (KV miss → load from Supabase → write to KV → cache)
//
// Self-bootstrapping: no manual sync step, no deploy script, no throwable error.
// The Worker populates KV itself the first time it runs.
export async function getPlans(env: Env): Promise<Record<string, Plan>> {
  // Layer 1: module memory (per isolate)
  if (_cache) return _cache

  // Layer 2: KV (globally shared)
  const kv = await env.CACHE.get(KV_KEY, 'json') as Record<string, Plan> | null
  if (kv) {
    _cache = kv
    return _cache
  }

  // Layer 3: KV miss → load from Supabase, write to KV, then serve.
  // Only happens once globally (after that KV always has the data).
  await syncPlansToKV(env)
  return _cache!
}

// getPlan returns a single plan by id.
// Throws if the plan doesn't exist — callers should handle this.
export async function getPlan(planId: string, env: Env): Promise<Plan> {
  const plans = await getPlans(env)
  const plan = plans[planId]
  if (!plan) {
    // Unknown plan_id (e.g. user has a stale plan_id from a deleted plan).
    // Return free plan as the safe default.
    return plans['free']!
  }
  return plan
}

// syncPlansToKV reads all plans from Supabase and writes them to KV.
// Called by POST /admin/plans/sync and from the deploy script.
export async function syncPlansToKV(env: Env): Promise<void> {
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) {
    throw new Error('Supabase not configured')
  }

  const res = await fetch(
    `${env.SUPABASE_URL}/rest/v1/plans?select=id,config,status`,
    {
      headers: {
        apikey:        env.SUPABASE_SERVICE_KEY,
        Authorization: `Bearer ${env.SUPABASE_SERVICE_KEY}`,
      },
    }
  )

  if (!res.ok) throw new Error(`Supabase error: ${res.status}`)

  const rows = await res.json() as Array<{ id: string; config: Plan; status: string }>
  const plans: Record<string, Plan> = {}

  for (const row of rows) {
    plans[row.id] = { ...row.config, id: row.id as PlanId, status: row.status as Plan['status'] }
  }

  // Write to KV — no TTL, lives there until explicitly replaced via next sync
  await env.CACHE.put(KV_KEY, JSON.stringify(plans))

  // Bust the module cache on this isolate
  _cache = plans
}

// invalidateCache clears the module-level cache for this isolate.
// Other isolates will re-read from KV on their next request.
export function invalidatePlanCache(): void {
  _cache = null
}

// getPriceForInterval returns the Stripe price ID for the requested billing interval.
// Falls back to stripe.priceId (monthly) when prices object is absent.
export function getPriceForInterval(plan: Plan, interval: BillingInterval): string | null {
  return plan.stripe.prices?.[interval] ?? plan.stripe.priceId ?? null
}

// getCreditsForInterval returns how many credits to grant for a billing period.
export function getCreditsForInterval(plan: Plan, interval: BillingInterval): number {
  if (interval === 'annual' && plan.credits.annual != null) return plan.credits.annual
  return plan.credits.includedMonthly
}
