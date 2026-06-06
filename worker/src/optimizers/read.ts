// Read optimizer — no-op.
// The model explicitly asked to read a file and gets the full content.
// Truncating file content based on size heuristics would lose data the model may need.

import type { InterceptEvent, InterceptResult } from '../types.js'

export function optimizeRead(_rawText: string, _event: InterceptEvent): string | null {
  return null
}

// Pre-phase: no limit injection — the model decides how much of the file it needs.
export function buildReadPreResult(_event: InterceptEvent): InterceptResult | null {
  return null
}

export function handlesRead(event: InterceptEvent): boolean {
  return event.tool?.name === 'Read'
}
