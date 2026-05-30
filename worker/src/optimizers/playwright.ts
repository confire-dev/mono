import type { InterceptEvent } from '../types.js'

const NOISE_IFRAMES = [/reCAPTCHA/i, /chat widget/i, /LiveChat/i, /Intercom/i, /Zendesk/i, /HubSpot/i, /Drift/i, /Crisp/i]

export function optimizePlaywright(rawText: string, keepRefs = false): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  const isSnapshot = rawText.includes('[ref=') ||
    (rawText.includes('- generic') && rawText.includes('- main')) ||
    rawText.includes('- button ') || rawText.includes('- textbox ')
  if (!isSnapshot) return null

  let text = rawText
  if (!keepRefs) text = text.replace(/\s*\[ref=[^\]]+\]/g, '')
  text = text.replace(/\s*\[cursor=pointer\]/g, '')

  // Strip noisy iframe blocks
  const lines = text.split('\n')
  const cleaned: string[] = []
  let i = 0
  while (i < lines.length) {
    const line = lines[i]!
    const m = line.match(/^(\s*)- iframe:/)
    if (m) {
      const indent = m[1]!.length
      const block = [line]; i++
      while (i < lines.length) {
        const childIndent = (lines[i]!.match(/^(\s*)/) ?? ['',''])[1]!.length
        if (lines[i]!.trim() === '' || childIndent > indent) { block.push(lines[i]!); i++ }
        else break
      }
      if (!NOISE_IFRAMES.some(re => re.test(block.join('\n')))) cleaned.push(...block)
    } else { cleaned.push(line); i++ }
  }
  text = cleaned.join('\n')
  text = text.replace(/\n{3,}/g, '\n\n').trim()
  return text === rawText.trim() ? null : text
}

export function handlesPlaywright(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  return n.includes('playwright') || n.includes('browser_') || n.includes('chrome') || n.includes('puppeteer')
}
