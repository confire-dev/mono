// Amplitude bucketing helpers.
//
// All values forwarded to Amplitude are bucketed so:
//   1. No raw user content ever reaches Amplitude.
//   2. Individual sessions can't be fingerprinted from metric values.
//   3. Amplitude is useful for cohort analysis without storing
//      personally-identifiable usage detail.
//
// Supabase receives exact values for the user's dashboard.

export function byteBucket(n: number): string {
  if (n < 1_000)     return '<1k'
  if (n < 10_000)    return '1k_10k'
  if (n < 50_000)    return '10k_50k'
  if (n < 100_000)   return '50k_100k'
  if (n < 500_000)   return '100k_500k'
  return '>500k'
}

export function reductionBucket(ratio: number): string {
  const pct = ratio * 100
  if (pct < 20)  return '0_20'
  if (pct < 50)  return '20_50'
  if (pct < 80)  return '50_80'
  if (pct < 90)  return '80_90'
  if (pct < 95)  return '90_95'
  return '>95'
}

export function durationBucket(ms: number): string {
  if (ms < 10)   return '<10ms'
  if (ms < 50)   return '10_50ms'
  if (ms < 200)  return '50_200ms'
  if (ms < 1000) return '200ms_1s'
  return '>1s'
}

// sanitizeToolType ensures no raw tool names with sensitive info reach Amplitude.
// Maps to a safe canonical category.
export function sanitizeToolType(toolType: string): string {
  const known = ['bash','read','webfetch','glob','grep','edit','write',
    'figma','github','atlassian','clickup','slack','amplitude','fireflies',
    'notion','playwright','zapier','google_drive']
  const lower = toolType.toLowerCase()
  for (const k of known) {
    if (lower.includes(k)) return k
  }
  return 'other'
}

// ── MCP / provenance category helpers ────────────────────────────────────────
//
// Never send raw MCP server names or origin domains to Amplitude.
// Use these helpers to categorize them first.

const MCP_KNOWN_PREFIXES: Record<string, string> = {
  github:     'code_hosting',
  gitlab:     'code_hosting',
  bitbucket:  'code_hosting',
  linear:     'project_management',
  jira:       'project_management',
  slack:      'messaging',
  discord:    'messaging',
  stripe:     'payments',
  supabase:   'database',
  postgres:   'database',
  mysql:      'database',
  vercel:     'deployment',
  fly:        'deployment',
  netlify:    'deployment',
  figma:      'design',
  notion:     'docs',
  confluence: 'docs',
  datadog:    'monitoring',
  sentry:     'monitoring',
  amplitude:  'analytics',
}

export function categorizeMCPServer(serverName: string): {
  known: boolean
  category: string
} {
  if (!serverName) return { known: false, category: 'unknown' }
  const lower = serverName.toLowerCase()
  for (const [prefix, category] of Object.entries(MCP_KNOWN_PREFIXES)) {
    if (lower.includes(prefix)) return { known: true, category }
  }
  return { known: false, category: 'custom_internal' }
}

const KNOWN_PUBLIC_DOMAINS = new Set([
  'github.com', 'gitlab.com', 'bitbucket.org',
  'npmjs.com', 'pypi.org', 'crates.io',
  'stackoverflow.com', 'developer.mozilla.org',
  'docs.rs', 'pkg.go.dev',
])

export function categorizeOrigin(originDomain: string): {
  category: 'local' | 'external_domain' | 'mcp' | 'unknown'
  knownPublic: boolean
} {
  if (!originDomain) return { category: 'unknown', knownPublic: false }
  const lower = originDomain.toLowerCase()
  if (lower === 'local' || lower === 'localhost' || lower.startsWith('127.') || lower === 'workspace') {
    return { category: 'local', knownPublic: false }
  }
  if (lower === 'mcp') {
    return { category: 'mcp', knownPublic: false }
  }
  const knownPublic = KNOWN_PUBLIC_DOMAINS.has(lower)
  return { category: 'external_domain', knownPublic }
}
