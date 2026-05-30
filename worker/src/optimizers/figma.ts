// Figma optimizer — ported from leanmcp/optimizers/figma.js
// Handles both the official Figma MCP JSX output and sparse XML metadata fallback.

import type { InterceptEvent } from '../types.js'

// ── JSX strippers ─────────────────────────────────────────────────────────
const DATA_NODE_ID_RE   = /\s*data-node-id="[^"]*"/g
const CSS_VAR_RE        = /\bvar\(--[^,)]+,\s*([^)]+)\)/g
const SUPER_CRITICAL_RE = /\nSUPER CRITICAL:[\s\S]*/
const LINE_ASSET_DECL_RE = /^const (imgLine\d+|imgDivider\w*) = "[^"]+";$/gm
const LINE_IMG_ELEM_RE  = /<img[^>]*src=\{(imgLine\d+|imgDivider\w*)\}[^>]*\/>/g
const CONTENT_STRETCH_RE = /\bcontent-stretch\s*/g
const MULTI_BLANK_RE    = /\n{3,}/g

// ── Sparse metadata ───────────────────────────────────────────────────────
const SECTION_RE = /^  <frame id="([^"]+)" name="([^"]+)"[^>]*?(?:y="(\d+)")?[^>]*width="(19[012]\d)"[^>]*height="(\d+)"/gm
const ASSET_NAME_RE = /\b(asset|illustration|logo|hero|visual|graphic|banner|flow|decoration|deal)\b/gi
const NODE_TYPE_RE  = /<(frame|instance|text|vector|group)/g
const INSTANCE_RE   = /<instance /g
const HIDDEN_RE     = /hidden="true"/g
const TEXT_LABEL_RE = /<text [^>]*name="([^"]+)"/g
const SECTION_ASSET_RE = /name="([^"]*(?:asset|logo|hero|flow|deal|illustration)[^"]*)"/gi

function findLineAssets(code: string): Set<string> {
  const s = new Set<string>()
  for (const m of code.matchAll(/^const (imgLine\d+|imgDivider\w*) = "/gm)) {
    if (m[1]) s.add(m[1])
  }
  return s
}

function optimizeSparseMetadata(xml: string): string {
  const sections: Array<{id:string; name:string; y:number; h:number}> = []

  for (const m of xml.matchAll(SECTION_RE)) {
    const [, id, name, yStr, , hStr] = m
    if (!id || !name || /^Frame \d+$/.test(name)) continue
    sections.push({ id, name, y: parseInt(yStr ?? '0', 10), h: parseInt(hStr ?? '0', 10) })
  }

  const totalNodes  = (xml.match(NODE_TYPE_RE)  ?? []).length
  const hiddenNodes = (xml.match(HIDDEN_RE)      ?? []).length
  const instances   = (xml.match(INSTANCE_RE)   ?? []).length
  const assetCount  = (xml.match(ASSET_NAME_RE)  ?? []).length

  const lines: string[] = [
    '## Figma — Sparse Metadata (design too large for full code output)',
    `Total nodes: ${totalNodes} | Hidden: ${hiddenNodes} | Component instances: ${instances} | Detected assets: ${assetCount}`,
    '',
    '## Page Sections (fetch individually with get_design_context)',
    '',
  ]

  sections.sort((a, b) => a.y - b.y)

  for (let i = 0; i < sections.length; i++) {
    const s = sections[i]!
    lines.push(`### ${s.name}`)
    lines.push(`  node-id: ${s.id}  |  y: ${s.y}px  |  height: ${s.h}px`)

    const start = xml.indexOf(`id="${s.id}"`)
    const next  = sections[i + 1]
    const end   = next ? xml.indexOf(`id="${next.id}"`) : xml.length
    const slice = xml.slice(start, end)

    const sNodes = (slice.match(NODE_TYPE_RE)  ?? []).length
    const sInst  = (slice.match(INSTANCE_RE)   ?? []).length
    const sText  = [...slice.matchAll(TEXT_LABEL_RE)].map(m => m[1]).slice(0, 3)
    const sAsset = [...slice.matchAll(SECTION_ASSET_RE)].map(m => m[1]).slice(0, 4)

    lines.push(`  ${sNodes} nodes | ${sInst} components${sAsset.length ? ` | assets: ${sAsset.join(', ')}` : ''}`)
    if (sText.length) lines.push(`  text labels: ${sText.join(' · ')}`)
    lines.push('')
  }

  lines.push('## Next steps')
  lines.push('Call get_design_context on each section node-id above.')
  lines.push('Recommended order (largest/most complex first):')
  for (const s of [...sections].sort((a, b) => b.h - a.h).slice(0, 5)) {
    lines.push(`  - ${s.name}: ${s.id}`)
  }
  return lines.join('\n')
}

export function optimizeFigma(rawText: string): string | null {
  if (!rawText || typeof rawText !== 'string') return null

  const isSparse  = rawText.includes('<frame id=') && rawText.includes('IMPORTANT:')
  const isXmlTree = rawText.trimStart().startsWith('<frame id=')
  if (isSparse || isXmlTree) return optimizeSparseMetadata(rawText)

  const isFigmaJSX =
    rawText.includes('data-node-id=') ||
    rawText.includes('export default function') ||
    rawText.includes('figma.com/api/mcp/asset')
  if (!isFigmaJSX) return null

  let code = rawText
  code = code.replace(DATA_NODE_ID_RE, '')
  code = code.replace(CSS_VAR_RE, (_, fallback: string) => (fallback ?? '').trim())
  code = code.replace(CONTENT_STRETCH_RE, '')
  code = code.replace(SUPER_CRITICAL_RE, '')

  const lineAssets = findLineAssets(code)
  if (lineAssets.size > 0) {
    code = code.replace(LINE_IMG_ELEM_RE, (_, name: string) =>
      lineAssets.has(name)
        ? `{/* confire: replace with CSS text-decoration or border-bottom */}`
        : `<img src={${name}} />`
    )
    code = code.replace(LINE_ASSET_DECL_RE, '')
    code = '// confire: imgLine* → CSS text-decoration/border-bottom\n' + code
  }

  code = code.replace(MULTI_BLANK_RE, '\n\n').trim()
  return code === rawText ? null : code
}

export function handlesFigma(event: InterceptEvent): boolean {
  const name = event.tool?.name?.toLowerCase() ?? ''
  return name.includes('figma') || (event.tool?.isMcp === true && event.tool.mcpServer === 'figma')
}
