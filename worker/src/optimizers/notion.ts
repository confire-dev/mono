import type { InterceptEvent } from '../types.js'

const PREAMBLE_RE    = /^Here is the result of "[^"]*" for the [^\n]+\n+/
const PAGE_TAG_RE    = /<page[^>]*>\n?/
const CLOSE_TAG_RE   = /\n?<\/page>\s*$/
const CONTENT_TAG_RE = /<\/?content>\n?/g
const PROPERTIES_RE  = /<properties>[^<]*<\/properties>\n?/
const UNKNOWN_RE     = /<unknown[^>]*\/>\n?/g
const EMPTY_BLOCK_RE = /<empty-block\/>\n?/g
const SPAN_COLOR_RE  = /<span color="[^"]*">([^<]*)<\/span>/g
const MULTI_NL_RE    = /\n{3,}/g

function parseAncestorPath(text: string): { breadcrumb: string | null; cleaned: string } {
  const match = text.match(/<ancestor-path>([\s\S]*?)<\/ancestor-path>/)
  if (!match) return { breadcrumb: null, cleaned: text }
  const pages = [...match[1]!.matchAll(/<parent-page[^>]*title="([^"]*)"[^/]*\/>/g)]
  const breadcrumb = pages.length ? `📄 ${pages.map(m => m[1]).join(' › ')}` : null
  const cleaned = text.replace(/<ancestor-path>[\s\S]*?<\/ancestor-path>\n?/, '')
  return { breadcrumb, cleaned }
}

export function optimizeNotion(rawText: string): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  let text = rawText
  try {
    const p = JSON.parse(rawText) as Record<string, string>
    if (p['text']) text = p['text']
  } catch { /* raw text */ }

  const isNotion = text.includes('<page url=') || text.includes('<ancestor-path>') || !!text.match(/^Here is the result of "view"/)
  if (!isNotion) return null

  text = text.replace(PREAMBLE_RE, '')
  text = text.replace(PAGE_TAG_RE, '')
  text = text.replace(CLOSE_TAG_RE, '')
  const { breadcrumb, cleaned } = parseAncestorPath(text)
  text = cleaned
  text = text.replace(PROPERTIES_RE, '')
  text = text.replace(CONTENT_TAG_RE, '')
  text = text.replace(UNKNOWN_RE, '')
  text = text.replace(EMPTY_BLOCK_RE, '')
  text = text.replace(SPAN_COLOR_RE, '$1')
  text = text.replace(MULTI_NL_RE, '\n\n')
  if (breadcrumb) text = `${breadcrumb}\n\n${text.trimStart()}`
  text = text.trim()

  const original = (() => { try { return (JSON.parse(rawText) as Record<string,string>)['text'] ?? rawText } catch { return rawText } })()
  if (text.length >= original.length * 0.98) return null
  return text
}

export function handlesNotion(event: InterceptEvent): boolean {
  return (event.tool?.name?.toLowerCase() ?? '').includes('notion')
}
