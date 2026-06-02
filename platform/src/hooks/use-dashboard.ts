import { useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
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

const PLAN_NAMES: Record<Plan, string> = {
  free:       'Free',
  dev:        'Dev',
  dev_annual: 'Dev (Annual)',
  pro:        'Pro',
  pro_annual: 'Pro (Annual)',
  enterprise: 'Enterprise',
}

// ── fetchers ──────────────────────────────────────────────────────────────────

async function fetchMe(workerBase: string, token: string): Promise<MeData> {
  const res = await fetch(`${workerBase}/api/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error('Failed to load account')
  const d = await res.json() as {
    email: string; name?: string; plan: Plan; subscription_status: string;
    used: number; limit: number; purchasedCredits: number; effectiveLimit: number;
    limits: PlanLimits; features: PlanFeatures
  }
  return {
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
  }
}

async function fetchFirewall(workerBase: string, token: string): Promise<FirewallGroup[]> {
  const res = await fetch(`${workerBase}/v1/policy`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  const overrides: Record<string, boolean> = res.ok
    ? ((await res.json() as any).group_overrides ?? {})
    : {}
  return Object.entries(GROUP_LABELS).map(([id, label]) => ({
    id, label, enabled: overrides[id] !== false,
  }))
}

// ── hook ──────────────────────────────────────────────────────────────────────

export function useDashboard(apiKey: string | null, workerBase: string) {
  const supabase    = useMemo(() => getSupabaseClient(), [])
  const queryClient = useQueryClient()

  // Resolve Supabase user ID once — used as cache key for all Supabase queries
  const { data: uid } = useQuery({
    queryKey:  ['uid'],
    queryFn:   async () => (await supabase.auth.getUser()).data.user?.id ?? null,
    staleTime: Infinity,
    enabled:   !!apiKey,
  })

  const meQuery = useQuery({
    queryKey: ['me', apiKey],
    queryFn:  () => fetchMe(workerBase, apiKey!),
    enabled:  !!apiKey,
  })

  const firewallQuery = useQuery({
    queryKey: ['firewall', apiKey],
    queryFn:  () => fetchFirewall(workerBase, apiKey!),
    enabled:  !!apiKey,
  })

  const sessionsQuery = useQuery({
    queryKey: ['sessions', uid],
    queryFn:  async () => {
      const { data } = await supabase
        .from('cli_sessions')
        .select('id,device_id,cli_version,integration,started_at,ended_at,total_tool_calls,optimized_calls,raw_bytes,optimized_bytes')
        .eq('user_id', uid!)
        .order('started_at', { ascending: false })
        .limit(10)
      return (data ?? []).map((s: any) => ({
        ...s,
        saved_bytes: s.raw_bytes - s.optimized_bytes,
        is_active:   !s.ended_at,
      })) as CliSession[]
    },
    enabled: !!uid,
  })

  const recentCallsQuery = useQuery({
    queryKey: ['recent-calls', uid],
    queryFn:  async () => {
      const { data } = await supabase
        .from('tool_call_summaries')
        .select('id,tool_type,integration,optimizer,raw_bytes,optimized_bytes,was_cached,created_at')
        .eq('user_id', uid!)
        .order('created_at', { ascending: false })
        .limit(20)
      return (data ?? []).map((c: any) => ({
        ...c,
        saved_bytes:     c.raw_bytes - c.optimized_bytes,
        reduction_ratio: c.raw_bytes > 0 ? Math.round((1 - c.optimized_bytes / c.raw_bytes) * 100) : 0,
        mode:            c.optimizer?.startsWith('local/') ? 'local' : 'remote',
      })) as ToolCallSummary[]
    },
    enabled: !!uid,
  })

  const allCallBytesQuery = useQuery({
    queryKey: ['call-bytes', uid],
    queryFn:  async () => {
      const { data } = await supabase
        .from('tool_call_summaries')
        .select('raw_bytes,optimized_bytes,created_at')
        .eq('user_id', uid!)
      return (data ?? []) as { raw_bytes: number; optimized_bytes: number; created_at: string }[]
    },
    enabled: !!uid,
  })

  const apiKeysQuery = useQuery({
    queryKey: ['api-keys', uid],
    queryFn:  async () => {
      const { data } = await supabase
        .from('api_keys')
        .select('id,key_prefix,key_suffix,device_id,last_used_at,created_at')
        .eq('user_id', uid!)
        .is('revoked_at', null)
        .order('created_at', { ascending: false })
      return (data ?? []) as ApiKey[]
    },
    enabled: !!uid,
  })

  // ── mutations ──────────────────────────────────────────────────────────────

  const revokeKey = useMutation({
    mutationFn: async (keyId: string) => {
      const res = await fetch(`${workerBase}/api/keys/revoke`, {
        method:  'POST',
        headers: { Authorization: `Bearer ${apiKey}`, 'Content-Type': 'application/json' },
        body:    JSON.stringify({ keyId }),
      })
      if (!res.ok) throw new Error('Revoke failed')
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['api-keys', uid] }),
  })

  const toggleFirewallGroup = useMutation({
    mutationFn: async ({ id, enabled }: { id: string; enabled: boolean }) => {
      const res = await fetch(`${workerBase}/v1/policy/groups`, {
        method:  'PATCH',
        headers: { Authorization: `Bearer ${apiKey}`, 'Content-Type': 'application/json' },
        body:    JSON.stringify({ [id]: enabled }),
      })
      if (!res.ok) throw new Error('Toggle failed')
    },
    onMutate: async ({ id, enabled }) => {
      // Optimistic update
      await queryClient.cancelQueries({ queryKey: ['firewall', apiKey] })
      const prev = queryClient.getQueryData<FirewallGroup[]>(['firewall', apiKey])
      queryClient.setQueryData<FirewallGroup[]>(['firewall', apiKey],
        old => old?.map(g => g.id === id ? { ...g, enabled } : g) ?? [])
      return { prev }
    },
    onError: (_err, _vars, ctx) => {
      // Roll back on failure
      queryClient.setQueryData(['firewall', apiKey], ctx?.prev)
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: ['firewall', apiKey] }),
  })

  // ── derived stats ──────────────────────────────────────────────────────────

  const allCallBytes    = allCallBytesQuery.data ?? []
  const totalSavedBytes = allCallBytes.reduce((s, c) => s + Math.max(0, c.raw_bytes - c.optimized_bytes), 0)

  const loading = !apiKey
    ? false
    : meQuery.isLoading || firewallQuery.isLoading || apiKeysQuery.isLoading

  const error = meQuery.error?.message
    ?? firewallQuery.error?.message
    ?? sessionsQuery.error?.message
    ?? null

  return {
    me:           meQuery.data ?? null,
    sessions:     sessionsQuery.data ?? [],
    recentCalls:  recentCallsQuery.data ?? [],
    allCallBytes,
    apiKeys:      apiKeysQuery.data ?? [],
    firewall:     firewallQuery.data ?? [],
    loading,
    error,
    totalSavedBytes,
    totalSavedTokens:  Math.round(totalSavedBytes / 4),
    totalCallsAllTime: allCallBytes.length,
    activeSessions:    (sessionsQuery.data ?? []).filter(s => s.is_active).length,
    // mutations
    revokeKey,
    toggleFirewallGroup,
  }
}
