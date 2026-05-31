// Thin stdin→stdout runner used by Go integration parity tests.
// Usage: npx tsx src/optimizers/run.ts <optimizer> < input.txt
// Exits 0. Prints optimized text, or original if no optimization applies.

import { optimizeBash } from './bash.js'
import { optimizeWebFetch } from './webfetch.js'
import { optimizeGeneric } from './generic.js'
import { optimizeWebSearch } from './websearch.js'
import type { InterceptEvent } from '../types.js'

const name = process.argv[2] ?? ''
const chunks: Buffer[] = []
process.stdin.on('data', c => chunks.push(c))
process.stdin.on('end', () => {
  const input = Buffer.concat(chunks).toString('utf8')
  let result: string | null = null

  if (name === 'bash') {
    const event: InterceptEvent = {
      host: 'claude-code', strategy: 'hooks', phase: 'tool.post', session: { id: 'run' },
      tool: { name: 'Bash', output: input, input: {}, isMcp: false },
    }
    result = optimizeBash(input, event)
  } else if (name === 'webfetch') {
    result = optimizeWebFetch(input)
  } else if (name === 'generic') {
    result = optimizeGeneric(input)
  } else if (name === 'websearch') {
    result = optimizeWebSearch(input)
  } else {
    process.stderr.write(`unknown optimizer: ${name}\n`)
    process.exit(1)
  }

  process.stdout.write(result ?? input)
})
