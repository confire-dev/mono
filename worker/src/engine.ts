import type { InterceptEvent, InterceptResult } from './types.js'
import { runOptimizers } from './optimizers/index.js'

// handle is the single entry-point for both the Worker and (in the future) the Go daemon
// when it falls back to the bundled local optimizer.
// It routes by phase and delegates to the optimizer registry.
export function handle(event: InterceptEvent): InterceptResult {
  // pre-compact is disabled in v1 — ships dark
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
      // These phases are handled at the daemon level (notifications, stats, warming).
      // The Worker just returns passthrough for them — the daemon adds additionalContext.
      return { kind: 'passthrough' }
    default:
      return { kind: 'passthrough' }
  }
}

function handleToolPost(event: InterceptEvent): InterceptResult {
  if (!event.tool?.output) return { kind: 'passthrough' }
  return runOptimizers(event)
}
