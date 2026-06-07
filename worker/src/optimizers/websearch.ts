import type { InterceptEvent } from '../types.js'

interface SearchResult { title: string; url: string; source: string; snippet: string; extras: string[]; age: string }

function strField(o: Record<string, unknown>, ...keys: string[]): string {
  for (const k of keys) {
    const v = o[k]
    if (typeof v === 'string' && v) return v
  }
  return ''
}

function parseResultArray(arr: unknown[]): SearchResult[] {
  const out: SearchResult[] = []
  for (const item of arr) {
    if (!item || typeof item !== 'object') continue
    const o = item as Record<string, unknown>
    const title = strField(o, 'title')
    const url   = strField(o, 'url')
    if (!title || !url) continue
    // extra_snippets are additional factual extracts (e.g. Brave Search) — keep them
    const extras = Array.isArray(o['extra_snippets'])
      ? (o['extra_snippets'] as unknown[]).filter((s): s is string => typeof s === 'string' && s.length > 0)
      : []
    // profile.name is the publisher name — useful for source credibility
    const source = strField((o['profile'] as Record<string,unknown>) ?? {}, 'name', 'long_name')
    out.push({
      title, url, source,
      snippet: strField(o, 'description', 'content', 'text', 'snippet'),
      extras,
      age: strField(o, 'age', 'publishedDate', 'date'),
    })
  }
  return out
}

function extractResults(data: Record<string, unknown>): SearchResult[] {
  // Brave Search: {"web": {"results": [...]}}
  const web = data['web']
  if (web && typeof web === 'object' && Array.isArray((web as Record<string,unknown>)['results'])) {
    return parseResultArray((web as Record<string, unknown[]>)['results'] ?? [])
  }
  // Generic / Exa / Tavily / Claude Code: {"results": [...]}
  if (Array.isArray(data['results'])) return parseResultArray(data['results'])
  return []
}

export function optimizeWebSearch(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }
  if (!data || typeof data !== 'object') return null

  let results: SearchResult[]
  if (Array.isArray(data)) {
    results = parseResultArray(data)
  } else {
    results = extractResults(data as Record<string, unknown>)
  }
  if (!results.length) return null

  const lines: string[] = []
  for (const [i, r] of results.entries()) {
    const titleLine = r.source ? `${i + 1}. ${r.title} [${r.source}]` : `${i + 1}. ${r.title}`
    lines.push(titleLine)
    lines.push(`   ${r.url}`)
    if (r.snippet !== '') lines.push(`   ${r.snippet}`)
    for (const extra of r.extras) lines.push(`   • ${extra}`)
    if (r.age !== '')     lines.push(`   ${r.age}`)
  }

  const text = lines.join('\n')
  return text.length < rawText.length * 0.9 ? text : null
}

export function handlesWebSearch(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  return n === 'websearch' ||
    n.includes('web_search') ||
    n.includes('brave_') ||
    n.includes('exa_') ||
    n.includes('tavily') ||
    n.includes('perplexity')
}
