import type { InterceptEvent } from '../types.js'

function msToDate(ms: unknown): string | null {
  if (!ms) return null
  try { return new Date(parseInt(String(ms))).toISOString().slice(0, 10) } catch { return null }
}
function cleanUser(u: unknown): string | null {
  const o = u as Record<string,string> | null
  return o?.['email'] ?? o?.['username'] ?? null
}

function resolveCustomField(cf: Record<string,unknown>): [string, unknown] | null {
  const val = cf['value']
  if (val === null || val === undefined || val === '') return null
  const type = cf['type'] as string
  const name = cf['name'] as string
  if (type === 'url' && val) return [name, val]
  if (type === 'drop_down') {
    const opts = (cf['type_config'] as Record<string,unknown>)?.['options'] as Record<string,string>[]
    const idx = typeof val === 'number' ? val : parseInt(String(val))
    const opt = opts?.[idx]
    return opt ? [name, opt['name']] : null
  }
  if (type === 'labels' && val) {
    if (Array.isArray(val)) return val.length ? [name, val.map((v: Record<string,string>) => v['label'] ?? v).join(', ')] : null
    return [name, val]
  }
  if (type === 'number') return [name, val]
  if (type === 'text' && val) return [name, String(val)]
  if (type === 'date' && val) return [name, msToDate(val)]
  if (type === 'checkbox') return [name, !!val]
  if (val && typeof val === 'string') return [name, val]
  return null
}

export function optimizeClickUp(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }
  if (!data || typeof data !== 'object') return null
  const d = data as Record<string,unknown>
  const task = d['id'] ? d : (d['task'] ?? null) as Record<string,unknown> | null
  if (!task) return null
  const t = task as Record<string,unknown>

  const customFields: Record<string, unknown> = {}
  for (const cf of (t['custom_fields'] as Record<string,unknown>[]) ?? []) {
    const r = resolveCustomField(cf)
    if (r) customFields[r[0]] = r[1]
  }

  const result: Record<string, unknown> = {
    id: t['id'], name: t['name'],
    status: (t['status'] as Record<string,string>)?.['status'] ?? t['status'],
    priority: (t['priority'] as Record<string,string>)?.['priority'] ?? t['priority'],
    sprint: (t['locations'] as Record<string,string>[])?.[0]?.['name'] ?? (t['list'] as Record<string,string>)?.['name'],
    url: t['url'], parent: t['parent'],
    assignees: ((t['assignees'] as unknown[]) ?? []).map(cleanUser).filter(Boolean),
    due_date: t['due_date'] ? msToDate(t['due_date']) : null,
    created: msToDate(t['date_created']), updated: msToDate(t['date_updated']),
    description: t['text_content'] ?? t['description'],
  }
  if (Object.keys(customFields).length) result['custom_fields'] = customFields
  if ((t['subtasks'] as unknown[])?.length) {
    result['subtasks'] = (t['subtasks'] as Record<string,unknown>[]).map(s => ({
      id: s['id'], name: s['name'],
      status: (s['status'] as Record<string,string>)?.['status'] ?? s['status'],
    }))
  }
  if ((t['comments'] as unknown[])?.length) {
    result['comments'] = (t['comments'] as Record<string,unknown>[]).map(c => ({
      author: cleanUser(c['user'] ?? c), date: msToDate(c['date']),
      body: String(c['comment_text'] ?? c['text_content'] ?? c['body'] ?? ''),
    }))
  }
  return JSON.stringify(result, null, 2)
}

export function handlesClickUp(event: InterceptEvent): boolean {
  return (event.tool?.name?.toLowerCase() ?? '').includes('clickup')
}
