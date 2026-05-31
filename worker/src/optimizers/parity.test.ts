// Parity tests — run TS optimizers against the shared testdata fixtures and
// verify they produce the expected behaviour. The Go integration test
// (cli/optimizer/parity_test.go) runs the same fixtures through the Go
// implementations and compares outputs, confirming cross-language equivalence.

import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { optimizeBash } from './bash.js'
import { optimizeWebFetch } from './webfetch.js'
import { optimizeGeneric } from './generic.js'
import type { InterceptEvent } from '../types.js'

const FIXTURES = join(fileURLToPath(new URL('.', import.meta.url)), '../../..', 'testdata/optimizer')

function fixture(name: string): string {
  return readFileSync(join(FIXTURES, name), 'utf8')
}

function bashEvent(input: string): InterceptEvent {
  return { tool: { name: 'Bash', output: input, input: {} } }
}

// ── Bash ─────────────────────────────────────────────────────────────────────

describe('bash parity', () => {
  const input = fixture('bash-test-runner.txt')

  it('reduces output size', () => {
    const result = optimizeBash(input, bashEvent(input))
    expect(result).not.toBeNull()
    expect(result!.length).toBeLessThan(input.length)
  })

  it('strips passing test lines from pure-pass suites', () => {
    const result = optimizeBash(input, bashEvent(input))!
    // Lines from PASS-only suites (login.test.ts, Button.test.tsx) are stripped
    expect(result).not.toMatch(/✓ redirects to dashboard/)
    expect(result).not.toMatch(/✓ shows error on invalid/)
    expect(result).not.toMatch(/✓ renders correctly/)
    // Lines within the FAIL block (checkout.test.ts) may be kept — that's correct
  })

  it('preserves failing test details', () => {
    const result = optimizeBash(input, bashEvent(input))!
    expect(result).toMatch(/creates stripe checkout session/)
    expect(result).toMatch(/checkout\.test\.ts/)
  })

  it('includes omitted-tests header', () => {
    const result = optimizeBash(input, bashEvent(input))!
    expect(result).toMatch(/passing tests omitted/)
  })

  it('preserves test summary line', () => {
    const result = optimizeBash(input, bashEvent(input))!
    expect(result).toMatch(/1 failed/)
  })
})

// ── WebFetch ──────────────────────────────────────────────────────────────────

describe('webfetch parity', () => {
  const input = fixture('webfetch-page.html')

  it('reduces output size', () => {
    const result = optimizeWebFetch(input)
    expect(result).not.toBeNull()
    expect(result!.length).toBeLessThan(input.length)
  })

  it('strips <script> blocks', () => {
    const result = optimizeWebFetch(input)!
    expect(result).not.toMatch(/<script/)
    expect(result).not.toMatch(/googletagmanager/)
    expect(result).not.toMatch(/intercomSettings/)
  })

  it('strips <style> blocks', () => {
    const result = optimizeWebFetch(input)!
    expect(result).not.toMatch(/<style/)
    expect(result).not.toMatch(/linear-gradient/)
  })

  it('strips <nav> blocks', () => {
    const result = optimizeWebFetch(input)!
    expect(result).not.toMatch(/<nav/)
  })

  it('strips <head> block', () => {
    const result = optimizeWebFetch(input)!
    expect(result).not.toMatch(/<head/)
    expect(result).not.toMatch(/googletagmanager\.com/)
  })

  it('strips <figure> and <picture> blocks', () => {
    const result = optimizeWebFetch(input)!
    expect(result).not.toMatch(/<figure/)
    expect(result).not.toMatch(/<picture/)
  })

  it('preserves main content text', () => {
    const result = optimizeWebFetch(input)!
    expect(result).toMatch(/Build faster, ship more/)
    expect(result).toMatch(/Intelligent optimizations/)
    expect(result).toMatch(/50\+ integrations/)
  })
})

// ── Generic ───────────────────────────────────────────────────────────────────

describe('generic parity', () => {
  const rawInput = fixture('generic-mcp.json')

  it('reduces output size', () => {
    const text = JSON.parse(rawInput).content[0].text as string
    const result = optimizeGeneric(text)
    expect(result).not.toBeNull()
    expect(result!.length).toBeLessThan(text.length)
  })

  it('strips avatar_url and gravatar_id', () => {
    const text = JSON.parse(rawInput).content[0].text as string
    const result = optimizeGeneric(text)!
    const parsed = JSON.parse(result)
    expect(parsed?.user?.avatar_url).toBeUndefined()
    expect(parsed?.user?.gravatar_id).toBeUndefined()
  })

  it('strips noise URL fields', () => {
    const text = JSON.parse(rawInput).content[0].text as string
    const result = optimizeGeneric(text)!
    const parsed = JSON.parse(result)
    expect(parsed?.diff_url).toBeUndefined()
    expect(parsed?.patch_url).toBeUndefined()
    expect(parsed?.issue_url).toBeUndefined()
    expect(parsed?.commits_url).toBeUndefined()
  })

  it('strips node_id fields', () => {
    const text = JSON.parse(rawInput).content[0].text as string
    const result = optimizeGeneric(text)!
    const parsed = JSON.parse(result)
    expect(parsed?.node_id).toBeUndefined()
    expect(parsed?.user?.node_id).toBeUndefined()
  })

  it('preserves meaningful fields', () => {
    const text = JSON.parse(rawInput).content[0].text as string
    const result = optimizeGeneric(text)!
    const parsed = JSON.parse(result)
    expect(parsed?.number).toBe(42)
    expect(parsed?.title).toMatch(/auth callback/)
    expect(parsed?.state).toBe('open')
  })

  it('optimizes JSON inside MCP content wrapper', () => {
    // Full MCP envelope — genericOptimizer should strip inside content[0].text
    const result = optimizeGeneric(rawInput)
    // optimizeGeneric only parses raw JSON strings, not MCP envelopes
    // The MCP envelope optimization is handled by Go's tryOptimizeContent
    // For TS this path goes through extractText → optimizeGeneric separately
    expect(result).toBeNull() // MCP envelope itself isn't stripped by generic
  })
})
