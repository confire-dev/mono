// API key authentication middleware.
// Validates the Bearer token from the Authorization header against Supabase.

import type { Env } from '../types.js'
import { validateApiKey, checkUsage, type Profile, type UsageInfo } from './supabase.js'

export interface AuthResult {
  ok: true
  user: Profile
  usage: UsageInfo
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
      id: 'dev', email: 'dev@local', plan: 'free',
      subscription_status: 'none', is_banned: false,
    }
    return { ok: true, user: devUser, usage: { ok: true, used: 0, limit: 500, plan: 'free', total_credits: 500 } }
  }

  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }

  const user = await validateApiKey(cfg, rawKey)
  if (!user) return { ok: false, status: 401, error: 'invalid API key' }

  const usage = await checkUsage(cfg, user.id, user.plan)
  return { ok: true, user, usage }
}
