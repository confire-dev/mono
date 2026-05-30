import type { InterceptEvent } from '../types.js'
import { optimizeSlack } from './slack.js'
import { optimizeGeneric } from './generic.js'

function detectApp(data: unknown): string | null {
  const str = JSON.stringify(data)
  if (str.includes('message_ts') || str.includes('slack.com')) return 'slack'
  if (str.includes('gmail') || str.includes('labelIds')) return 'gmail'
  if (str.includes('github.com') || str.includes('pull_request')) return 'github'
  return null
}

function optimizeDiscovery(data: Record<string,unknown>): string | null {
  const apps = (data['popular_apps'] ?? data['apps']) as Record<string,unknown>[]
  if (!apps?.length) return null
  return JSON.stringify({
    apps: apps.map(a => ({
      name: a['app'] ?? a['name'], api: a['selected_api'] ?? a['api'],
      actions: a['actions'] ? `${(a['actions'] as Record<string,number>)['read'] ?? 0}r/${(a['actions'] as Record<string,number>)['write'] ?? 0}w/${(a['actions'] as Record<string,number>)['search'] ?? 0}s` : undefined,
    })).filter(a => a.name && a.api),
    total: data['total_apps_available'] ?? data['total'],
    usage: 'Use api with enable_zapier_action, then list_enabled_zapier_actions for exact action keys.',
  }, null, 2)
}

function optimizeEnabled(data: Record<string,unknown>): string | null {
  const actions = (data['actions'] ?? data['apps']) as Record<string,unknown>[]
  if (!Array.isArray(actions) || !actions.length) return null
  return JSON.stringify({
    actions: actions.map(a => ({
      key: a['action_key'] ?? a['key'], app: a['app'] ?? a['selected_api'],
      description: a['description'] ? String(a['description']).slice(0, 100) : undefined, tool: a['tool_name'],
      params: a['params'] ? Object.fromEntries(Object.entries(a['params'] as Record<string,Record<string,string>>).filter(([,v]) => v['required']).map(([k,v]) => [k, v['type'] ?? 'string'])) : undefined,
    })),
  }, null, 2)
}

export function optimizeZapier(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }
  if (!data || typeof data !== 'object') return null
  const d = data as Record<string,unknown>

  if (d['popular_apps'] || (Array.isArray(d['apps']) && (d['apps'] as Record<string,unknown>[])[0]?.['selected_api'])) return optimizeDiscovery(d)
  if (Array.isArray(d['actions']) && (d['actions'] as Record<string,unknown>[])[0]?.['action_key']) return optimizeEnabled(d)

  const app = detectApp(data)
  if (app === 'slack') return optimizeSlack(typeof data === 'string' ? data : JSON.stringify(data))
  return optimizeGeneric(rawText)
}

export function handlesZapier(event: InterceptEvent): boolean {
  return (event.tool?.name?.toLowerCase() ?? '').includes('zapier')
}
