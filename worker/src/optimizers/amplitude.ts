import type { InterceptEvent } from '../types.js'

const TEST_PROJECT_RE = /\[test\]|^test$|^kb[\s-]+test|calendly|mailchimp|alphabot|rnd|api\.testing|sandbox/i

function tsToDate(ts: unknown): string {
  try { return new Date(parseInt(String(ts)) * 1000).toISOString().slice(0, 10) } catch { return String(ts) }
}

function flattenDefinition(defn: unknown): string[] {
  if (!defn || typeof defn !== 'object') return []
  const d = defn as Record<string, unknown>
  const criteria = new Set<string>()
  const clauses = (d['orClauses'] ?? d['andClauses'] ?? []) as unknown[][]
  for (const group of clauses) {
    const items = Array.isArray(group[0]) ? group[0] : group
    for (const clause of ((items as unknown as Record<string,unknown>)['orClauses'] ?? items) as Record<string,unknown>[]) {
      const c = (clause['orClauses'] as Record<string,unknown>[])?.[0] ?? clause
      const prop = c['type_value'] as string
      const vals = (c['operator_value'] as unknown[]) ?? []
      const op   = (c['operator'] as string) ?? '='
      const days = c['time_value']
      const neg  = clause['negated'] ? 'NOT ' : ''
      if (prop && vals.length) criteria.add(`${neg}${prop} ${op} [${vals.join(', ')}]${days ? ` (last ${days} days)` : ''}`)
    }
  }
  return [...criteria]
}

function optimizeCohort(cohort: Record<string,unknown>): Record<string,unknown> {
  return {
    id: cohort['id'], name: cohort['name'], project_id: cohort['appId'], size: cohort['size'],
    count_by: (cohort['countGroup'] as Record<string,string>)?.['name'] ?? 'User',
    last_computed: cohort['lastComputed'] ? tsToDate(cohort['lastComputed']) : null,
    owners: cohort['owners'] ?? [], is_archived: cohort['isArchived'] ?? false,
    criteria: flattenDefinition(cohort['definition']),
  }
}

export function optimizeAmplitude(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }
  if (!data || typeof data !== 'object') return null
  const d = data as Record<string,unknown>

  if (Array.isArray(d['cohorts'])) {
    const optimized = (d['cohorts'] as Record<string,unknown>[]).map(optimizeCohort)
    return JSON.stringify(optimized.length === 1 ? optimized[0] : optimized, null, 2)
  }

  if (d['user'] && (d['org'] || d['projects'])) {
    const u = d['user'] as Record<string,string>
    const o = d['org'] as Record<string,unknown>
    const result: Record<string,unknown> = {}
    if (u) result['user'] = { email: u['email'], name: u['fullName'] ?? `${u['firstName']} ${u['lastName']}`.trim(), role: u['orgRole'] }
    if (o) result['org'] = { id: o['id'], name: o['name'], plan: o['plan'], ...(o['aiContext'] ? { aiContext: o['aiContext'] } : {}) }
    if (Array.isArray(d['projects'])) {
      result['projects'] = (d['projects'] as Record<string,unknown>[])
        .filter(p => !TEST_PROJECT_RE.test(p['appName'] as string))
        .map(p => ({ id: p['appId'], name: p['appName'], ...(p['description'] ? { description: p['description'] } : {}) }))
    }
    return JSON.stringify(result, null, 2)
  }

  const events = Array.isArray(data) ? data : (d['events'] ?? d['data']) as unknown[]
  if (Array.isArray(events) && (events[0] as Record<string,unknown>)?.['eventType'] !== undefined) {
    const result = (events as Record<string,unknown>[]).map(e => ({
      name: e['eventType'] ?? e['value'],
      displayName: e['displayName'] !== e['eventType'] ? e['displayName'] : undefined,
      description: e['description'] || undefined, category: e['category'] || undefined,
      isActive: e['isActive'] ?? e['visible'], totals: e['totals'] ?? e['totalCount'],
    })).filter(e => e.name)
    return JSON.stringify(result, null, 2)
  }
  return null
}

export function handlesAmplitude(event: InterceptEvent): boolean {
  return (event.tool?.name?.toLowerCase() ?? '').includes('amplitude')
}
