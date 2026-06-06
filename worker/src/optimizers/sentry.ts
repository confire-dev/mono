import type { InterceptEvent } from '../types.js'

// Pure coaching / navigation noise — model never needs these
const PRESENTATION_RE = /\*\*Suggested presentation:\*\*[^\n]*\n\n?/g
const QUERY_TRANS_RE  = /## Query Translation\n[\s\S]*?(?=\n## |\nFound |\n\[View in Sentry\]|$)/gm
const NEXT_STEPS_RE   = /\n## Next Steps\n[\s\S]*$/m
const RESP_NOTES_RE   = /\n## Response Notes\n[\s\S]*$/m
const SLUG_NOTE_RE    = /\*Note: Use the organization slug[^*]*\*/g
// Visual decoration only
const HR_RE           = /^─+\n?/gm
// Framework / dependency stack frames — strip from Full Stacktrace only
// Kept: user code frames. "Most Relevant Frame" section is never touched.
const NOISE_FRAME_RE  = /^  (?:File "[^"]*(?:node_modules|site-packages|dist-packages|lib\/python)[^"]*"[^\n]*\n(?:    [^\n]+\n)*|at [^\n]*(?:node_modules)[^\n]*\n)/gm

export function optimizeSentry(rawText: string): string | null {
  if (!rawText) return null

  let out = rawText

  out = out.replace(PRESENTATION_RE, '')
  out = out.replace(QUERY_TRANS_RE, '')
  out = out.replace(SLUG_NOTE_RE, '')
  out = out.replace(HR_RE, '')

  // Remove framework frames but log how many so the model knows they existed
  const dropped = [...out.matchAll(NOISE_FRAME_RE)].length
  if (dropped > 0) {
    out = out.replace(NOISE_FRAME_RE, '')
    out = out.replace(
      /(\*\*Full Stacktrace:\*\*\n)/,
      `$1[confire: ${dropped} framework frame${dropped > 1 ? 's' : ''} omitted]\n`,
    )
  }

  // Trailing coaching sections — strip last so earlier regexes don't need to be greedy past them
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
