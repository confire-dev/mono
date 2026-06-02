export type Risk = 'HIGH' | 'MEDIUM' | 'LOW' | 'NONE'

export interface ClassifyResult {
  risk: Risk
  category: string | null
  matchedPattern: string | null
}

const NONE: ClassifyResult = { risk: 'NONE', category: null, matchedPattern: null }

export function classify(_payload: string): ClassifyResult {
  return NONE
}
