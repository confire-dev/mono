// Content-hash cache using Cloudflare KV.
// Identical tool outputs return the cached optimization without re-running the engine.
// TTL: 24 hours (tool output for the same content won't change).

import type { InterceptResult } from '../types.js'

const TTL_SECONDS = 86_400 // 24 hours

// cacheKey computes a stable key from the tool output content.
export async function cacheKey(toolOutput: unknown): Promise<string> {
  const content = typeof toolOutput === 'string' ? toolOutput : JSON.stringify(toolOutput)
  const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(content))
  const hash = Array.from(new Uint8Array(buf)).map(b => b.toString(16).padStart(2, '0')).join('')
  return `opt:${hash.slice(0, 32)}`
}

// cacheGet retrieves a cached InterceptResult.
export async function cacheGet(kv: KVNamespace, key: string): Promise<InterceptResult | null> {
  const raw = await kv.get(key)
  if (!raw) return null
  try { return JSON.parse(raw) as InterceptResult } catch { return null }
}

// cachePut stores an InterceptResult with TTL.
export async function cachePut(kv: KVNamespace, key: string, result: InterceptResult): Promise<void> {
  if (result.kind === 'passthrough') return // don't cache passthroughs
  await kv.put(key, JSON.stringify(result), { expirationTtl: TTL_SECONDS })
}
