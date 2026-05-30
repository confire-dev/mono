import type { InterceptEvent } from '../types.js'

const BOT_RE = /\[bot\]$/i
const MAX_PATCH = 150
const MAX_PATCH_TEST = 80
const TEST_FILE_RE = /\.(test|spec)\.[jt]sx?$|\/tests?\/|playwright|__tests__/

function trunc(s: unknown, n: number): string {
  const str = String(s ?? '')
  return str.length > n ? str.slice(0, n) + `\n…[truncated at ${n} chars]` : str
}
function cleanUser(u: unknown): string | null {
  const o = u as Record<string,unknown> | null
  return (o?.['login'] as string) ?? null
}
function isBot(u: unknown): boolean {
  return BOT_RE.test((u as Record<string,unknown>)?.['login'] as string ?? '')
}
function truncPatch(patch: string | undefined, filename: string): string | null {
  if (!patch) return null
  const max = TEST_FILE_RE.test(filename) ? MAX_PATCH_TEST : MAX_PATCH
  const lines = patch.split('\n')
  if (lines.length <= max) return patch
  return lines.slice(0, max).join('\n') + `\n…[+${lines.length - max} lines truncated]`
}

export function optimizeGitHub(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }

  if (!data || typeof data !== 'object') return null
  const d = data as Record<string, unknown>

  // Single PR object
  if (d['number'] && d['head']) {
    const result = {
      number: d['number'], title: d['title'],
      body: trunc(d['body'], 2000), state: d['state'], draft: d['draft'],
      mergeable: d['mergeable_state'], author: cleanUser(d['user']),
      branch: (d['head'] as Record<string,unknown>)?.['ref'],
      base: (d['base'] as Record<string,unknown>)?.['ref'],
      created: (d['created_at'] as string)?.slice(0, 10),
      updated: (d['updated_at'] as string)?.slice(0, 10),
      merged: d['merged'], commits: d['commits'],
      additions: d['additions'], deletions: d['deletions'], changed_files: d['changed_files'],
      labels: (d['labels'] as Array<Record<string,unknown>>)?.map(l => l['name']) ?? [],
      reviewers: (d['requested_reviewers'] as unknown[])?.map(cleanUser) ?? [],
    }
    return JSON.stringify(result, null, 2)
  }

  // Files array
  if (Array.isArray(data) && (data[0] as Record<string,unknown>)?.['filename'] !== undefined) {
    return JSON.stringify(data.map((f: Record<string,unknown>) => ({
      file: f['filename'], status: f['status'], '+': f['additions'], '-': f['deletions'],
      ...(f['patch'] ? { diff: truncPatch(f['patch'] as string, f['filename'] as string) } : {}),
    })), null, 2)
  }

  // Reviews array
  if (Array.isArray(data) && (data[0] as Record<string,unknown>)?.['commit_id'] !== undefined) {
    const reviews = data
      .filter(r => !isBot((r as Record<string,unknown>)['user']))
      .filter(r => {
        const rv = r as Record<string,unknown>
        return rv['body'] || ['APPROVED','CHANGES_REQUESTED'].includes(rv['state'] as string)
      })
      .map((r: Record<string,unknown>) => ({
        author: cleanUser(r['user']), state: r['state'], body: trunc(r['body'], 400),
      }))
    return JSON.stringify(reviews, null, 2)
  }

  // Comments array
  if (Array.isArray(data) && (data[0] as Record<string,unknown>)?.['body'] !== undefined) {
    const humans = data.filter(c => !isBot((c as Record<string,unknown>)['user']))
    const bots = data.length - humans.length
    const result: unknown[] = humans.slice(-10).map((c: Record<string,unknown>) => ({
      author: cleanUser(c['user']),
      ...(c['path'] ? { file: c['path'], line: c['line'] } : {}),
      body: trunc(c['body'], 500),
    }))
    if (bots > 0) result.push({ _meta: { bot_comments_dropped: bots } })
    return JSON.stringify(result, null, 2)
  }
  return null
}

export function handlesGitHub(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  return n.includes('github') || n.includes('gh_')
}
