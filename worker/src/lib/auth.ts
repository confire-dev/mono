// API key authentication middleware.
// Validates the Bearer token from the Authorization header against Supabase.
// Returns only the authenticated Profile — entitlement checks happen in optimize.ts
// using canUseRemoteOptimizer() from entitlement.ts.

import type { Env } from '../types.js'
import { validateApiKey, type Profile } from './supabase.js'

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
  const user = await validateApiKey(cfg, rawKey)
  if (!user) return { ok: false, status: 401, error: 'invalid API key' }

  return { ok: true, user }
}
