// WebFetch optimizer — strips HTML/CSS/JS markup noise, keeps main text content.
// Never truncates — the model fetched this page and gets all its text content.

const HEAD_RE     = /<head\b[^>]*>[\s\S]*?<\/head>/gi
const SCRIPT_RE   = /<script\b[^>]*>[\s\S]*?<\/script>/gi
const STYLE_RE    = /<style\b[^>]*>[\s\S]*?<\/style>/gi
const SVG_RE      = /<svg\b[^>]*>[\s\S]*?<\/svg>/gi
const FIGURE_RE   = /<(figure|picture)\b[^>]*>[\s\S]*?<\/\1>/gi
const IMG_RE      = /<img\b[^>]*>/gi
const NAV_RE      = /<(nav|header|footer|aside|menu)\b[^>]*>[\s\S]*?<\/\1>/gi
const COMMENT_RE  = /<!--[\s\S]*?-->/g
const TAG_RE      = /<[^>]+>/g
const ATTR_RE     = /\s+(class|style|id|data-[a-z-]+|aria-[a-z-]+|onclick|on\w+)="[^"]*"/g
const MULTI_NL_RE = /\n{3,}/g
const MULTI_SP_RE = /[ \t]{2,}/g

export function optimizeWebFetch(rawText: string): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  if (rawText.length < 2_000) return null

  // Only process HTML-like content
  if (!rawText.includes('<') || !rawText.includes('>')) return null

  let text = rawText
  text = text.replace(HEAD_RE, '')
  text = text.replace(SCRIPT_RE, '')
  text = text.replace(STYLE_RE, '')
  text = text.replace(SVG_RE, '[svg]')
  text = text.replace(FIGURE_RE, '')
  text = text.replace(IMG_RE, '[image]')
  text = text.replace(NAV_RE, '')
  text = text.replace(COMMENT_RE, '')
  text = text.replace(ATTR_RE, '')
  text = text.replace(TAG_RE, ' ')
  text = text.replace(MULTI_SP_RE, ' ')
  text = text.replace(MULTI_NL_RE, '\n\n')
  text = text.trim()

  return text.length < rawText.length * 0.7 ? text : null
}

export function handlesWebFetch(event: { tool?: { name?: string } }): boolean {
  return event.tool?.name === 'WebFetch'
}
