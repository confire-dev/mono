// Bash optimizer — strips terminal control noise only.
// Content (test output, build output, logs) is passed through intact.
// The model asked Bash to run a command — it gets the full result.

import type { InterceptEvent } from '../types.js'

// Terminal progress indicators — format noise, never content
const PROGRESS_RE = /[─-╿▀-▟■-◿]|={3,}|#{3,}|\[=+>?\s*\]|\d+%.*\r/
// Carriage-return overwrite lines (terminal animation artifacts)
const CR_LINE_RE  = /^[^\n]*\r[^\n]*/gm

export function optimizeBash(rawText: string, _event: InterceptEvent): string | null {
  if (!rawText || typeof rawText !== 'string') return null

  let result = rawText
    .replace(CR_LINE_RE, '')
    .split('\n')
    .filter(l => !PROGRESS_RE.test(l))
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()

  return result.length < rawText.length * 0.9 ? result : null
}

export function handlesBash(event: InterceptEvent): boolean {
  return event.tool?.name === 'Bash'
}
