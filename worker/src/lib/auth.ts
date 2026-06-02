// API key authentication middleware.
// Validates the Bearer token from the Authorization header against Supabase.
// Returns only the authenticated Profile — entitlement checks happen in optimize.ts
// using canUseRemoteOptimizer() from entitlement.ts.

import type { Env } from '../types.js'
import { validateApiKey, getProfileById, type Profile } from './supabase.js'

async function verifySupabaseJWT(token: string, supabaseUrl: string, anonKey: string): Promise<{ sub: string; email: string } | null> {
  const res = await fetch(`${supabaseUrl}/auth/v1/user`, {
    headers: { apikey: anonKey, Authorization: `Bearer ${token}` },
  })
  if (!res.ok) return null
  const user = await res.json() as { id: string; email: string }
  return user.id && user.email ? { sub: user.id, email: user.email } : null
}

export interface AuthResult {
  ok: true
  user: Profile
}
export interface AuthError {
  ok: false
  status: number
  error: string
}

export async function authenticate(request: Request, env: Env): Promise<AuthResult | AuthError> {
  const auth = request.headers.get('Authorization')
  if (!auth?.startsWith('Bearer ')) {
    return { ok: false, status: 401, error: 'missing Authorization header' }
  }
  const rawKey = auth.slice(7).trim()
  if (!rawKey) {
    return { ok: false, status: 401, error: 'empty API key' }
  }

  // In development (no Supabase configured), accept any non-empty key.
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) {
    const devUser: Profile = {
      id: 'dev', email: 'dev@local', plan: 'free', plan_id: 'free',
      subscription_status: 'none', is_banned: false,
    }
    return { ok: true, user: devUser }
  }

  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }

  // Try confire API key first, then Supabase JWT (dashboard requests).
  let user = await validateApiKey(cfg, rawKey)
  if (!user && env.SUPABASE_ANON_KEY) {
    const verified = await verifySupabaseJWT(rawKey, env.SUPABASE_URL, env.SUPABASE_ANON_KEY)
    if (verified) user = await getProfileById(cfg, verified.sub)
  }
  if (!user) return { ok: false, status: 401, error: 'invalid API key' }

  return { ok: true, user }
}
