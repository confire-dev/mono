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

export type Plan = 'free' | 'dev' | 'dev_annual' | 'pro' | 'pro_annual' | 'enterprise'
export type SubscriptionStatus = 'none' | 'trialing' | 'active' | 'past_due' | 'canceled' | 'incomplete'

export interface Profile {
  id: string
  email: string
  name?: string
  plan: Plan
  plan_id: Plan
  billing_interval?: 'monthly' | 'annual'
  subscription_status: SubscriptionStatus
  stripe_customer_id?: string
  stripe_subscription_id?: string
  subscription_current_period_start?: string
  subscription_current_period_end?: string
  cancel_at_period_end: boolean
  is_banned: boolean
}

export interface BillingItem {
  id: string
  type: string
  config: {
    name: string
    creditsPerUnit: number
    maxQuantity: number
    creditType: string
    stripe: { productId: string; priceId: string }
  }
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
// Two-step: api_keys has FK to auth.users (not profiles), so we can't embed
// profiles(*) in one PostgREST query. Look up user_id first, then fetch profile.
// Also updates last_used_at asynchronously.
export async function validateApiKey(cfg: SupabaseConfig, rawKey: string): Promise<Profile | null> {
  const hash = await sha256(rawKey)

  const keyRes = await sbFetch(cfg, 'GET',
    `/rest/v1/api_keys?select=user_id&key_hash=eq.${encodeURIComponent(hash)}&revoked_at=is.null&limit=1`)
  if (!keyRes.ok) return null
  const keyRows = await keyRes.json() as Array<{ user_id: string }>
  if (!keyRows.length) return null

  const userId = keyRows[0]!.user_id

  // Async side-effect: update last_used_at. Don't block.
  sbFetch(cfg, 'PATCH', `/rest/v1/api_keys?key_hash=eq.${encodeURIComponent(hash)}`,
    { last_used_at: new Date().toISOString() }).catch(() => {})

  const profileRes = await sbFetch(cfg, 'GET',
    `/rest/v1/profiles?id=eq.${encodeURIComponent(userId)}&limit=1`)
  if (!profileRes.ok) return null
  const profiles = await profileRes.json() as Profile[]
  return profiles[0] ?? null
}

// ── Security event recording ───────────────────────────────────────────────

// getCreditBalance returns the user's credit pool (included + bonus + purchased).
export async function getCreditBalance(
  cfg: SupabaseConfig,
  userId: string,
): Promise<CreditBalance | null> {
  const res = await sbFetch(cfg, 'GET',
    `/rest/v1/credit_balances?user_id=eq.${encodeURIComponent(userId)}&select=user_id,included_credits,bonus_credits,purchased_credits,total_credits&limit=1`)
  if (!res.ok) return null
  const rows = await res.json() as CreditBalance[]
  return rows[0] ?? null
}

// consumePurchasedCredits deducts from the purchased pool only (top-up overflow usage).
// Returns false if insufficient purchased credits.
export async function consumePurchasedCredits(
  cfg: SupabaseConfig,
  userId: string,
  amount: number,
  sessionId?: string,
): Promise<boolean> {
  const res = await sbFetch(cfg, 'POST', '/rest/v1/rpc/consume_purchased_credits', {
    p_user_id:    userId,
    p_amount:     amount,
    p_session_id: sessionId ?? null,
  })
  if (!res.ok) return false
  return (await res.json()) as boolean
}

// recordSecurityEvent writes a security event to the security_events table.
export async function recordSecurityEvent(
  cfg: SupabaseConfig,
  params: {
    userId: string
    sessionId: string
    toolName: string
    eventType: string
    riskLevel: string
    actionTaken: string
    patternMatched?: string
    bypassed?: boolean
  }
): Promise<void> {
  await sbFetch(cfg, 'POST', '/rest/v1/security_events', {
    user_id:         params.userId,
    session_id:      params.sessionId || null,
    tool_name:       params.toolName,
    event_type:      params.eventType,
    risk_level:      params.riskLevel,
    action_taken:    params.actionTaken,
    pattern_matched: params.patternMatched ?? null,
    bypassed:        params.bypassed ?? false,
  })
}

// ── Top-up credits ────────────────────────────────────────────────────────────

// grantPurchasedCredits adds one-time purchased credits to a user's balance.
// Idempotent via stripe_event_id — safe to call on webhook retry.
export async function grantPurchasedCredits(
  cfg: SupabaseConfig,
  params: { userId: string; credits: number; stripeEventId: string },
): Promise<void> {
  await sbFetch(cfg, 'POST', '/rest/v1/rpc/grant_credits', {
    p_user_id:         params.userId,
    p_purchased:       params.credits,
    p_source:          'stripe_topup',
    p_stripe_event_id: params.stripeEventId,
  })
}

// ── Profile & user management ──────────────────────────────────────────────

// getProfileById fetches an existing profile by Supabase user ID. Returns null if not found.
export async function getProfileById(cfg: SupabaseConfig, userId: string): Promise<Profile | null> {
  const res = await sbFetch(cfg, 'GET', `/rest/v1/profiles?id=eq.${encodeURIComponent(userId)}&limit=1`)
  if (!res.ok) return null
  const rows = await res.json() as Profile[]
  return rows[0] ?? null
}

// upsertProfile creates or updates a profile from a Supabase auth user.
// Call this only on first login / key generation — not on every request.
// Only email (and name when provided) are merged on conflict — plan and
// subscription_status are authoritative from Stripe webhooks and must not
// be overwritten here (they have DB defaults for new rows).
// Only email (and name when provided) are merged on conflict — plan and
// subscription_status are authoritative from Stripe webhooks and must not
// be overwritten here (they have DB defaults for new rows).
export async function upsertProfile(
  cfg: SupabaseConfig,
  supabaseUserId: string,
  email: string,
  name?: string
): Promise<Profile> {
  const body: Record<string, unknown> = { id: supabaseUserId, email }
  if (name != null) body.name = name
  const res = await sbFetch(cfg, 'POST', '/rest/v1/profiles',
    body,
    { 'Prefer': 'resolution=merge-duplicates,return=representation' })
  const rows = await res.json() as Profile[]
  // Credit balance is seeded by the DB trigger on profiles INSERT.
  // No action needed here — ON CONFLICT DO NOTHING in the trigger prevents duplicates.
  return rows[0]!
}

// revokeApiKey sets revoked_at on a key owned by userId. Scoped to the owner so
// users can only revoke their own keys.
export async function revokeApiKey(cfg: SupabaseConfig, userId: string, keyId: string): Promise<void> {
  await sbFetch(cfg, 'PATCH',
    `/rest/v1/api_keys?id=eq.${encodeURIComponent(keyId)}&user_id=eq.${encodeURIComponent(userId)}&revoked_at=is.null`,
    { revoked_at: new Date().toISOString() })
}

// generateApiKey creates a new API key for a user, stores its hash, returns raw key.
export async function generateApiKey(
  cfg: SupabaseConfig,
  userId: string,
  deviceId?: string,
  deviceName?: string,
): Promise<string> {
  const rawKey = `cf_live_${secureRandom(32)}`
  const hash   = await sha256(rawKey)

  // Try with key_suffix first; fall back without it if the column doesn't exist yet.
  let res = await sbFetch(cfg, 'POST', '/rest/v1/api_keys',
    {
      user_id:    userId,
      key_hash:   hash,
      key_prefix: rawKey.slice(0, 12),
      key_suffix: rawKey.slice(-4),
      device_id:  deviceName ?? deviceId ?? null,
    })

  if (!res.ok) {
    const body = await res.text()
    // column "key_suffix" of relation "api_keys" does not exist
    if (body.includes('key_suffix')) {
      res = await sbFetch(cfg, 'POST', '/rest/v1/api_keys',
        {
          user_id:    userId,
          key_hash:   hash,
          key_prefix: rawKey.slice(0, 12),
          device_id:  deviceName ?? deviceId ?? null,
        })
    }
    if (!res.ok) {
      const msg = res.ok ? '' : await res.text().catch(() => res.statusText)
      throw new Error(`api_keys insert failed (${res.status}): ${msg}`)
    }
  }

  await writeAudit(cfg, userId, 'api_key_generated', { prefix: rawKey.slice(0, 12), device_name: deviceName, device_id: deviceId })
  return rawKey
}

// revokeCurrentKey revokes the key identified by its hash — used by the CLI on logout.
export async function revokeCurrentKey(cfg: SupabaseConfig, rawKey: string): Promise<void> {
  const hash = await sha256(rawKey)
  await sbFetch(cfg, 'PATCH',
    `/rest/v1/api_keys?key_hash=eq.${encodeURIComponent(hash)}&revoked_at=is.null`,
    { revoked_at: new Date().toISOString() })
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
      ended_at:    null,  // clear ended_at so a restarted session shows Active again
    },
    { 'Prefer': 'resolution=merge-duplicates' })
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
// Credits are granted on invoice.paid — not here — to avoid double-granting.
export async function handleSubscriptionUpdated(
  cfg: SupabaseConfig,
  event: {
    stripeEventId: string
    stripeCustomerId: string
    subscriptionId: string
    status: string
    planId: string
    billingInterval: 'monthly' | 'annual'
    currentPeriodStart: number
    currentPeriodEnd: number
    cancelAtPeriodEnd?: boolean
  }
): Promise<void> {
  const status = mapStripeStatus(event.status)

  await sbFetch(cfg, 'PATCH',
    `/rest/v1/profiles?stripe_customer_id=eq.${encodeURIComponent(event.stripeCustomerId)}`, {
      stripe_subscription_id:            event.subscriptionId,
      plan:                              event.planId,
      plan_id:                           event.planId,
      billing_interval:                  event.billingInterval,
      subscription_status:               status,
      subscription_current_period_start: new Date(event.currentPeriodStart * 1000).toISOString(),
      subscription_current_period_end:   new Date(event.currentPeriodEnd * 1000).toISOString(),
      cancel_at_period_end:              event.cancelAtPeriodEnd ?? false,
      updated_at:                        new Date().toISOString(),
    })
}

// Called on invoice.paid — grants included credits for the billing period.
// Idempotency key = Stripe invoice ID (safe on webhook retry).
export async function handleInvoicePaid(
  cfg: SupabaseConfig,
  event: {
    stripeCustomerId: string
    stripeInvoiceId: string
    planId: string
    billingInterval: 'monthly' | 'annual'
  }
): Promise<void> {
  const profileRes = await sbFetch(cfg, 'GET',
    `/rest/v1/profiles?stripe_customer_id=eq.${encodeURIComponent(event.stripeCustomerId)}&select=id&limit=1`)
  if (!profileRes.ok) return
  const profiles = await profileRes.json() as Array<{ id: string }>
  if (!profiles.length) return
  const userId = profiles[0]!.id

  // Resolve credits from plans table
  const plansRes = await sbFetch(cfg, 'GET',
    `/rest/v1/plans?id=eq.${encodeURIComponent(event.planId)}&select=config&limit=1`)
  const planRows = plansRes.ok
    ? await plansRes.json() as Array<{ config: { credits: { includedMonthly: number; annual?: number } } }>
    : []
  const credits = event.billingInterval === 'annual'
    ? (planRows[0]?.config?.credits?.annual ?? planRows[0]?.config?.credits?.includedMonthly ?? 500)
    : (planRows[0]?.config?.credits?.includedMonthly ?? 500)

  await sbFetch(cfg, 'POST', '/rest/v1/rpc/grant_credits', {
    p_user_id:         userId,
    p_included:        credits,
    p_source:          'subscription',
    p_stripe_event_id: event.stripeInvoiceId,
  })
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

// ── Billing items ─────────────────────────────────────────────────────────────

export async function fetchBillingItem(cfg: SupabaseConfig, id: string): Promise<BillingItem | null> {
  const res = await sbFetch(cfg, 'GET',
    `/rest/v1/billing_items?id=eq.${encodeURIComponent(id)}&active=eq.true&limit=1`)
  if (!res.ok) return null
  const rows = await res.json() as BillingItem[]
  return rows[0] ?? null
}

// markWebhookEventProcessed inserts the event ID into stripe_webhook_events.
// Returns true if this event is new (should be processed).
// Returns false if already processed (should be skipped — idempotent).
export async function markWebhookEventProcessed(
  cfg: SupabaseConfig,
  eventId: string,
  eventType: string,
): Promise<boolean> {
  const res = await sbFetch(cfg, 'POST', '/rest/v1/stripe_webhook_events',
    { id: eventId, type: eventType },
    { 'Prefer': 'resolution=ignore-duplicates,return=representation' })
  if (!res.ok) return true // if DB is down, process anyway (fail open)
  const rows = await res.json() as unknown[]
  return rows.length > 0
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
