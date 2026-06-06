// Read optimizer — emergency-only cap for very large file reads.
// No structural truncation by default; reads pass through intact
// unless they exceed the emergency threshold.

import type { InterceptEvent } from '../types.js'

const EMERGENCY_BYTES = 1 * 1024 * 1024  // 1 MB
const EMERGENCY_HEAD  = 800 * 1024        // keep first 800 KB

export function optimizeRead(rawText: string, _event: InterceptEvent): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  if (rawText.length <= EMERGENCY_BYTES) return null

  const dropped = rawText.length - EMERGENCY_HEAD
  return rawText.slice(0, EMERGENCY_HEAD) +
    `\n[confire: ${dropped} bytes omitted — file exceeds 1MB; use offset/limit to read further]`
}

export function handlesRead(event: InterceptEvent): boolean {
  return event.tool?.name === 'Read'
}
