import type { InterceptEvent, InterceptResult } from '../types.js'
import { optimizeFigma, handlesFigma } from './figma.js'
import { optimizeGeneric } from './generic.js'
import { optimizeBash, handlesBash } from './bash.js'
import { optimizeRead, handlesRead } from './read.js'
import { optimizeWebFetch, handlesWebFetch } from './webfetch.js'
import { optimizeGitHub, handlesGitHub } from './github.js'
import { optimizeJira, optimizeConfluence, handlesAtlassian, isConfluence } from './atlassian.js'
import { optimizeClickUp, handlesClickUp } from './clickup.js'
import { optimizeSlack, handlesSlack } from './slack.js'
import { optimizeAmplitude, handlesAmplitude } from './amplitude.js'
import { optimizeFireflies, handlesFireflies } from './fireflies.js'
import { optimizeNotion, handlesNotion } from './notion.js'
import { optimizePlaywright, handlesPlaywright } from './playwright.js'
import { optimizeZapier, handlesZapier } from './zapier.js'
import { optimizeGoogleDrive, handlesGoogleDrive } from './google-drive.js'
import { optimizeWebSearch, handlesWebSearch } from './websearch.js'

// extractText pulls the relevant string from a tool_response.
// Handles all common shapes:
//   Native Claude Code tools  → raw string
//   MCP tools                 → {content:[{type:"text",text:"..."}]}
//   OpenAI / Groq / Llama     → {choices:[{message:{content:"..."}}]}
//   Gemini / Vertex AI        → {candidates:[{content:{parts:[{text:"..."}]}}]}
//   output-field variant      → {output:"..."}
//   Fallback                  → JSON.stringify (generic optimizer handles noise)
export function extractText(toolResponse: unknown): string | null {
  if (typeof toolResponse === 'string') return toolResponse
  if (!toolResponse || typeof toolResponse !== 'object') return null
  const r = toolResponse as Record<string, unknown>

  // MCP format: {content:[{type:"text",text:"..."}]}
  if (Array.isArray(r['content'])) {
    return (r['content'] as unknown[])
      .filter((c): c is Record<string, unknown> => !!c && typeof c === 'object' && (c as Record<string,unknown>)['type'] === 'text')
      .map(c => String(c['text'] ?? ''))
      .join('\n') || null
  }

  // OpenAI / Groq / Llama: {choices:[{message:{content:"..."}, delta:{content:"..."}}]}
  if (Array.isArray(r['choices']) && (r['choices'] as unknown[]).length > 0) {
    const first = (r['choices'] as unknown[])[0] as Record<string, unknown>
    const msg = (first['message'] ?? first['delta']) as Record<string, unknown> | undefined
    if (msg && typeof msg['content'] === 'string' && msg['content']) return msg['content']
  }

  // Gemini / Vertex AI: {candidates:[{content:{parts:[{text:"..."}]}}]}
  if (Array.isArray(r['candidates']) && (r['candidates'] as unknown[]).length > 0) {
    const first = (r['candidates'] as unknown[])[0] as Record<string, unknown>
    const content = first['content'] as Record<string, unknown> | undefined
    if (content && Array.isArray(content['parts']) && (content['parts'] as unknown[]).length > 0) {
      const text = ((content['parts'] as unknown[])[0] as Record<string, unknown>)['text']
      if (typeof text === 'string' && text) return text
    }
  }

  if (typeof r['output'] === 'string') return r['output']
  return JSON.stringify(toolResponse)
}

// rebuildOutput puts the optimized string back into the tool's original shape
// so updatedToolOutput passes the outputSchema safeParse on the Claude Code side.
export function rebuildOutput(original: unknown, optimizedText: string): unknown {
  if (typeof original === 'string') return optimizedText
  if (!original || typeof original !== 'object') return optimizedText
  const r = original as Record<string, unknown>
  if (Array.isArray(r['content'])) {
    const newContent = (r['content'] as unknown[]).map((c, i) => {
      if (!c || typeof c !== 'object' || (c as Record<string,unknown>)['type'] !== 'text') return c
      if (i === 0) return { ...(c as object), text: optimizedText }
      return { ...(c as object), text: '' }
    })
    return { ...r, content: newContent }
  }
  if ('output' in r) return { ...r, output: optimizedText }
  return optimizedText
}

// dispatch routes to the correct optimizer function for a given event.
// Returns the optimized string, or null if no optimization applied.
function dispatch(rawText: string, event: InterceptEvent): string | null {
  if (handlesFigma(event))     return optimizeFigma(rawText)
  if (handlesGitHub(event))    return optimizeGitHub(rawText)
  if (handlesAtlassian(event)) {
    return isConfluence(event.tool?.name ?? '')
      ? optimizeConfluence(rawText)
      : optimizeJira(rawText)
  }
  if (handlesClickUp(event))    return optimizeClickUp(rawText)
  if (handlesSlack(event))      return optimizeSlack(rawText)
  if (handlesAmplitude(event))  return optimizeAmplitude(rawText)
  if (handlesFireflies(event))  return optimizeFireflies(rawText)
  if (handlesNotion(event))     return optimizeNotion(rawText)
  if (handlesPlaywright(event)) return optimizePlaywright(rawText)
  if (handlesZapier(event))     return optimizeZapier(rawText)
  if (handlesGoogleDrive(event)) return optimizeGoogleDrive(rawText)
  if (handlesBash(event))       return optimizeBash(rawText, event)
  if (handlesRead(event))       return optimizeRead(rawText, event)
  if (handlesWebFetch(event))   return optimizeWebFetch(rawText)
  if (handlesWebSearch(event))  return optimizeWebSearch(rawText)
  // Generic fallback — handles any unrecognized tool
  return optimizeGeneric(rawText)
}

export function runOptimizers(event: InterceptEvent): InterceptResult {
  if (!event.tool?.output) return { kind: 'passthrough' }

  const rawText = extractText(event.tool.output)
  if (!rawText) return { kind: 'passthrough' }

  let optimized: string | null = null
  try {
    optimized = dispatch(rawText, event)
  } catch {
    return { kind: 'passthrough' }
  }

  if (!optimized || optimized === rawText) return { kind: 'passthrough' }

  const before = rawText.length
  const after  = optimized.length
  if (after >= before) return { kind: 'passthrough' }

  return {
    kind: 'replace-output',
    toolOutput: rebuildOutput(event.tool.output, optimized),
    stats: { beforeBytes: before, afterBytes: after, optimizer: resolveOptimizerName(event) },
  }
}

export function resolveOptimizerName(event: InterceptEvent): string {
  const n = event.tool?.name?.toLowerCase() ?? ''
  const s = event.tool?.mcpServer?.toLowerCase() ?? ''
  for (const key of ['figma','github','atlassian','clickup','slack','amplitude','fireflies','notion','playwright','zapier','google_drive','googledrive','bash','read','webfetch','websearch','brave_','exa_','tavily','perplexity']) {
    if (n.includes(key) || s.includes(key)) return key
  }
  return 'generic'
}
