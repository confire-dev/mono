import type { InterceptEvent } from '../types.js'

const FOUND_PREFIX_RE = /^Found \d+ issues?:\n/

// Strip apiMetrics from any JSON depth — internal rate-limit telemetry, not content
function dropApiMetrics(obj: unknown): unknown {
  if (Array.isArray(obj)) return obj.map(dropApiMetrics)
  if (obj && typeof obj === 'object') {
    const o = obj as Record<string, unknown>
    const out: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(o)) {
      if (k === 'apiMetrics') continue
      if (k === 'metadata' && v && typeof v === 'object' && 'apiMetrics' in (v as object)) {
        const { apiMetrics: _, ...rest } = v as Record<string, unknown>
        if (Object.keys(rest).length > 0) out[k] = rest
        continue
      }
      out[k] = dropApiMetrics(v)
    }
    return out
  }
  return obj
}

export function optimizeLinear(rawText: string): string | null {
  if (!rawText) return null

  // JSON path (official remote MCP returns JSON with apiMetrics bloat)
  try {
    const parsed = JSON.parse(rawText)
    const stripped = dropApiMetrics(parsed)
    const out = JSON.stringify(stripped, null, 2)
    return out.length < rawText.length ? out : null
  } catch { /* not JSON */ }

  // Plain text path (community server) — strip noisy prefix only
  const out = rawText.replace(FOUND_PREFIX_RE, '')
  return out.trim() !== rawText.trim() ? out.trim() : null
}

export function handlesLinear(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  const s = event.tool?.mcpServer?.toLowerCase() ?? ''
  return n.startsWith('linear_') || s.includes('linear')
}
