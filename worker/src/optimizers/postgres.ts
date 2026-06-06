import type { InterceptEvent } from '../types.js'

const STRONG_PG_TOOLS = new Set(['execute_sql', 'explain_query', 'get_top_queries', 'analyze_db_health', 'run_sql'])
const WEAK_PG_TOOLS   = new Set(['list_schemas', 'list_objects', 'get_object_details', 'analyze_workload_indexes', 'analyze_query_indexes'])
const PG_SERVER_RE    = /\b(?:postgres|postgresql|neon|supabase|mysql|sqlite)\b/

// Python datetime.datetime(Y, M, D, H, Mi, S[, us]) → ISO-8601 string
const DATETIME_RE = /datetime\.datetime\((\d+),\s*(\d+),\s*(\d+),\s*(\d+),\s*(\d+),\s*(\d+)(?:,\s*\d+)?\)/g
// Python literal keywords that differ from JSON
const PY_TRUE  = /\bTrue\b/g
const PY_FALSE = /\bFalse\b/g
const PY_NONE  = /\bNone\b/g

// System schemas that are never user-relevant
const SYSTEM_SCHEMAS = new Set(['information_schema', 'pg_catalog', 'pg_toast'])

// Row threshold: only truncate when no LIMIT was specified AND result is huge
const MAX_ROWS = 50

function normalizePythonRepr(s: string): string {
  return s
    .replace(DATETIME_RE, (_, Y, M, D, H, Mi, S) =>
      `"${pad(Y, 4)}-${pad(M)}-${pad(D)}T${pad(H)}:${pad(Mi)}:${pad(S)}Z"`
    )
    .replace(PY_TRUE,  'true')
    .replace(PY_FALSE, 'false')
    .replace(PY_NONE,  'null')
}

function pad(n: string, len = 2): string {
  return String(n).padStart(len, '0')
}

function hasExplicitLimit(event: InterceptEvent): boolean {
  const input = event.tool?.input
  if (!input || typeof input !== 'object') return false
  const q = String(
    (input as Record<string, unknown>)['query'] ??
    (input as Record<string, unknown>)['sql']   ??
    ''
  )
  return /\bLIMIT\b/i.test(q)
}

function optimizeExecuteSql(rawText: string, event: InterceptEvent): string | null {
  if (!rawText || rawText.trim() === 'No results') return null

  const normalized = normalizePythonRepr(rawText)

  let rows: unknown[]
  try {
    const parsed = JSON.parse(normalized)
    if (!Array.isArray(parsed)) return JSON.stringify(parsed, null, 2)
    rows = parsed
  } catch {
    // Couldn't parse as JSON after normalization — return the cleaned text if it changed
    return normalized !== rawText ? normalized : null
  }

  const total = rows.length

  // User asked for exactly this many rows — honour it completely
  if (hasExplicitLimit(event) || total <= MAX_ROWS) {
    const out = JSON.stringify(rows, null, 2)
    return out !== rawText ? out : null
  }

  // No LIMIT on a large result: likely exploratory — truncate and tell the model
  const truncated = rows.slice(0, MAX_ROWS)
  return (
    JSON.stringify(truncated, null, 2) +
    `\n[confire: showing ${MAX_ROWS} of ${total} rows — add LIMIT to query for the full result set]`
  )
}

function optimizeListSchemas(rawText: string): string | null {
  const normalized = normalizePythonRepr(rawText)
  let parsed: unknown
  try { parsed = JSON.parse(normalized) } catch { return null }

  if (!Array.isArray(parsed)) return null

  const user = parsed.filter((s: unknown) => {
    const name = String((s as Record<string, unknown>)?.['schema_name'] ?? '')
    return !SYSTEM_SCHEMAS.has(name) && !name.startsWith('pg_')
  })

  if (user.length === parsed.length) {
    // No system schemas to drop — still worth pretty-printing if format changed
    return normalized !== rawText ? JSON.stringify(parsed, null, 2) : null
  }
  return JSON.stringify(user, null, 2)
}

export function optimizePostgres(rawText: string, event: InterceptEvent): string | null {
  if (!rawText) return null
  const n = event.tool?.name ?? ''

  if (n === 'execute_sql' || n === 'run_sql') return optimizeExecuteSql(rawText, event)
  if (n === 'list_schemas')                   return optimizeListSchemas(rawText)

  // For other tools: at minimum normalize Python repr format
  const normalized = normalizePythonRepr(rawText)
  if (normalized === rawText) return null
  try {
    return JSON.stringify(JSON.parse(normalized), null, 2)
  } catch {
    return normalized
  }
}

export function handlesPostgres(event: InterceptEvent): boolean {
  const n = event.tool?.name ?? ''
  const s = event.tool?.mcpServer?.toLowerCase() ?? ''
  if (STRONG_PG_TOOLS.has(n)) return true
  if (WEAK_PG_TOOLS.has(n) && PG_SERVER_RE.test(s)) return true
  return false
}
