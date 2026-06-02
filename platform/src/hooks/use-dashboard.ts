// useDashboard — fetches all data needed to render the dashboard.
//
// Sources:
//   /api/me        — profile, plan, usage period, credit balance
//   /v1/policy     — firewall group overrides
//   Supabase       — recent CLI sessions, recent tool call summaries, API keys
//
// All fetches run in parallel. Individual failures return nulls so the page
// degrades gracefully rather than showing an error for everything.

import { useState, useEffect, useMemo } from 'react'
import { getSupabaseClient } from '@/lib/supabase-client'
import type { CliSession, ToolCallSummary, ApiKey } from '@/lib/types'

export type Plan = 'free' | 'dev' | 'dev_annual' | 'pro' | 'pro_annual' | 'enterprise'

export interface PlanLimits {
  cloudOptimizationsMonthly: number
  cloudTokensMonthly: number
  retainedHistoryDays: number
  cliSessions: number
}

export interface PlanFeatures {
  localOptimization: boolean
  remoteOptimization: boolean
  usageDashboard: boolean
  advancedUsageDashboard: boolean
  optimizationHistory: boolean
  sessionMemoryGuard: boolean
  firewallGroupToggles: boolean
  [key: string]: boolean
}

export interface MeData {
  email: string
  name?: string
  plan: Plan
  planName: string
  subscriptionStatus: string
  used: number
  limit: number
  purchasedCredits: number
  effectiveLimit: number
  limits: PlanLimits
  features: PlanFeatures
}

export interface FirewallGroup {
  id: string
  label: string
  enabled: boolean
}

const GROUP_LABELS: Record<string, string> = {
  'mcp.risk_classifier': 'Risk Classifier',
  'mcp.secrets':         'Secret Redaction',
  'mcp.injection':       'Injection Detection',
  'mcp.unicode':         'Unicode Sanitizer',
  'mcp.normalizer':      'Output Normalizer',
  'mcp.budget':          'Token Budget',
  'mcp.cross_tool':      'Cross-Tool Prevention',
}

export interface DashboardData {
  me:       MeData | null
  sessions: CliSession[]
  recentCalls: ToolCallSummary[]
  apiKeys:  ApiKey[]
  firewall: FirewallGroup[]
  loading:  boolean
  error:    string | null
  // Derived stats
  totalSavedBytes:   number
  totalSavedTokens:  number
  totalCallsAllTime: number
  activeSessions:    number
}

const PLAN_NAMES: Record<Plan, string> = {
  free:       'Free',
  dev:        'Dev',
  dev_annual: 'Dev (Annual)',
  pro:        'Pro',
  pro_annual: 'Pro (Annual)',
  enterprise: 'Enterprise',
}

export function useDashboard(apiKey: string | null): DashboardData {
  const supabase = useMemo(() => getSupabaseClient(), [])

  const [me,           setMe]           = useState<MeData | null>(null)
  const [sessions,     setSessions]     = useState<CliSession[]>([])
  const [recentCalls,  setRecentCalls]  = useState<ToolCallSummary[]>([])
  const [apiKeys,      setApiKeys]      = useState<ApiKey[]>([])
  const [firewall,     setFirewall]     = useState<FirewallGroup[]>([])
  const [loading,      setLoading]      = useState(true)
  const [error,        setError]        = useState<string | null>(null)

  useEffect(() => {
    if (!apiKey) {
      setLoading(false)
      return
    }

    const headers = { Authorization: `Bearer ${apiKey}` }

    const workerBase = import.meta.env.PUBLIC_WORKER_URL ?? ''

    async function load() {
      try {
        const [meRes, policyRes] = await Promise.allSettled([
          fetch(`${workerBase}/api/me`, { headers }),
          fetch(`${workerBase}/v1/policy`, { headers }),
        ])

        // /api/me
        if (meRes.status === 'fulfilled' && meRes.value.ok) {
          const d = await meRes.value.json() as {
            email: string; name?: string; plan: Plan; subscription_status: string;
            used: number; limit: number; purchasedCredits: number; effectiveLimit: number;
            limits: PlanLimits; features: PlanFeatures
          }
          setMe({
            email:              d.email,
            name:               d.name,
            plan:               d.plan,
            planName:           PLAN_NAMES[d.plan] ?? d.plan,
            subscriptionStatus: d.subscription_status,
            used:               d.used,
            limit:              d.limit,
            purchasedCredits:   d.purchasedCredits,
            effectiveLimit:     d.effectiveLimit,
            limits:             d.limits,
            features:           d.features,
          })
        }

        // /v1/policy group overrides
        if (policyRes.status === 'fulfilled' && policyRes.value.ok) {
          const d = await policyRes.value.json() as { group_overrides: Record<string, boolean> }
          const overrides = d.group_overrides ?? {}
          setFirewall(
            Object.entries(GROUP_LABELS).map(([id, label]) => ({
              id,
              label,
              enabled: overrides[id] !== false, // default on
            }))
          )
        } else {
          // No overrides — all groups on by default
          setFirewall(
            Object.entries(GROUP_LABELS).map(([id, label]) => ({ id, label, enabled: true }))
          )
        }

        // Supabase direct queries
        const uid = (await supabase.auth.getUser()).data.user?.id
        if (uid) {
          const [sessRes, callsRes, keysRes] = await Promise.allSettled([
            supabase
              .from('cli_sessions')
              .select('id,device_id,cli_version,integration,started_at,ended_at,total_tool_calls,optimized_calls,raw_bytes,optimized_bytes')
              .eq('user_id', uid)
              .order('started_at', { ascending: false })
              .limit(10),

            supabase
              .from('tool_call_summaries')
              .select('id,tool_type,integration,optimizer,raw_bytes,optimized_bytes,was_cached,created_at')
              .eq('user_id', uid)
              .order('created_at', { ascending: false })
              .limit(20),

            supabase
              .from('api_keys')
              .select('id,key_prefix,device_id,last_used_at,created_at')
              .eq('user_id', uid)
              .is('revoked_at', null)
              .order('created_at', { ascending: false }),
          ])

          if (sessRes.status === 'fulfilled' && sessRes.value.data) {
            setSessions(
              sessRes.value.data.map((s: any) => ({
                ...s,
                saved_bytes: s.raw_bytes - s.optimized_bytes,
                is_active:   !s.ended_at,
              }))
            )
          }

          if (callsRes.status === 'fulfilled' && callsRes.value.data) {
            setRecentCalls(
              callsRes.value.data.map((c: any) => ({
                ...c,
                saved_bytes:      c.raw_bytes - c.optimized_bytes,
                reduction_ratio:  c.raw_bytes > 0
                  ? Math.round((1 - c.optimized_bytes / c.raw_bytes) * 100)
                  : 0,
                mode: c.optimizer?.startsWith('local/') ? 'local' : 'remote',
              }))
            )
          }

          if (keysRes.status === 'fulfilled' && keysRes.value.data) {
            setApiKeys(keysRes.value.data)
          }
        }
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load dashboard')
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [apiKey, supabase])

  // Derived stats
  const totalSavedBytes   = sessions.reduce((s, x) => s + x.saved_bytes, 0)
  const totalSavedTokens  = Math.round(totalSavedBytes / 4)
  const totalCallsAllTime = sessions.reduce((s, x) => s + x.total_tool_calls, 0)
  const activeSessions    = sessions.filter(s => s.is_active).length

  return {
    me, sessions, recentCalls, apiKeys, firewall,
    loading, error,
    totalSavedBytes, totalSavedTokens, totalCallsAllTime, activeSessions,
  }
}
