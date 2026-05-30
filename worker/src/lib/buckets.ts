// Amplitude bucketing helpers.
//
// All values forwarded to Amplitude are bucketed so:
//   1. No raw user content ever reaches Amplitude.
//   2. Individual sessions can't be fingerpainted from metric values.
//   3. Amplitude is useful for cohort analysis ("did users with >80% savings convert?")
//      without storing personally-identifiable usage detail.
//
// Supabase receives exact values for the user's dashboard.

export function byteBucket(n: number): string {
  if (n < 1_000)     return '<1k'
  if (n < 10_000)    return '1k_10k'
  if (n < 50_000)    return '10k_50k'
  if (n < 100_000)   return '50k_100k'
  if (n < 500_000)   return '100k_500k'
  return '>500k'
}

export function reductionBucket(ratio: number): string {
  const pct = ratio * 100
  if (pct < 20)  return '0_20'
  if (pct < 50)  return '20_50'
  if (pct < 80)  return '50_80'
  if (pct < 90)  return '80_90'
  if (pct < 95)  return '90_95'
  return '>95'
}

export function durationBucket(ms: number): string {
  if (ms < 10)  return '<10ms'
  if (ms < 50)  return '10_50ms'
  if (ms < 200) return '50_200ms'
  if (ms < 1000) return '200ms_1s'
  return '>1s'
}

// sanitizeToolType ensures no raw tool names with sensitive info reach Amplitude.
// Maps to a safe canonical category.
export function sanitizeToolType(toolType: string): string {
  const known = ['bash','read','webfetch','glob','grep','edit','write',
    'figma','github','atlassian','clickup','slack','amplitude','fireflies',
    'notion','playwright','zapier','google_drive']
  const lower = toolType.toLowerCase()
  for (const k of known) {
    if (lower.includes(k)) return k
  }
  return 'other'
}
