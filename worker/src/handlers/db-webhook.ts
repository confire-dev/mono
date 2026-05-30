// Supabase Database Webhook handler — POST /webhooks/supabase
//
// Configure in Supabase dashboard:
//   Database → Webhooks → Create a new hook
//   Table:  plans
//   Events: INSERT, UPDATE, DELETE
//   URL:    https://api.confire.dev/webhooks/supabase
//   HTTP Headers:
//     Authorization: Bearer <SUPABASE_WEBHOOK_SECRET>
//
// When any row in the `plans` table changes, Supabase calls this endpoint.
// The Worker re-syncs all plans from DB → KV so every isolate picks up
// the change on its next request.
//
// This makes `plans` the genuine source of truth:
//   UPDATE plans SET config = ... in Supabase dashboard
//   → webhook fires to Worker (~<1s)
//   → Worker re-syncs plans to KV
//   → all Workers see the new plan on their next KV read (~1ms)

import type { Env } from '../types.js'
import { syncPlansToKV, invalidatePlanCache } from '../lib/plans.js'

// Supabase database webhook payload
interface SupabaseWebhookPayload {
  type:       'INSERT' | 'UPDATE' | 'DELETE'
  table:      string
  schema:     string
  record:     Record<string, unknown> | null
  old_record: Record<string, unknown> | null
}

export async function handleSupabaseWebhook(request: Request, env: Env): Promise<Response> {
  // ── Validate the webhook secret ─────────────────────────────────────────
  // Set SUPABASE_WEBHOOK_SECRET in wrangler secrets.
  // Supabase sends it as: Authorization: Bearer <secret>
  if (env.SUPABASE_WEBHOOK_SECRET) {
    const auth = request.headers.get('Authorization') ?? ''
    if (auth !== `Bearer ${env.SUPABASE_WEBHOOK_SECRET}`) {
      return Response.json({ error: 'unauthorized' }, { status: 401 })
    }
  }

  let payload: SupabaseWebhookPayload
  try {
    payload = await request.json() as SupabaseWebhookPayload
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }

  // Only react to changes on the `plans` table
  if (payload.table !== 'plans') {
    return Response.json({ ok: true, ignored: true })
  }

  // Re-sync all plans from Supabase → KV
  // (We sync all rather than patching one row to keep KV consistent.)
  try {
    invalidatePlanCache()           // bust this isolate's module cache
    await syncPlansToKV(env)        // re-read DB → write KV → re-populate cache
    return Response.json({
      ok: true,
      event: payload.type,
      plan_id: payload.record?.['id'] ?? payload.old_record?.['id'],
    })
  } catch (e) {
    // Log but don't 5xx — Supabase will retry on failure which could cause loops
    console.error('[confire] plan sync from webhook failed:', e)
    return Response.json({ ok: false, error: String(e) }, { status: 200 })
  }
}
