// Bash optimizer — strips verbose noise from shell command output.
// Strategy: keep signal (failures, errors, last N lines), drop passing tests / progress bars.

import type { InterceptEvent } from '../types.js'

const MAX_LINES   = 200   // keep at most this many lines
const TAIL_LINES  = 50    // if over limit, keep first 20 + last (MAX-20) lines
const MAX_BYTES   = 40_000

// Patterns that indicate a test runner pass line (keep only on failure)
const TEST_PASS_RE    = /^\s*(✓|✔|PASS|passing|ok\s+\S|·\s+✓|\d+ passing)/i
const TEST_FAIL_RE    = /^\s*(✗|✘|FAIL|failing|not ok|×|\d+ failing|Error:|AssertionError)/i
const PROGRESS_RE     = /[─-╿▀-▟■-◿]|={3,}|#{3,}|\[=+>?\s*\]|\d+%.*\r/
const BUILD_ERROR_RE  = /error TS\d+|error\[E\d+\]|SyntaxError:|Cannot find|Module not found|ld: error|make\[1\].*Error/i
const BUILD_SUCCESS_RE = /Successfully compiled|Compiled \d+ files|webpack.*built|cargo.*Finished|Build succeeded/i

export function optimizeBash(rawText: string, event: InterceptEvent): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  if (rawText.length < 500) return null // small output → not worth it

  const lines = rawText.split('\n')
  if (lines.length < 20 && rawText.length < MAX_BYTES) return null

  const cmd = getCommand(event)
  let result = rawText

  // Test runner output: keep failures + summary, strip individual passing tests
  if (isTestOutput(rawText, cmd)) {
    result = filterTestOutput(lines)
  }
  // Build output with errors: drop progress/success lines
  else if (isBuildOutput(rawText)) {
    result = filterBuildOutput(lines)
  }
  // Very long output: keep head + tail
  else if (lines.length > MAX_LINES) {
    const head = lines.slice(0, 20)
    const tail = lines.slice(-TAIL_LINES)
    const dropped = lines.length - 20 - TAIL_LINES
    result = [...head, `\n[confire: ${dropped} lines omitted]\n`, ...tail].join('\n')
  }

  // Strip progress bar lines
  result = result.split('\n')
    .filter(l => !PROGRESS_RE.test(l))
    .join('\n')

  // Byte-budget head+tail for few-lines-but-large outputs
  const HEAD_BYTES = 20_000
  const TAIL_BYTES = 10_000
  if (result.length > MAX_BYTES && result.length > HEAD_BYTES + TAIL_BYTES) {
    const dropped = result.length - HEAD_BYTES - TAIL_BYTES
    result = result.slice(0, HEAD_BYTES) +
      `\n[confire: ${dropped} bytes omitted]\n` +
      result.slice(-TAIL_BYTES)
  } else if (result.length > MAX_BYTES) {
    result = result.slice(0, MAX_BYTES) + `\n[confire: output truncated at ${MAX_BYTES} bytes]`
  }

  return result.length < rawText.length * 0.9 ? result : null
}

function isTestOutput(text: string, cmd: string): boolean {
  return TEST_PASS_RE.test(text) || TEST_FAIL_RE.test(text) ||
    /\b(jest|vitest|mocha|pytest|cargo test|go test|npm test|yarn test)\b/i.test(cmd)
}

function isBuildOutput(text: string): boolean {
  return BUILD_ERROR_RE.test(text) && BUILD_SUCCESS_RE.test(text)
}

function filterBuildOutput(lines: string[]): string {
  const errorLines: string[] = []
  let progressCount = 0
  for (const line of lines) {
    if (BUILD_ERROR_RE.test(line)) { errorLines.push(line) }
    else if (BUILD_SUCCESS_RE.test(line) || PROGRESS_RE.test(line)) { progressCount++ }
    else { errorLines.push(line) }
  }
  if (progressCount === 0) return lines.join('\n')
  return `[confire: ${progressCount} build progress lines omitted]\n` + errorLines.join('\n')
}

function filterTestOutput(lines: string[]): string {
  const out: string[] = []
  let passCount = 0
  let inFailBlock = false

  for (const line of lines) {
    if (TEST_FAIL_RE.test(line)) { inFailBlock = true; out.push(line); continue }
    if (TEST_PASS_RE.test(line) && !inFailBlock) { passCount++; continue }
    // Summary lines (contain counts) always kept
    if (/\d+\s+(passing|failing|skipped|pending)/i.test(line)) { inFailBlock = false; out.push(line); continue }
    out.push(line)
  }
  if (passCount > 0) out.unshift(`[confire: ${passCount} passing tests omitted]`)
  return out.join('\n')
}

function getCommand(event: InterceptEvent): string {
  const input = event.tool?.input
  if (!input || typeof input !== 'object') return ''
  return String((input as Record<string, unknown>)['command'] ?? '')
}

export function handlesBash(event: InterceptEvent): boolean {
  return event.tool?.name === 'Bash'
}
