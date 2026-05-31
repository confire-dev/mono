// Generic optimizer — ported from leanmcp/optimizers/generic.js
// Universal noise stripping for any tool, any host.
// Applied last as fallback when no specific optimizer matched.

const ALWAYS_STRIP = new Set([
  'self', 'node_id', 'gravatar_id', 'avatar', 'avatarUrls', 'avatar_url',
  'iconUrl', 'icon_url', 'icon', 'expand', 'schema', 'renderedFields',
  'versionedRepresentations', 'editmeta', 'operations', 'names',
  'author_association', 'active_lock_reason', 'auto_merge',
  'merge_commit_sha', 'performed_via_github_app',
  'timeline_url', 'repository_url', 'events_url', 'commits_url',
  'review_comments_url', 'review_comment_url', 'comments_url',
  'statuses_url', 'diff_url', 'patch_url', 'issue_url',
  'followers_url', 'following_url', 'gists_url', 'starred_url',
  'subscriptions_url', 'organizations_url', 'repos_url', 'received_events_url',
  'color', 'initials', 'profilePicture', 'gravatar', '_links',
])

const KEEP_URL_FIELDS = new Set(['url', 'html_url', 'web_url', 'webUrl', 'permalink'])
const MS_TS_RE = /^\d{13}$/

function msToDate(v: number | string): string {
  try { return new Date(Number(v)).toISOString().slice(0, 10) } catch { return String(v) }
}

function isMsTimestamp(v: unknown): boolean {
  return (typeof v === 'string' && MS_TS_RE.test(v)) ||
         (typeof v === 'number' && v > 1e12 && v < 2e13)
}

function isEmpty(v: unknown): boolean {
  if (v === null || v === undefined || v === '') return true
  if (Array.isArray(v) && v.length === 0) return true
  if (typeof v === 'object' && !Array.isArray(v) && Object.keys(v as object).length === 0) return true
  return false
}

function isUrlField(key: string): boolean {
  if (KEEP_URL_FIELDS.has(key)) return false
  return key.endsWith('_url') || key.endsWith('Url') || key.endsWith('URL') || key.endsWith('Uri') || key.endsWith('uri')
}

function cleanUser(obj: unknown): unknown {
  if (!obj || typeof obj !== 'object') return obj
  const o = obj as Record<string, unknown>
  const hasUser = 'login' in o || 'displayName' in o || 'email' in o || 'username' in o
  if (!hasUser) return obj
  return o['email'] ?? o['displayName'] ?? o['login'] ?? o['username'] ?? obj
}

function strip(obj: unknown, depth = 0): unknown {
  if (depth > 12) return obj
  if (Array.isArray(obj)) {
    return obj.map(i => strip(i, depth + 1)).filter(i => !isEmpty(i))
  }
  if (obj && typeof obj === 'object') {
    const out: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
      if (ALWAYS_STRIP.has(k)) continue
      if (isUrlField(k)) continue
      if (isEmpty(v)) continue
      if (isMsTimestamp(v)) { out[k] = msToDate(v as number); continue }
      if (['user','author','creator','assignee','reporter'].includes(k)) {
        const c = cleanUser(v)
        if (c && !isEmpty(c)) out[k] = c
        continue
      }
      if (['watchers','participants','subscribers'].includes(k) && Array.isArray(v)) {
        const names = (v as unknown[]).map(cleanUser).filter(Boolean)
        if (names.length) out[k] = names
        continue
      }
      const s = strip(v, depth + 1)
      if (!isEmpty(s)) out[k] = s
    }
    return out
  }
  return obj
}

const MIN_SAVINGS_PCT = 10

export function optimizeGeneric(rawText: string): string | null {
  if (!rawText) return null
  let parsed: unknown
  try { parsed = JSON.parse(rawText) } catch { return null }

  const stripped = strip(parsed)
  const raw      = JSON.stringify(parsed)
  const shrunk   = JSON.stringify(stripped)
  const savedPct = Math.round((1 - shrunk.length / raw.length) * 100)
  if (savedPct < MIN_SAVINGS_PCT) return null
  return JSON.stringify(stripped, null, 2)
}
