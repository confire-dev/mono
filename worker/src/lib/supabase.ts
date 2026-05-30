// Supabase client helpers — Cloudflare Worker side.
// Uses the Supabase REST API directly (no npm SDK — stays V8-safe).
// Worker always uses the SERVICE ROLE KEY, which bypasses RLS.
// Never expose the service role key to the CLI.
//
// Supabase is the source of truth for: usage, credits, billing state,
// sessions, tool call summaries, audit logs, and the user dashboard.
// Amplitude is for behavioral analytics only — never for billing decisions.

export interface SupabaseConfig {
  url: string         // SUPABASE_URL
  serviceKey: string  // SUPABASE_SERVICE_KEY (bypasses RLS)
}

// ── Domain types ───────────────────────────────────────────────────────────

export type Plan = 'free' | 'pro'
export type SubscriptionStatus = 'none' | 'trialing' | 'active' | 'past_due' | 'canceled' | 'incomplete'

export interface Profile {
  id: string
  email: string
  name?: string
  plan: Plan
  subscription_status: SubscriptionStatus
  stripe_customer_id?: string
  stripe_subscription_id?: string
  subscription_current_period_end?: string
  is_banned: boolean
}

export interface CreditBalance {
  user_id: string
  included_credits: number
  bonus_credits: number
  purchased_credits: number
  total_credits: number
}

export interface UsageInfo {
  ok: boolean          // within limit
  used: number         // requests this month
  limit: number        // plan limit
  plan: Plan
  total_credits: number
}

export interface ApiKey {
  id: string
  user_id: string
  key_hash: string
  key_prefix: string
  device_id?: string
}

// Limits are now defined in plans.ts — no hardcoded numbers here.
// MAX_RAW_PAYLOAD_BYTES kept as a Worker-level hard cap (before plan lookup).
export const MAX_RAW_PAYLOAD_BYTES = 50 * 1024 * 1024 // 50MB absolute max before plan check

// ── Auth / key validation ──────────────────────────────────────────────────

// validateApiKey looks up a key by its SHA-256 hash and returns the profile.
// Also updates last_used_at asynchronously.
export async function validateApiKey(cfg: SupabaseConfig, rawKey: string): Promise<Profile | null> {
  const hash = await sha256(rawKey)
  const res = await sbFetch(cfg, 'GET',
    `/rest/v1/api_keys?select=user_id,profiles(*)&key_hash=eq.${encodeURIComponent(hash)}&revoked_at=is.null&limit=1`)
  if (!res.ok) return null
  const rows = await res.json() as Array<{ profiles: Profile }>
  if (!rows.length || !rows[0]?.profiles) return null

  // Async side-effect: update last_used_at. Don't block.
  sbFetch(cfg, 'PATCH', `/rest/v1/api_keys?key_hash=eq.${encodeURIComponent(hash)}`,
    { last_used_at: new Date().toISOString() }).catch(() => {})

  return rows[0].profiles
}

// ── Usage & credits ────────────────────────────────────────────────────────

// UsagePeriod is the current period's usage numbers, from the usage_periods table.
export interface UsagePeriod {
  id: string
  cloudOptimizationsUsed: number
  cloudTokensUsed: number
  localOptimizationsCount: number
  savedTokens: number
}

// getUsageThisPeriod fetches the current active usage_period for the user.
// If no period exists for the current month, creates one.
// Replaces the old checkUsage / monthly_usage approach.
// The caller uses canUseRemoteOptimizer() from entitlement.ts to decide if allowed.
export async function getUsageThisPeriod(
  cfg: SupabaseConfig,
  userId: string,
  planId: string,
): Promise<UsagePeriod> {
  // Call the DB stored procedure which gets-or-creates the current period
  const res = await sbFetch(cfg, 'POST', '/rest/v1/rpc/get_or_create_active_period', {
    p_user_id: userId,
    p_plan_id: planId,
  })

  if (!res.ok) {
    // Fallback: return zeros (allow the call, don't block on DB error)
    return { id: '', cloudOptimizationsUsed: 0, cloudTokensUsed: 0, localOptimizationsCount: 0, savedTokens: 0 }
  }

  const periodId = (await res.json()) as string

  const periodRes = await sbFetch(cfg, 'GET',
    `/rest/v1/usage_periods?id=eq.${periodId}&select=id,cloud_optimizations_used,cloud_tokens_used,local_optimizations_count,saved_tokens&limit=1`)

  if (!periodRes.ok) {
    return { id: periodId, cloudOptimizationsUsed: 0, cloudTokensUsed: 0, localOptimizationsCount: 0, savedTokens: 0 }
  }

  const rows = await periodRes.json() as Array<{
    id: string
    cloud_optimizations_used: number
    cloud_tokens_used: number
    local_optimizations_count: number
    saved_tokens: number
  }>

  const r = rows[0]
  if (!r) return { id: periodId, cloudOptimizationsUsed: 0, cloudTokensUsed: 0, localOptimizationsCount: 0, savedTokens: 0 }

  return {
    id:                      r.id,
    cloudOptimizationsUsed:  r.cloud_optimizations_used,
    cloudTokensUsed:         r.cloud_tokens_used,
    localOptimizationsCount: r.local_optimizations_count,
    savedTokens:             r.saved_tokens,
  }
}

// Legacy UsageInfo — keep for auth.ts compatibility (will be removed next pass)
export interface UsageInfo {
  ok: boolean
  used: number
  limit: number
  plan: Plan
  total_credits: number
}

// recordOptimization increments the active usage_period and writes a tool call summary.
export async function recordOptimization(
  cfg: SupabaseConfig,
  params: {
    userId: string
    planId: string
    sessionId: string
    toolType: string
    integration: string
    optimizer: string
    rawBytes: number
    optimizedBytes: number
    durationMs?: number
    wasCached: boolean
    analyticsConsented: boolean
  }
): Promise<void> {
  const bytesSaved  = params.rawBytes - params.optimizedBytes
  const rawTokens   = Math.round(params.rawBytes / 4)   // ~4 bytes per token
  const savedTokens = Math.round(bytesSaved / 4)

  // 1. Increment the active usage_period (new approach — replaces monthly_usage)
  await sbFetch(cfg, 'POST', '/rest/v1/rpc/increment_usage_period', {
    p_user_id:          params.userId,
    p_plan_id:          params.planId,
    p_cloud_opts:       1,
    p_cloud_tokens:     rawTokens,
    p_saved_tokens:     savedTokens,
  })

  // 2. Write per-call detail (powers the dashboard)
  await sbFetch(cfg, 'POST', '/rest/v1/tool_call_summaries', {
    user_id:         params.userId,
    cli_session_id:  params.sessionId || null,
    tool_type:       params.toolType,
    integration:     params.integration,
    optimizer:       params.optimizer,
    raw_bytes:       params.rawBytes,
    optimized_bytes: params.optimizedBytes,
    was_cached:      params.wasCached,
    credits_used:    1,
  })
}

// ── Profile & user management ──────────────────────────────────────────────

// upsertProfile creates or updates a profile from a Supabase auth user.
// Called after successful OAuth → API key generation.
export async function upsertProfile(
  cfg: SupabaseConfig,
  supabaseUserId: string,
  email: string,
  name?: string
): Promise<Profile> {
  const res = await sbFetch(cfg, 'POST', '/rest/v1/profiles',
    { id: supabaseUserId, email, name: name ?? null, plan: 'free', subscription_status: 'none' },
    { 'Prefer': 'resolution=merge-duplicates,return=representation' })
  const rows = await res.json() as Profile[]
  // Seed credit balance from the plans table (DB is source of truth).
  const planRow = await sbFetch(cfg, 'GET',
    `/rest/v1/plans?id=eq.free&select=config&limit=1`)
  const planData = planRow.ok
    ? (await planRow.json() as Array<{ config: { credits: { includedMonthly: number } } }>)[0]
    : null
  const includedCredits = planData?.config?.credits?.includedMonthly ?? 500

  await sbFetch(cfg, 'POST', '/rest/v1/rpc/grant_credits', {
    p_user_id:  supabaseUserId,
    p_included: includedCredits,
    p_source:   'system',
  })
  return rows[0]!
}

// generateApiKey creates a new API key for a user, stores its hash, returns raw key.
export async function generateApiKey(
  cfg: SupabaseConfig,
  userId: string,
  deviceId?: string
): Promise<string> {
  const rawKey = `cf_live_${secureRandom(32)}`
  const hash   = await sha256(rawKey)
  await sbFetch(cfg, 'POST', '/rest/v1/api_keys',
    { user_id: userId, key_hash: hash, key_prefix: rawKey.slice(0, 12), device_id: deviceId ?? null })
  await writeAudit(cfg, userId, 'api_key_generated', { prefix: rawKey.slice(0, 12), device_id: deviceId })
  return rawKey
}

// ── Sessions ───────────────────────────────────────────────────────────────

export async function upsertCliSession(
  cfg: SupabaseConfig,
  sessionId: string,
  userId: string,
  params: { deviceId?: string; cliVersion?: string; integration?: string }
): Promise<void> {
  await sbFetch(cfg, 'POST', '/rest/v1/cli_sessions',
    {
      id:          sessionId,
      user_id:     userId,
      device_id:   params.deviceId ?? null,
      cli_version: params.cliVersion ?? null,
      integration: params.integration ?? 'claude_code',
    },
    { 'Prefer': 'resolution=ignore-duplicates' })
}

export async function closeCliSession(
  cfg: SupabaseConfig,
  sessionId: string,
  stats: { totalCalls: number; optimizedCalls: number; rawBytes: number; optimizedBytes: number }
): Promise<void> {
  await sbFetch(cfg, 'PATCH', `/rest/v1/cli_sessions?id=eq.${encodeURIComponent(sessionId)}`, {
    ended_at:        new Date().toISOString(),
    total_tool_calls: stats.totalCalls,
    optimized_calls: stats.optimizedCalls,
    raw_bytes:       stats.rawBytes,
    optimized_bytes: stats.optimizedBytes,
  })
}

// ── Stripe webhook handlers ────────────────────────────────────────────────

// Called when a subscription is created or updated.
export async function handleSubscriptionUpdated(
  cfg: SupabaseConfig,
  event: {
    stripeEventId: string
    stripeCustomerId: string
    subscriptionId: string
    status: string
    plan: string
    currentPeriodEnd: number
  }
): Promise<void> {
  // Map Stripe status to our subscription_status
  const status = mapStripeStatus(event.status)
  const plan   = mapStripePlan(event.plan)

  // Update profile billing state
  await sbFetch(cfg, 'PATCH',
    `/rest/v1/profiles?stripe_customer_id=eq.${encodeURIComponent(event.stripeCustomerId)}`, {
      stripe_subscription_id:          event.subscriptionId,
      plan,
      subscription_status:             status,
      subscription_current_period_end: new Date(event.currentPeriodEnd * 1000).toISOString(),
      updated_at:                      new Date().toISOString(),
    })

  // Grant included credits for the new period (idempotent via stripe_event_id)
  const profileRes = await sbFetch(cfg, 'GET',
    `/rest/v1/profiles?stripe_customer_id=eq.${encodeURIComponent(event.stripeCustomerId)}&select=id&limit=1`)
  const profiles = await profileRes.json() as Array<{ id: string }>
  if (!profiles.length) return

  if (status === 'active' || status === 'trialing') {
    // Stripe webhook has no access to KV (env not available here).
    // Query the plans table directly for the includedMonthly value.
    const plansRes = await sbFetch(cfg, 'GET',
      `/rest/v1/plans?id=eq.${encodeURIComponent(plan)}&select=config&limit=1`)
    const planRows = plansRes.ok
      ? await plansRes.json() as Array<{ config: { credits: { includedMonthly: number } } }>
      : []
    const includedCredits = planRows[0]?.config?.credits?.includedMonthly ?? 500

    await sbFetch(cfg, 'POST', '/rest/v1/rpc/grant_credits', {
      p_user_id:          profiles[0]!.id,
      p_included:         includedCredits,
      p_source:           'subscription',
      p_stripe_event_id:  event.stripeEventId,
    })
  }
}

export async function handleSubscriptionCanceled(
  cfg: SupabaseConfig,
  stripeCustomerId: string
): Promise<void> {
  await sbFetch(cfg, 'PATCH',
    `/rest/v1/profiles?stripe_customer_id=eq.${encodeURIComponent(stripeCustomerId)}`, {
      plan:                'free',
      subscription_status: 'canceled',
      updated_at:          new Date().toISOString(),
    })
}

// ── Promotions ─────────────────────────────────────────────────────────────

export interface Promotion {
  id: string
  code: string
  discount_type: 'percent' | 'fixed_credits' | 'trial_days'
  discount_value: number
  stripe_coupon_id?: string
  max_redemptions?: number
  redemption_count: number
  valid_until?: string
  first_period_only: boolean
}

export async function getPromotion(cfg: SupabaseConfig, code: string): Promise<Promotion | null> {
  const res = await sbFetch(cfg, 'GET',
    `/rest/v1/promotions?code=eq.${encodeURIComponent(code)}&limit=1`)
  if (!res.ok) return null
  const rows = await res.json() as Promotion[]
  if (!rows.length) return null

  const promo = rows[0]!
  // Check expiry and redemption limit
  if (promo.valid_until && new Date(promo.valid_until) < new Date()) return null
  if (promo.max_redemptions != null && promo.redemption_count >= promo.max_redemptions) return null
  return promo
}

export async function redeemPromotion(
  cfg: SupabaseConfig,
  userId: string,
  promotionId: string
): Promise<boolean> {
  // Atomic: insert redemption (UNIQUE constraint prevents duplicates)
  const res = await sbFetch(cfg, 'POST', '/rest/v1/promotion_redemptions',
    { user_id: userId, promotion_id: promotionId },
    { 'Prefer': 'resolution=ignore-duplicates,return=representation' })
  if (!res.ok) return false
  const rows = await res.json() as unknown[]
  if (!rows.length) return false // already redeemed

  // Increment usage count on the promotion
  await sbFetch(cfg, 'POST', '/rest/v1/rpc/rpc_increment_promo_count',
    { p_promotion_id: promotionId }).catch(() => {})
  return true
}

// ── Audit ──────────────────────────────────────────────────────────────────

export async function writeAudit(
  cfg: SupabaseConfig,
  userId: string | null,
  eventType: string,
  metadata: Record<string, unknown> = {}
): Promise<void> {
  await sbFetch(cfg, 'POST', '/rest/v1/audit_events', {
    user_id:    userId,
    event_type: eventType,
    metadata,
  }).catch(() => {}) // audit failures must never block the main flow
}

// ── Helpers ────────────────────────────────────────────────────────────────

function mapStripeStatus(status: string): SubscriptionStatus {
  const map: Record<string, SubscriptionStatus> = {
    active: 'active', trialing: 'trialing',
    past_due: 'past_due', canceled: 'canceled',
    incomplete: 'incomplete', unpaid: 'past_due',
  }
  return map[status] ?? 'none'
}

function mapStripePlan(priceId: string): Plan {
  // Map Stripe price IDs to plan names.
  // Set STRIPE_PRO_PRICE_ID env var to your actual price ID.
  if (priceId.includes('pro')) return 'pro'
  return 'free'
}

async function sbFetch(
  cfg: SupabaseConfig,
  method: string,
  path: string,
  body?: unknown,
  extraHeaders?: Record<string, string>
): Promise<Response> {
  return fetch(`${cfg.url}${path}`, {
    method,
    headers: {
      apikey:          cfg.serviceKey,
      Authorization:   `Bearer ${cfg.serviceKey}`,
      'Content-Type':  'application/json',
      ...extraHeaders,
    },
    ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
  })
}

export async function sha256(input: string): Promise<string> {
  const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(input))
  return Array.from(new Uint8Array(buf)).map(b => b.toString(16).padStart(2, '0')).join('')
}

function secureRandom(bytes: number): string {
  const arr = new Uint8Array(bytes)
  crypto.getRandomValues(arr)
  return Array.from(arr).map(b => b.toString(16).padStart(2, '0')).join('')
}

export function monthKey(): string {
  const d = new Date()
  return `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}`
}
