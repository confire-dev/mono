// Types shared across the platform.
// Mirror billing/usage types from Supabase schema and the Worker API.

export type Plan = 'free' | 'dev' | 'dev_annual' | 'pro' | 'pro_annual' | 'enterprise'

export interface UserProfile {
  id: string
  email: string
  name?: string
  plan: Plan
  subscription_status: 'none' | 'trialing' | 'active' | 'past_due' | 'canceled'
  stripe_customer_id?: string
}

export interface CreditBalance {
  included_credits: number
  bonus_credits: number
  purchased_credits: number
  total_credits: number
}

export interface UsageSummary {
  request_count: number
  saved_bytes: number
  month: string
}

export interface ToolCallSummary {
  id: string
  tool_type: string
  integration: string
  optimizer: string
  raw_bytes: number
  optimized_bytes: number
  saved_bytes: number
  reduction_ratio: number
  duration_ms?: number
  was_cached: boolean
  created_at: string
  // Derived
  mode: 'local' | 'remote'
}

export interface CliSession {
  id: string
  device_id?: string
  cli_version?: string
  integration: string
  started_at: string
  ended_at?: string
  total_tool_calls: number
  optimized_calls: number
  raw_bytes: number
  optimized_bytes: number
  // Derived
  saved_bytes: number
  is_active: boolean
}

export interface ApiKey {
  id: string
  key_prefix: string
  key_suffix?: string
  device_id?: string
  last_used_at?: string
  created_at: string
}

// Token savings display (229k → 4.7k)
export interface SavingsDisplay {
  rawTokens: number
  optimizedTokens: number
  reductionPct: number
  label: string
}

// Roughly 4 bytes per token for display purposes
export function bytesToTokens(bytes: number): number {
  return Math.round(bytes / 4)
}

export function formatTokenCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000)     return `${(n / 1_000).toFixed(1)}k`
  return String(n)
}
