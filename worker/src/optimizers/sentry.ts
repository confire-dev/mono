import type { InterceptEvent } from '../types.js'

// Templated boilerplate — same text on every response, never contains issue data
const PRESENTATION_RE = /\*\*Suggested presentation:\*\*[^\n]*\n\n?/g
const QUERY_TRANS_RE  = /## Query Translation\n[\s\S]*?(?=\n## |\nFound |\n\[View in Sentry\]|$)/gm
const NEXT_STEPS_RE   = /\n## Next Steps\n[\s\S]*$/m
const RESP_NOTES_RE   = /\n## Response Notes\n[\s\S]*$/m
const SLUG_NOTE_RE    = /\*Note: Use the organization slug[^*]*\*/g
// Visual decoration only — zero information content
const HR_RE           = /^─+\n?/gm

export function optimizeSentry(rawText: string): string | null {
  if (!rawText) return null

  let out = rawText

  // Strip templated coaching sections (always identical boilerplate, never issue data)
  out = out.replace(PRESENTATION_RE, '')
  out = out.replace(QUERY_TRANS_RE, '')
  out = out.replace(SLUG_NOTE_RE, '')
  out = out.replace(HR_RE, '')
  out = out.replace(NEXT_STEPS_RE, '')
  out = out.replace(RESP_NOTES_RE, '')

  out = out.replace(/\n{3,}/g, '\n\n')
  const result = out.trim()
  return result !== rawText.trim() ? result : null
}

export function handlesSentry(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  const s = event.tool?.mcpServer?.toLowerCase() ?? ''
  return n.includes('sentry') || s.includes('sentry')
}
