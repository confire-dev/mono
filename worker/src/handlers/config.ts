// POST /v1/config/thresholds — update rate-policy thresholds for a paid user.
//
// Free users enforce the hardcoded loose defaults in the daemon (20 loops, 100 calls/min).
// Paid users can tighten these thresholds from the dashboard or CLI. Gated behind
// the `configurableThresholds` feature flag.
//
// Body: { runaway_loop_threshold?: number, call_rate_threshold?: number }
// Response: { thresholds: ThresholdConfig }

import type { Env } from '../types.js'
import { authenticate } from '../lib/auth.js'
import { getPlan } from '../lib/plans.js'
import { upsertUserThresholds } from '../lib/supabase.js'

export interface ThresholdConfig {
  runaway_loop_threshold?: number
  call_rate_threshold?: number
}

export async function handlePatchThresholds(request: Request, env: Env): Promise<Response> {
  const auth = await authenticate(request, env)
  if (!auth.ok) return Response.json({ error: auth.error }, { status: auth.status })

  // Gate: configurableThresholds is a paid-plan feature.
  // Block = free (enforced in daemon), tightening = paid (enforced here).
  const plan = await getPlan(auth.user.plan_id, env)
  if (!plan.features.configurableThresholds) {
    return Response.json(
      { error: 'plan_limit', message: 'Configurable rate thresholds require a paid plan.' },
      { status: 403 },
    )
  }

  let body: ThresholdConfig
  try {
    body = await request.json() as ThresholdConfig
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }

  // Validate: each provided value must be a positive integer.
  const errors: string[] = []
  if (body.runaway_loop_threshold !== undefined) {
    if (!Number.isInteger(body.runaway_loop_threshold) || body.runaway_loop_threshold < 1) {
      errors.push('runaway_loop_threshold must be a positive integer')
    }
  }
  if (body.call_rate_threshold !== undefined) {
    if (!Number.isInteger(body.call_rate_threshold) || body.call_rate_threshold < 1) {
      errors.push('call_rate_threshold must be a positive integer')
    }
  }
  if (errors.length > 0) {
    return Response.json({ error: 'validation_error', messages: errors }, { status: 400 })
  }

  if (body.runaway_loop_threshold === undefined && body.call_rate_threshold === undefined) {
    return Response.json({ error: 'no fields to update' }, { status: 400 })
  }

  await upsertUserThresholds(
    { url: env.SUPABASE_URL ?? '', serviceKey: env.SUPABASE_SERVICE_KEY ?? '' },
    auth.user.id,
    body,
  )

  return Response.json({ thresholds: body })
}
