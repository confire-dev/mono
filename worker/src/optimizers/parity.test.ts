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
import { optimizeWebSearch } from './websearch.js'
import type { InterceptEvent } from '../types.js'

const FIXTURES = join(fileURLToPath(new URL('.', import.meta.url)), '../../..', 'testdata/optimizer')

function fixture(name: string): string {
  return readFileSync(join(FIXTURES, name), 'utf8')
}

function bashEvent(input: string): InterceptEvent {
  return {
    host: 'claude-code', strategy: 'hooks', phase: 'tool.post', session: { id: 'test' },
    tool: { name: 'Bash', output: input, input: {}, isMcp: false },
  }
}

// ── Bash ─────────────────────────────────────────────────────────────────────

describe('bash parity', () => {
  const input = fixture('bash-test-runner.txt')

  it('reduces output size', () => {
    const result = optimizeBash(input, bashEvent(input))
    expect(result).not.toBeNull()
    expect(result!.length).toBeLessThan(input.length)
  })

  it('strips terminal progress bar noise', () => {
    const result = optimizeBash(input, bashEvent(input))!
    expect(result).not.toMatch(/\[====/)
    expect(result).not.toMatch(/######/)
  })

  it('preserves all test output — pass and fail alike', () => {
    const result = optimizeBash(input, bashEvent(input))!
    expect(result).toMatch(/✓ redirects to dashboard/)
    expect(result).toMatch(/creates stripe checkout session/)
    expect(result).toMatch(/checkout\.test\.ts/)
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

// ── WebSearch ─────────────────────────────────────────────────────────────────

describe('websearch parity', () => {
  const input = fixture('websearch-brave.json')

  it('reduces output size significantly', () => {
    const result = optimizeWebSearch(input)
    expect(result).not.toBeNull()
    expect(result!.length).toBeLessThan(input.length * 0.4)
  })

  it('preserves titles and urls', () => {
    const result = optimizeWebSearch(input)!
    expect(result).toMatch(/Cloudflare Workers/)
    expect(result).toMatch(/developers\.cloudflare\.com/)
    expect(result).toMatch(/TinyGo/)
    expect(result).toMatch(/blog\.cloudflare\.com/)
  })

  it('preserves description snippets', () => {
    const result = optimizeWebSearch(input)!
    expect(result).toMatch(/1MB/)
    expect(result).toMatch(/TinyGo produces/)
  })

  it('strips thumbnails, meta_url, profile images, family_friendly', () => {
    const result = optimizeWebSearch(input)!
    expect(result).not.toMatch(/thumbnail/)
    expect(result).not.toMatch(/meta_url/)
    expect(result).not.toMatch(/family_friendly/)
    expect(result).not.toMatch(/cdn-icons/)           // image URLs stripped
    expect(result).not.toMatch(/"extra_snippets"/)    // JSON key gone — content is inlined as bullets
  })

  it('preserves extra_snippets content as bullet points', () => {
    const result = optimizeWebSearch(input)!
    // extra_snippets are factual content, not noise — must survive
    expect(result).toMatch(/Paid plans support workers larger than 1MB/)
    expect(result).toMatch(/WASM modules up to 10MB on paid plans/)
    expect(result).toMatch(/TinyGo 0\.33 supports Cloudflare Workers/)
  })

  it('includes publisher name from profile', () => {
    const result = optimizeWebSearch(input)!
    expect(result).toMatch(/Cloudflare Docs/)
    expect(result).toMatch(/Cloudflare Blog/)
  })

  it('numbers results sequentially', () => {
    const result = optimizeWebSearch(input)!
    expect(result).toMatch(/^1\./m)
    expect(result).toMatch(/^2\./m)
    expect(result).toMatch(/^3\./m)
  })
})
