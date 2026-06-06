// Bash optimizer — strips verbose noise from shell command output.
// Strategy: keep signal (failures, errors, structural content), collapse
// repeated lines, strip ANSI/progress. Emergency cap at 512KB only.

import type { InterceptEvent } from '../types.js'

const EMERGENCY_BYTES = 512 * 1024   // 512 KB
const EMERGENCY_HEAD  = 200 * 1024   // keep first 200 KB
const EMERGENCY_TAIL  = EMERGENCY_BYTES - EMERGENCY_HEAD  // keep last 312 KB

const TEST_PASS_RE     = /^\s*(✓|✔|PASS|passing|ok\s+\S|·\s+✓|\d+ passing)/i
const TEST_FAIL_RE     = /^\s*(✗|✘|FAIL|failing|not ok|×|\d+ failing|Error:|AssertionError)/i
const PROGRESS_RE      = /[─-╿▀-▟■-◿]|={3,}|#{3,}|\[=+>?\s*\]|\d+%.*\r/
const ANSI_RE          = /\x1b\[[0-9;]*[a-zA-Z]/g
const BUILD_ERROR_RE   = /error TS\d+|error\[E\d+\]|SyntaxError:|Cannot find|Module not found|ld: error|make\[1\].*Error/i
const BUILD_SUCCESS_RE = /Successfully compiled|Compiled \d+ files|webpack.*built|cargo.*Finished|Build succeeded/i

export function optimizeBash(rawText: string, event: InterceptEvent): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  if (rawText.length < 500) return null

  // 1. Strip ANSI escape codes.
  let result = rawText.replace(ANSI_RE, '')

  const cmd = getCommand(event)
  let lines = result.split('\n')

  // 2. Structural cleanup by output type.
  if (isTestOutput(result, cmd)) {
    result = filterTestOutput(lines)
    lines = result.split('\n')
  } else if (isBuildOutput(result)) {
    result = filterBuildOutput(lines)
    lines = result.split('\n')
  }

  // 3. Strip progress-bar / spinner lines.
  lines = lines.filter(l => !PROGRESS_RE.test(l))

  // 4. Collapse consecutive identical lines (≥3 repeats).
  lines = collapseRepeatedLines(lines)

  result = lines.join('\n')

  // 5. Emergency byte cap — fires only for very large output.
  if (result.length > EMERGENCY_BYTES) {
    const dropped = result.length - EMERGENCY_HEAD - EMERGENCY_TAIL
    if (dropped > 0) {
      result = result.slice(0, EMERGENCY_HEAD) +
        `\n[confire: ${dropped} bytes omitted — output exceeded 512KB]\n` +
        result.slice(-EMERGENCY_TAIL)
    }
  }

  return result.length < rawText.length * 0.9 ? result : null
}

function collapseRepeatedLines(lines: string[]): string[] {
  const MIN_REPEAT = 3
  if (lines.length < MIN_REPEAT) return lines
  const out: string[] = []
  let i = 0
  while (i < lines.length) {
    const current = lines[i] ?? ''
    let j = i + 1
    while (j < lines.length && lines[j] === current) j++
    const count = j - i
    if (count >= MIN_REPEAT) {
      out.push(current)
      out.push(`[confire: ${count - 1} identical lines omitted]`)
    } else {
      out.push(...lines.slice(i, j))
    }
    i = j
  }
  return out
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
