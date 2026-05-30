import type { InterceptEvent } from '../types.js'

function trunc(s: unknown, n: number): string {
  const str = String(s ?? '')
  return str.length > n ? str.slice(0, n) + `\n…[truncated at ${n} chars]` : str
}
function cleanUser(u: unknown): string | null {
  const o = u as Record<string,unknown> | null
  if (!o) return null
  return (o['displayName'] as string) ?? (o['emailAddress'] as string) ?? (o['accountId'] as string) ?? null
}

function extractAdfText(node: unknown, depth = 0): string {
  if (!node || typeof node !== 'object') return ''
  const n = node as Record<string, unknown>
  const type = n['type'] as string
  const content = n['content'] as unknown[] | undefined
  if (type === 'text') return (n['text'] as string) ?? ''
  if (type === 'hardBreak') return '\n'
  if (type === 'paragraph') return (content?.map(c => extractAdfText(c)).join('') ?? '') + '\n'
  if (type === 'heading') return '#'.repeat((n['attrs'] as Record<string,number>)?.['level'] ?? 1) + ' ' + (content?.map(c => extractAdfText(c)).join('') ?? '') + '\n'
  if (type === 'bulletList' || type === 'orderedList') return content?.map(c => extractAdfText(c)).join('') ?? ''
  if (type === 'listItem') return '- ' + (content?.map(c => extractAdfText(c)).join('') ?? '') + '\n'
  if (type === 'codeBlock') return '```\n' + (content?.map(c => extractAdfText(c)).join('') ?? '') + '```\n'
  if (type === 'blockquote') return '> ' + (content?.map(c => extractAdfText(c)).join('') ?? '')
  if (content) return content.map(c => extractAdfText(c, depth + 1)).join('')
  return ''
}

function stripAdf(val: unknown): string | null {
  if (typeof val === 'string') return val
  const o = val as Record<string,unknown> | null
  if (o?.['type'] === 'doc') return extractAdfText(o)
  return null
}

function optimizeSingleIssue(issue: Record<string,unknown>): Record<string,unknown> {
  const f = (issue['fields'] ?? issue) as Record<string,unknown>
  const desc = stripAdf(f['description']) ?? f['description']
  const result: Record<string,unknown> = {
    key: issue['key'], summary: f['summary'],
    type: (f['issuetype'] as Record<string,string>)?.['name'],
    status: (f['status'] as Record<string,string>)?.['name'],
    priority: (f['priority'] as Record<string,string>)?.['name'],
    assignee: cleanUser(f['assignee']), reporter: cleanUser(f['reporter']),
    project: f['project'] ? `${(f['project'] as Record<string,string>)['key']} — ${(f['project'] as Record<string,string>)['name']}` : null,
    labels: f['labels'] ?? [],
    url: issue['webUrl'] ?? `https://jira.atlassian.net/browse/${issue['key']}`,
    description: trunc(desc, 3000),
  }
  const comments = ((f['comment'] as Record<string,unknown>)?.['comments'] as unknown[]) ?? []
  if (comments.length) {
    result['comments'] = (comments as Record<string,unknown>[]).slice(-5).map(c => ({
      author: cleanUser(c['author']), date: (c['created'] as string)?.slice(0, 10),
      body: trunc(stripAdf(c['body']), 400),
    }))
  }
  return result
}

export function optimizeJira(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }
  if (!data || typeof data !== 'object') return null
  const d = data as Record<string,unknown>
  const nodes = (d['issues'] as Record<string,unknown>)?.['nodes']
  if (Array.isArray(nodes)) return JSON.stringify(nodes.map(optimizeSingleIssue), null, 2)
  if (d['key'] && d['fields']) return JSON.stringify(optimizeSingleIssue(d), null, 2)
  return null
}

export function optimizeConfluence(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }
  if (!data || typeof data !== 'object') return null
  const nodes = ((data as Record<string,unknown>)['content'] as Record<string,unknown>)?.['nodes']
  if (!Array.isArray(nodes) || !nodes.length) return null
  const result = (nodes as Record<string,unknown>[]).map(page => ({
    id: page['id'], title: page['title'],
    space: (page['space'] as Record<string,string>)?.['key'],
    parent: (page['ancestors'] as Record<string,string>[])?.[page['ancestors'] ? (page['ancestors'] as unknown[]).length - 1 : 0]?.['title'] ?? null,
    last_modified: (page['lastModified'] as string)?.slice(0, 10),
    author: (page['author'] as Record<string,string>)?.['displayName'],
    url: (page['webUrl'] as string) ?? '',
    content: page['body'],
  }))
  return JSON.stringify(result, null, 2)
}

export function handlesAtlassian(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  return n.includes('atlassian') || n.includes('jira') || n.includes('confluence')
}
export function isConfluence(toolName: string): boolean {
  const n = toolName.toLowerCase()
  return n.includes('confluence') || n.includes('getconfluence') || n.includes('searchconfluence')
}
