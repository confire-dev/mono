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

function pad(n: string, len = 2): string {
  return String(n).padStart(len, '0')
}

// Normalize Python repr format to valid JSON — format only, zero content removed
function normalizePythonRepr(s: string): string {
  return s
    .replace(DATETIME_RE, (_, Y, M, D, H, Mi, S) =>
      `"${pad(Y, 4)}-${pad(M)}-${pad(D)}T${pad(H)}:${pad(Mi)}:${pad(S)}Z"`
    )
    .replace(PY_TRUE,  'true')
    .replace(PY_FALSE, 'false')
    .replace(PY_NONE,  'null')
}

export function optimizePostgres(rawText: string, _event: InterceptEvent): string | null {
  if (!rawText || rawText.trim() === 'No results') return null

  // Normalize Python repr format (datetime objects, True/False/None)
  // This is purely format normalization — no content is removed
  const normalized = normalizePythonRepr(rawText)
  if (normalized === rawText) return null

  // Try to parse and pretty-print after normalization
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

