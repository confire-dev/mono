import type { InterceptEvent } from '../types.js'

function cleanUser(u: unknown): string | null {
  const o = u as Record<string,unknown> | null
  return (o?.['login'] as string) ?? null
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
      body: d['body'], state: d['state'], draft: d['draft'],
      mergeable: d['mergeable_state'], author: cleanUser(d['user']),
      branch: (d['head'] as Record<string,unknown>)?.['ref'],
      base:   (d['base'] as Record<string,unknown>)?.['ref'],
      created: (d['created_at'] as string)?.slice(0, 10),
      updated: (d['updated_at'] as string)?.slice(0, 10),
      merged: d['merged'], commits: d['commits'],
      additions: d['additions'], deletions: d['deletions'], changed_files: d['changed_files'],
      labels:    (d['labels']               as Array<Record<string,unknown>>)?.map(l => l['name']) ?? [],
      reviewers: (d['requested_reviewers']   as unknown[])?.map(cleanUser) ?? [],
    }
    return JSON.stringify(result, null, 2)
  }

  // Files array
  if (Array.isArray(data) && (data[0] as Record<string,unknown>)?.['filename'] !== undefined) {
    return JSON.stringify(data.map((f: Record<string,unknown>) => ({
      file: f['filename'], status: f['status'], '+': f['additions'], '-': f['deletions'],
      ...(f['patch'] ? { diff: f['patch'] } : {}),
    })), null, 2)
  }

  // Reviews array
  if (Array.isArray(data) && (data[0] as Record<string,unknown>)?.['commit_id'] !== undefined) {
    const reviews = data
      .filter(r => {
        const rv = r as Record<string,unknown>
        return rv['body'] || ['APPROVED','CHANGES_REQUESTED'].includes(rv['state'] as string)
      })
      .map((r: Record<string,unknown>) => ({
        author: cleanUser(r['user']), state: r['state'], body: r['body'],
      }))
    return JSON.stringify(reviews, null, 2)
  }

  // PR list
  if (Array.isArray(data) && typeof (data[0] as Record<string,unknown>)?.['head'] === 'object') {
    return JSON.stringify((data as Record<string,unknown>[]).map(pr => ({
      number: pr['number'], title: pr['title'], state: pr['state'],
      draft: pr['draft'], author: cleanUser(pr['user']),
      branch: (pr['head'] as Record<string,unknown>)?.['ref'],
      updated: (pr['updated_at'] as string)?.slice(0, 10),
      labels: (pr['labels'] as Array<Record<string,unknown>>)?.map(l => l['name']) ?? [],
    })), null, 2)
  }

  // Comments array (all comments, all authors — model may need bot CI status too)
  if (Array.isArray(data) && (data[0] as Record<string,unknown>)?.['body'] !== undefined) {
    return JSON.stringify(data.map((c: Record<string,unknown>) => ({
      author: cleanUser(c['user']),
      ...(c['path'] ? { file: c['path'], line: c['line'] } : {}),
      body: c['body'],
    })), null, 2)
  }
  return null
}

export function handlesGitHub(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  return n.includes('github') || n.includes('gh_')
}
