import type { InterceptEvent, InterceptResult } from './types.js'
import { classify } from './security/classifier.js'
import { extractText } from './security/extract.js'

// handle is the single entry-point for all InterceptEvents.
// The Worker is now a sync/reporting backend — it classifies security events
// and returns passthrough for normal tool calls.
export function handle(event: InterceptEvent): InterceptResult {
  if (event.phase === 'context.pre-compact') {
    return { kind: 'passthrough' }
  }

  switch (event.phase) {
    case 'tool.pre':
      return { kind: 'passthrough' }
    case 'tool.post':
      return handleToolPost(event)
    case 'session.start':
    case 'session.end':
    case 'tool.batch.post':
    case 'turn.stop':
    case 'prompt.submit':
      return { kind: 'passthrough' }
    default:
      return { kind: 'passthrough' }
  }
}

function handleToolPost(event: InterceptEvent): InterceptResult {
  if (!event.tool?.output) return { kind: 'passthrough' }

  const text = extractText(event.tool.output)
  if (!text) return { kind: 'passthrough' }

  const result = classify(text)
  if (result.risk === 'NONE') return { kind: 'passthrough' }

  // High-risk tool output — flag but don't replace (daemon handles sanitization locally)
  // The worker's role is classification and telemetry, not content mutation.
  return { kind: 'passthrough' }
}
