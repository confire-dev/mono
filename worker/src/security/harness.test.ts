import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { classify } from './classifier.js'

const FIXTURES = join(fileURLToPath(new URL('.', import.meta.url)), '../../..', 'testdata/security')

interface AttackCase {
  id: string
  category: string
  description: string
  payload: string
  risk: string
  should_catch: boolean
  notes: string
}

interface BenignCase {
  id: string
  tool: string
  content: string
  notes: string
}

function loadJsonFiles(dir: string): string[] {
  const paths: string[] = []
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry)
    if (statSync(full).isDirectory()) {
      paths.push(...loadJsonFiles(full))
    } else if (entry.endsWith('.json')) {
      paths.push(full)
    }
  }
  return paths
}

function loadAttacks(): AttackCase[] {
  const attackDir = join(FIXTURES, 'attacks')
  return loadJsonFiles(attackDir).flatMap(f =>
    JSON.parse(readFileSync(f, 'utf8')) as AttackCase[]
  )
}

function loadBenign(): BenignCase[] {
  const benignDir = join(FIXTURES, 'benign')
  return loadJsonFiles(benignDir).flatMap(f =>
    JSON.parse(readFileSync(f, 'utf8')) as BenignCase[]
  )
}

describe('attack detection', () => {
  const attacks = loadAttacks()

  for (const attack of attacks) {
    it(attack.id, () => {
      const result = classify(attack.payload)
      if (attack.should_catch) {
        expect(result.risk, `missed ${attack.id}: ${attack.description}`).not.toBe('NONE')
      }
    })
  }
})

describe('false positives', () => {
  const benign = loadBenign()
  const FALSE_POSITIVE_THRESHOLD = 0.01

  it('false positive rate is below 1%', () => {
    let flagged = 0
    for (const b of benign) {
      const result = classify(b.content)
      if (result.risk !== 'NONE') flagged++
    }
    const rate = flagged / benign.length
    expect(
      rate,
      `false positive rate ${(rate * 100).toFixed(1)}% exceeds 1% (${flagged}/${benign.length} flagged)`
    ).toBeLessThanOrEqual(FALSE_POSITIVE_THRESHOLD)
  })

  for (const b of benign) {
    it(b.id, () => {
      const result = classify(b.content)
      expect(result.risk, `false positive on ${b.id} (${b.tool}): ${b.notes}`).toBe('NONE')
    })
  }
})
