import { useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getSupabaseClient } from '@/lib/supabase-client'
import type { CliSession, SecurityEvent, ApiKey } from '@/lib/types'

export type Plan = 'free' | 'dev' | 'team' | 'enterprise'

export interface PlanLimits {
  maxPayloadBytes: number
  retainedHistoryDays: number
  cliSessions: number
  maxCustomRules: number
}

export interface PlanFeatures {
  firewallEnabled: boolean
  customRules: boolean
  policySync: boolean
  usageDashboard: boolean
  advancedUsageDashboard: boolean
  cliSessionManagement: boolean
  exportData: boolean
  ssoSaml: boolean
  firewallGroupToggles: boolean
  securityEventHistory: boolean
  provenanceTracking: boolean
  sharedPolicies: boolean
  teamDashboard: boolean
  auditLogs: boolean
  agentInventory: boolean
  fleetControls: boolean
  [key: string]: boolean
}

export interface MeData {
  email: string
  name?: string
  plan: Plan
  planName: string
  subscriptionStatus: string
  billingInterval: 'monthly' | 'annual' | null
  periodEnd: string | null
  cancelAtPeriodEnd: boolean
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
  team:       'Team',
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
    billing_interval: 'monthly' | 'annual' | null;
    subscription_current_period_end: string | null;
    cancel_at_period_end: boolean;
    limits: PlanLimits; features: PlanFeatures
  }
  return {
    email:              d.email,
    name:               d.name,
    plan:               d.plan,
    planName:           PLAN_NAMES[d.plan] ?? d.plan,
    subscriptionStatus: d.subscription_status,
    billingInterval:    d.billing_interval ?? null,
    periodEnd:          d.subscription_current_period_end ?? null,
    cancelAtPeriodEnd:  d.cancel_at_period_end ?? false,
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
        .select('id,device_id,cli_version,integration,started_at,ended_at,total_tool_calls')
        .eq('user_id', uid!)
        .order('started_at', { ascending: false })
        .limit(10)
      return (data ?? []).map((s: any) => ({
        ...s,
        is_active: !s.ended_at,
      })) as CliSession[]
    },
    enabled: !!uid,
  })

  const securityEventsQuery = useQuery({
    queryKey: ['security-events', uid],
    queryFn:  async () => {
      const { data } = await supabase
        .from('security_events')
        .select('id,tool_name,event_type,risk_level,action_taken,pattern_matched,bypassed,created_at')
        .eq('user_id', uid!)
        .order('created_at', { ascending: false })
        .limit(50)
      return (data ?? []) as SecurityEvent[]
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

  const securityEvents = securityEventsQuery.data ?? []
  const protectedActions = securityEvents.filter(e => e.action_taken !== 'ALLOWED').length
  const threatsBlocked   = securityEvents.filter(e => e.action_taken === 'BLOCKED').length
  const secretsRedacted  = securityEvents.filter(e => e.event_type === 'SECRET_REDACTED').length
  const highRiskFlagged  = securityEvents.filter(e => e.risk_level === 'HIGH' || e.risk_level === 'CRITICAL').length

  const loading = !apiKey
    ? false
    : meQuery.isLoading || firewallQuery.isLoading || apiKeysQuery.isLoading

  const error = meQuery.error?.message
    ?? firewallQuery.error?.message
    ?? sessionsQuery.error?.message
    ?? null

  return {
    me:              meQuery.data ?? null,
    sessions:        sessionsQuery.data ?? [],
    securityEvents,
    apiKeys:         apiKeysQuery.data ?? [],
    firewall:        firewallQuery.data ?? [],
    loading,
    error,
    protectedActions,
    threatsBlocked,
    secretsRedacted,
    highRiskFlagged,
    activeSessions: (sessionsQuery.data ?? []).filter(s => s.is_active).length,
    // mutations
    revokeKey,
    toggleFirewallGroup,
  }
}
