import type { Env } from '../types.js'

const RULES_KV_KEY = 'firewall_rules_v1'
const RULES_CACHE_TTL = 60 * 60 // 1 hour in seconds

// handleGetRules serves the latest firewall rule bundle to CLI clients.
// The bundle is KV-cached with a 1-hour TTL so rule updates propagate quickly.
export async function handleGetRules(_request: Request, env: Env): Promise<Response> {
  const cached = await env.CACHE.get(RULES_KV_KEY, 'text')
  if (cached) {
    return new Response(cached, {
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': `public, max-age=${RULES_CACHE_TTL}`,
      },
    })
  }

  // No cached rules yet — return empty bundle.
  // Rules are pushed to KV via the admin API or Supabase webhook.
  const empty = JSON.stringify({ version: '0', rules: [], updated_at: new Date().toISOString() })
  return new Response(empty, {
    headers: { 'Content-Type': 'application/json' },
  })
}
