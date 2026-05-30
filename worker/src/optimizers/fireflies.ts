import type { InterceptEvent } from '../types.js'

function parseSummaryText(text: string): { meta: Record<string,string>; sections?: Record<string,string>; raw?: string } {
  const meta: Record<string,string> = {}
  const fieldRe = /^(Id|Title|DateString|Organizer Email):\s*([^\n]+)/gm
  let m: RegExpExecArray | null
  while ((m = fieldRe.exec(text)) !== null) meta[m[1]!] = m[2]!.trim()

  const summaryMatch = text.match(/Summary:\s*([\s\S]+)$/)
  if (!summaryMatch) return { meta, raw: text }

  const sectionNames = ['Action Items','Shorthand Bullet','Notes','Overview','Bullet Gist','Gist','Short Summary','Transcript Chapters','Keywords']
  const sectionRe = new RegExp(`,?\\s*(${sectionNames.join('|')}):`, 'g')
  const parts = summaryMatch[1]!.split(sectionRe)
  const sections: Record<string,string> = {}
  for (let i = 1; i < parts.length - 1; i += 2) {
    const name = parts[i]?.trim()
    const content = parts[i + 1]?.trim()
    if (name && content && content !== 'No transcript chapters') sections[name] = content
  }
  return { meta, sections }
}

function formatOptimalSummary(parsed: ReturnType<typeof parseSummaryText>): string {
  const { meta, sections, raw } = parsed
  if (!sections) return raw ?? ''
  const lines: string[] = []
  if (meta['Title']) lines.push(`# ${meta['Title']}`)
  if (meta['DateString']) lines.push(`Date: ${meta['DateString']!.slice(0, 10)}`)
  if (meta['Organizer Email']) lines.push(`Organizer: ${meta['Organizer Email']}`)
  lines.push('')
  if (sections['Short Summary']) { lines.push('## Summary'); lines.push(sections['Short Summary']!); lines.push('') }
  if (sections['Action Items']) { lines.push('## Action Items'); lines.push(sections['Action Items']!); lines.push('') }
  if (sections['Overview']) { lines.push('## Key Decisions'); lines.push(sections['Overview']!); lines.push('') }
  if (sections['Notes']) { lines.push('## Meeting Notes'); lines.push(sections['Notes']!) }
  return lines.join('\n').trim()
}

export function optimizeFireflies(rawText: string): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  const isFF = rawText.includes('Fireflies') || rawText.startsWith('Id:') ||
    (rawText.includes('DateString:') && rawText.includes('Sentences:')) ||
    (rawText.includes('Action Items') && rawText.includes('Short Summary'))
  if (!isFF) return null

  if (rawText.includes('Sentences:')) {
    const headerEnd = rawText.indexOf('Sentences:')
    const header = rawText.slice(0, headerEnd).trim()
    const sentences = rawText.slice(headerEnd)
    const id    = header.match(/^Id:\s*(.+)$/m)?.[1]?.trim()
    const date  = header.match(/^DateString:\s*(.+)$/m)?.[1]?.trim()?.slice(0, 10)
    const spkrs = header.match(/^Speakers:\s*(.+)$/m)?.[1]?.trim()
    const hint = [
      `[confire] ⚠️ Full transcript (${Math.round(rawText.length/1024)}KB, ~${Math.round(rawText.length/4)} tokens).`,
      `For summaries use get_summary (~6× smaller).`,
      `Meeting: ${date ?? 'unknown'} | Speakers: ${spkrs ?? 'unknown'}`,
      '---',
    ].join('\n')
    return hint + '\n' + sentences
  }

  const optimized = formatOptimalSummary(parseSummaryText(rawText))
  if (optimized.length >= rawText.length * 0.95) return null
  return optimized
}

export function handlesFireflies(event: InterceptEvent): boolean {
  return (event.tool?.name?.toLowerCase() ?? '').includes('fireflies')
}
