// GET  /v1/policy         — returns current user's firewall group overrides
// PATCH /v1/policy/groups — updates group overrides (requires firewallGroupToggles feature)
//
// GroupOverrides is a map[groupId → enabled]. Groups not present default to enabled.
// The CLI merges these into its cached policy on every sync.

import type { Env } from '../types.js'
import { authenticate } from '../lib/auth.js'
import { getPlan } from '../lib/plans.js'

export type GroupOverrides = Record<string, boolean>

// ── GET /v1/policy ────────────────────────────────────────────────────────────

export async function handleGetPolicy(request: Request, env: Env): Promise<Response> {
  const auth = await authenticate(request, env)
  if (!auth.ok) return Response.json({ error: auth.error }, { status: auth.status })

  const overrides = await getGroupOverrides(env, auth.user.user_id)
  return Response.json({ group_overrides: overrides })
}

// ── PATCH /v1/policy/groups ───────────────────────────────────────────────────

export async function handlePatchPolicyGroups(request: Request, env: Env): Promise<Response> {
  const auth = await authenticate(request, env)
  if (!auth.ok) return Response.json({ error: auth.error }, { status: auth.status })

  // Gate: only paid plans with firewallGroupToggles can customise group overrides.
  const plan = await getPlan(auth.user.plan_id, env)
  if (!plan.features.firewallGroupToggles) {
    return Response.json(
      { error: 'plan_limit', message: 'Firewall group toggles require a paid plan.' },
      { status: 403 },
    )
  }

  let body: { group_overrides: GroupOverrides }
  try {
    body = await request.json() as typeof body
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }

  if (!body?.group_overrides || typeof body.group_overrides !== 'object') {
    return Response.json({ error: 'group_overrides must be an object' }, { status: 400 })
  }

  await upsertGroupOverrides(env, auth.user.user_id, body.group_overrides)
  return Response.json({ group_overrides: body.group_overrides })
}

// ── Supabase helpers ──────────────────────────────────────────────────────────

async function getGroupOverrides(env: Env, userId: string): Promise<GroupOverrides> {
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) return {}

  const res = await fetch(
    `${env.SUPABASE_URL}/rest/v1/user_policy_overrides?user_id=eq.${userId}&select=group_id,enabled`,
    {
      headers: {
        apikey:        env.SUPABASE_SERVICE_KEY,
        Authorization: `Bearer ${env.SUPABASE_SERVICE_KEY}`,
      },
    },
  )
  if (!res.ok) return {}

  const rows = await res.json() as Array<{ group_id: string; enabled: boolean }>
  const out: GroupOverrides = {}
  for (const row of rows) {
    out[row.group_id] = row.enabled
  }
  return out
}

async function upsertGroupOverrides(env: Env, userId: string, overrides: GroupOverrides): Promise<void> {
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) return

  const rows = Object.entries(overrides).map(([group_id, enabled]) => ({
    user_id: userId,
    group_id,
    enabled,
    updated_at: new Date().toISOString(),
  }))

  if (rows.length === 0) return

  await fetch(
    `${env.SUPABASE_URL}/rest/v1/user_policy_overrides`,
    {
      method: 'POST',
      headers: {
        apikey:         env.SUPABASE_SERVICE_KEY,
        Authorization:  `Bearer ${env.SUPABASE_SERVICE_KEY}`,
        'Content-Type': 'application/json',
        Prefer:         'resolution=merge-duplicates',
      },
      body: JSON.stringify(rows),
    },
  )
}
