// Read optimizer — shrinks large file reads and can short-circuit repeated reads.
// Works on both tool.pre (inject limit before read) and tool.post (truncate after).

import type { InterceptEvent, InterceptResult } from '../types.js'

const MAX_LINES  = 500
const MAX_BYTES  = 80_000

export function optimizeRead(rawText: string, event: InterceptEvent): string | null {
  // tool.pre: inject a line limit into the Read input
  // (handled by the caller checking event.phase; here we only handle tool.post)
  if (!rawText || typeof rawText !== 'string') return null
  if (rawText.length <= MAX_BYTES) return null

  const lines = rawText.split('\n')
  if (lines.length <= MAX_LINES) {
    // Over byte limit but not line limit — byte-truncate
    const truncated = rawText.slice(0, MAX_BYTES)
    const dropped = rawText.length - MAX_BYTES
    return truncated + `\n[confire: ${dropped} bytes truncated — use offset/limit params to read more]`
  }

  // Over both limits — line-truncate
  const head = lines.slice(0, MAX_LINES)
  const dropped = lines.length - MAX_LINES
  return head.join('\n') + `\n[confire: ${dropped} lines truncated — use offset/limit params to read more]`
}

// buildPreResult is called by the host adapter for PreToolUse on a Read call.
// It injects a line limit into the tool input to avoid reading a huge file at all.
export function buildReadPreResult(event: InterceptEvent): InterceptResult | null {
  const input = event.tool?.input as Record<string, unknown> | undefined
  if (!input) return null

  // Only inject limit if no limit already specified
  if (input['limit'] || input['offset']) return null

  const path = String(input['file_path'] ?? '')
  // Heuristic: don't cap source files we're actively editing
  if (!path) return null

  return {
    kind: 'replace-input',
    toolInput: { ...input, limit: MAX_LINES },
    stats: { beforeBytes: 0, afterBytes: 0, optimizer: 'read-pre' },
  }
}

export function handlesRead(event: InterceptEvent): boolean {
  return event.tool?.name === 'Read'
}
