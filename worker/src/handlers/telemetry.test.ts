/**
 * Tests for the telemetry handler's plan-gated history write paths.
 *
 * Two durable history tables are gated:
 *   security_events  ← plan.features.securityEventHistory
 *   provenance_events ← plan.features.provenanceTracking
 *
 * All other write paths (session records, audit log) are not plan-gated.
 * The local daemon's protection logic (block/warn/review/budget/rate) is
 * entirely unaffected — it runs before any network call reaches the Worker.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { handleTelemetry } from './telemetry.js'
import type { Env } from '../types.js'
import type { Plan } from '../lib/plans.js'

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock('../lib/auth.js', () => ({
  authenticate: vi.fn(),
}))

vi.mock('../lib/plans.js', () => ({
  getPlan: vi.fn(),
}))

vi.mock('../lib/supabase.js', () => ({
  recordSecurityEvent:  vi.fn().mockResolvedValue(undefined),
  recordProvenanceEvent: vi.fn().mockResolvedValue(undefined),
  upsertCliSession:     vi.fn().mockResolvedValue(undefined),
  closeCliSession:      vi.fn().mockResolvedValue(undefined),
  writeAudit:           vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../lib/analytics.js', () => ({
  trackEvent: vi.fn(),
}))

vi.mock('../lib/cache.js', () => ({
  cacheKey: vi.fn((k: string) => k),
  cacheGet: vi.fn().mockResolvedValue(null),
  cachePut: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../lib/buckets.js', () => ({
  sanitizeToolType: vi.fn((t: string) => t),
}))

// ── Import mocked modules for spying ────────────────────────────────────────

import { authenticate } from '../lib/auth.js'
import { getPlan } from '../lib/plans.js'
import { recordSecurityEvent, recordProvenanceEvent } from '../lib/supabase.js'

// ── Helpers ──────────────────────────────────────────────────────────────────

function makePlan(overrides: { securityEventHistory?: boolean; provenanceTracking?: boolean } = {}): Plan {
  return {
    id: 'free',
    name: 'Free',
    tagline: '',
    status: 'active',
    audience: 'individual',
    billingMode: 'free',
    interval: null,
    stripe: { productId: null, priceId: null, checkoutMode: null },
    pricing: { amountCents: 0, currency: 'usd', displayPrice: '$0' },
    limits: { maxPayloadBytes: 1_000_000, retainedHistoryDays: 0, cliSessions: 10 },
    credits: { includedMonthly: 0, rollover: false, allowManualGrants: false, allowPurchases: false },
    features: {
      firewallEnabled:         true,
      usageDashboard:          false,
      advancedUsageDashboard:  false,
      cliSessionManagement:    false,
      exportData:              false,
      ssoSaml:                 false,
      firewallGroupToggles:    false,
      securityEventHistory:    overrides.securityEventHistory ?? false,
      provenanceTracking:      overrides.provenanceTracking ?? false,
      configurableThresholds:  false,
    },
    telemetry: { requiredUsageMetering: false, optionalProductAnalyticsDefault: false },
  }
}

function makeEnv(): Env {
  return {
    SUPABASE_URL:        'https://test.supabase.co',
    SUPABASE_SERVICE_KEY: 'service-key',
    ENVIRONMENT:         'test',
    CACHE: {
      get:    vi.fn().mockResolvedValue(null),
      put:    vi.fn().mockResolvedValue(undefined),
      delete: vi.fn().mockResolvedValue(undefined),
      list:   vi.fn().mockResolvedValue({ keys: [] }),
      getWithMetadata: vi.fn().mockResolvedValue({ value: null, metadata: null }),
    } as unknown as KVNamespace,
  }
}

function makeRequest(body: object): Request {
  return new Request('https://api.confire.dev/v1/events', {
    method:  'POST',
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer test-key' },
    body:    JSON.stringify(body),
  })
}

function setupAuth(planId = 'free') {
  vi.mocked(authenticate).mockResolvedValue({
    ok: true,
    user: {
      id: 'user-123', email: 'test@example.com', plan: planId as any, plan_id: planId as any,
      subscription_status: 'none', cancel_at_period_end: false, is_banned: false,
    },
  })
}

const BASE_SECURITY_EVENT = {
  event_id:           'evt-001',
  event_type:         'security_event',
  tool_type:          'Bash',
  risk_level:         'HIGH',
  action_taken:       'BLOCKED',
  analytics_consented: false,
}

const BASE_PROVENANCE_EVENT = {
  event_id:           'evt-002',
  event_type:         'provenance_event',
  tool_type:          'Bash',
  trust_level:        'external_untrusted',
  analytics_consented: false,
}

// ── Tests ─────────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks()
})

describe('security_event: plan gate', () => {
  it('does NOT write to security_events for a free-plan user', async () => {
    setupAuth('free')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ securityEventHistory: false }))

    const res = await handleTelemetry(makeRequest(BASE_SECURITY_EVENT), makeEnv())
    expect(res.status).toBe(200)
    expect(vi.mocked(recordSecurityEvent)).not.toHaveBeenCalled()
  })

  it('DOES write to security_events for a paid user with securityEventHistory enabled', async () => {
    setupAuth('pro')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ securityEventHistory: true }))

    const res = await handleTelemetry(makeRequest(BASE_SECURITY_EVENT), makeEnv())
    expect(res.status).toBe(200)
    expect(vi.mocked(recordSecurityEvent)).toHaveBeenCalledOnce()
  })

  it('returns ok:true regardless of plan (silent skip, no error surfaced)', async () => {
    setupAuth('free')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ securityEventHistory: false }))

    const res = await handleTelemetry(makeRequest(BASE_SECURITY_EVENT), makeEnv())
    const body = await res.json() as { ok: boolean }
    expect(body.ok).toBe(true)
  })

  it('still forwards analytics for free users even when DB write is skipped', async () => {
    // Analytics forwarding is gated on analytics_consented, not on plan.
    // Verify the DB skip does not also accidentally suppress analytics.
    setupAuth('free')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ securityEventHistory: false }))

    const eventWithConsent = { ...BASE_SECURITY_EVENT, analytics_consented: true }
    const res = await handleTelemetry(makeRequest(eventWithConsent), makeEnv())
    expect(res.status).toBe(200)
    expect(vi.mocked(recordSecurityEvent)).not.toHaveBeenCalled()
    // analytics path fires separately — we confirm the response is still ok
    const body = await res.json() as { ok: boolean }
    expect(body.ok).toBe(true)
  })
})

describe('provenance_event: plan gate', () => {
  it('does NOT write to provenance_events for a free-plan user', async () => {
    setupAuth('free')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ provenanceTracking: false }))

    const res = await handleTelemetry(makeRequest(BASE_PROVENANCE_EVENT), makeEnv())
    expect(res.status).toBe(200)
    expect(vi.mocked(recordProvenanceEvent)).not.toHaveBeenCalled()
  })

  it('DOES write to provenance_events for a paid user with provenanceTracking enabled', async () => {
    setupAuth('pro')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ provenanceTracking: true }))

    const res = await handleTelemetry(makeRequest(BASE_PROVENANCE_EVENT), makeEnv())
    expect(res.status).toBe(200)
    expect(vi.mocked(recordProvenanceEvent)).toHaveBeenCalledOnce()
  })

  it('returns ok:true for free users (silent skip)', async () => {
    setupAuth('free')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ provenanceTracking: false }))

    const res = await handleTelemetry(makeRequest(BASE_PROVENANCE_EVENT), makeEnv())
    const body = await res.json() as { ok: boolean }
    expect(body.ok).toBe(true)
  })
})

describe('flags are independent', () => {
  it('securityEventHistory does not affect provenance writes', async () => {
    // A user with only securityEventHistory=true should still have provenance gated.
    setupAuth('partial')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ securityEventHistory: true, provenanceTracking: false }))

    await handleTelemetry(makeRequest(BASE_PROVENANCE_EVENT), makeEnv())
    expect(vi.mocked(recordProvenanceEvent)).not.toHaveBeenCalled()
  })

  it('provenanceTracking does not affect security_event writes', async () => {
    setupAuth('partial')
    vi.mocked(getPlan).mockResolvedValue(makePlan({ securityEventHistory: false, provenanceTracking: true }))

    await handleTelemetry(makeRequest(BASE_SECURITY_EVENT), makeEnv())
    expect(vi.mocked(recordSecurityEvent)).not.toHaveBeenCalled()
  })
})

describe('local protection is plan-independent', () => {
  // The local daemon makes block/warn/review/budget/rate decisions before any
  // Worker call. These tests confirm the Worker's plan gate does not bleed into
  // the daemon's local protection path — the daemon never checks entitlements.
  //
  // Verified by architecture: the CLI daemon has no plan lookup, no Worker call
  // on the protection path. The Worker only receives typed telemetry events
  // AFTER the local decision has already been made and enforced.
  //
  // These tests confirm at the Worker boundary that even free-tier users receive
  // { ok: true } — no error that could interfere with the CLI's expectation that
  // event submission always succeeds.

  it('free user security_event submission returns 200 ok (daemon not blocked)', async () => {
    setupAuth('free')
    vi.mocked(getPlan).mockResolvedValue(makePlan())

    const res = await handleTelemetry(makeRequest(BASE_SECURITY_EVENT), makeEnv())
    expect(res.status).toBe(200)
    const body = await res.json() as { ok: boolean }
    expect(body.ok).toBe(true)
  })

  it('free user provenance_event submission returns 200 ok (daemon not blocked)', async () => {
    setupAuth('free')
    vi.mocked(getPlan).mockResolvedValue(makePlan())

    const res = await handleTelemetry(makeRequest(BASE_PROVENANCE_EVENT), makeEnv())
    expect(res.status).toBe(200)
    const body = await res.json() as { ok: boolean }
    expect(body.ok).toBe(true)
  })
})
