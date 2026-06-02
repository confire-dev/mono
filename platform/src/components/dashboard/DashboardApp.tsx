"use client"

import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as echarts from 'echarts/core'
import { LineChart, BarChart as EBarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, BrushComponent, ToolboxComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { Button, Text, TimeseriesChart } from '@cloudflare/kumo'
import { useAuth } from '@/hooks/use-auth'
import { useDashboard } from '@/hooks/use-dashboard'
import { formatTokenCount } from '@/lib/types'

echarts.use([LineChart, EBarChart, GridComponent, TooltipComponent, BrushComponent, ToolboxComponent, CanvasRenderer])

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 30_000 } },
})

// ── helpers ───────────────────────────────────────────────────────────────────

function timeAgo(iso: string): string {
  const diff = (Date.now() - new Date(iso).getTime()) / 1000
  if (diff < 60)    return `${Math.floor(diff)}s ago`
  if (diff < 3600)  return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

function fmtBytes(n: number): string {
  if (n >= 1_048_576) return `${(n / 1_048_576).toFixed(1)} MB`
  if (n >= 1_024)     return `${(n / 1_024).toFixed(1)} KB`
  return `${n} B`
}

const INTEGRATION_COLOR: Record<string, string> = {
  'claude-code': '#f4811f',
  cursor:        '#7b68ee',
  vscode:        '#007acc',
  default:       '#6b7280',
}

// ── sub-components ────────────────────────────────────────────────────────────

function StatCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div style={{
      background: 'var(--confire-surface)',
      border: '1px solid var(--confire-border)',
      borderRadius: 8,
      padding: '20px 24px',
    }}>
      <div style={{ fontSize: 13, color: 'var(--confire-text-muted)', marginBottom: 8 }}>{label}</div>
      <div style={{ fontSize: 28, fontWeight: 600, color: 'var(--confire-text)', lineHeight: 1 }}>{value}</div>
      {sub && <div style={{ fontSize: 12, color: 'var(--confire-text-muted)', marginTop: 6 }}>{sub}</div>}
    </div>
  )
}

function SectionHeader({ title }: { title: string }) {
  return (
    <div style={{
      fontSize: 13,
      fontWeight: 600,
      color: 'var(--confire-text-muted)',
      textTransform: 'uppercase',
      letterSpacing: '0.06em',
      marginBottom: 12,
    }}>{title}</div>
  )
}

function Card({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <div style={{
      background: 'var(--confire-surface)',
      border: '1px solid var(--confire-border)',
      borderRadius: 8,
      overflow: 'hidden',
      ...style,
    }}>
      {children}
    </div>
  )
}

function CardHeader({ title, action }: { title: string; action?: React.ReactNode }) {
  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '16px 20px',
      borderBottom: '1px solid var(--confire-border)',
    }}>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--confire-text)' }}>{title}</span>
      {action}
    </div>
  )
}

function EmptyRow({ label }: { label: string }) {
  return (
    <div style={{ padding: '24px 20px', textAlign: 'center', color: 'var(--confire-text-muted)', fontSize: 13 }}>
      {label}
    </div>
  )
}

function PlanBadge({ plan }: { plan: string }) {
  const color = plan === 'free' ? '#6b7280' : plan.startsWith('enterprise') ? '#a78bfa' : '#f4811f'
  return (
    <span style={{
      background: color + '22',
      color,
      border: `1px solid ${color}44`,
      borderRadius: 4,
      fontSize: 11,
      fontWeight: 600,
      padding: '2px 8px',
      textTransform: 'uppercase',
      letterSpacing: '0.05em',
    }}>{plan.replace('_', ' ')}</span>
  )
}

// ── usage chart ──────────────────────────────────────────────────────────────

function UsageChart({ calls }: { calls: { raw_bytes: number; optimized_bytes: number; created_at: string }[] }) {
  // Bucket by UTC day → tokens saved
  const byDay = new Map<string, number>()
  for (const c of calls) {
    const day = c.created_at.slice(0, 10) // "YYYY-MM-DD"
    byDay.set(day, (byDay.get(day) ?? 0) + Math.max(0, Math.round((c.raw_bytes - c.optimized_bytes) / 4)))
  }
  const data: [number, number][] = Array.from(byDay.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([day, tokens]) => [new Date(day).getTime(), tokens])

  if (data.length === 0) return null

  return (
    <Card style={{ gridColumn: '1 / -1' }}>
      <CardHeader title="Tokens Saved — Daily" />
      <div style={{ padding: '12px 8px 4px' }}>
        <TimeseriesChart
          echarts={echarts}
          type="line"
          data={[{ name: 'Tokens saved', data, color: 'var(--confire-accent, #f4811f)' }]}
          tooltipValueFormat={v => `${Math.round(v).toLocaleString()} tokens`}
          height={180}
        />
      </div>
    </Card>
  )
}

// ── devices section ──────────────────────────────────────────────────────────


function DevicesCard({
  apiKeys,
  revokeKey,
}: {
  apiKeys: { id: string; key_prefix: string; key_suffix?: string; device_id?: string; last_used_at?: string; created_at: string }[]
  revokeKey: { mutate: (id: string) => void; isPending: boolean; variables?: string }
}) {
  const visible = apiKeys

  return (
    <Card>
      <CardHeader title="Devices" />
      {visible.length === 0 ? (
        <EmptyRow label="No devices connected — install the CLI to get started" />
      ) : (
        visible.map((k, i) => {
          const name     = k.device_id || k.key_prefix
          const keyHint  = k.key_suffix ? `••••${k.key_suffix}` : k.key_prefix.slice(0, 8) + '••••'
          const lastSeen = k.last_used_at ?? k.created_at
          return (
            <div key={k.id} style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '12px 20px',
              borderBottom: i < visible.length - 1 ? '1px solid var(--confire-border)' : undefined,
            }}>
              <div>
                <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--confire-text)' }}>{name}</div>
                <div style={{ fontSize: 11, color: 'var(--confire-text-muted)', marginTop: 2, fontFamily: 'monospace' }}>
                  {keyHint} · last active {timeAgo(lastSeen)}
                </div>
              </div>
              <button
                disabled={revokeKey.isPending && revokeKey.variables === k.id}
                onClick={() => revokeKey.mutate(k.id)}
                style={{
                  fontSize: 12,
                  color: '#f87171',
                  background: 'transparent',
                  border: '1px solid #f8717144',
                  borderRadius: 4,
                  padding: '3px 10px',
                  cursor: 'pointer',
                  opacity: revokeKey.isPending && revokeKey.variables === k.id ? 0.5 : 1,
                }}
              >
                {revokeKey.isPending && revokeKey.variables === k.id ? '…' : 'Revoke'}
              </button>
            </div>
          )
        })
      )}
    </Card>
  )
}

// ── firewall section ──────────────────────────────────────────────────────────

function FirewallSection({
  groups,
  canToggle,
  toggle,
}: {
  groups: { id: string; label: string; enabled: boolean }[]
  canToggle: boolean
  toggle: (id: string, enabled: boolean) => void
}) {
  return (
    <Card>
      <CardHeader title="MCP Firewall" />
      {groups.map((g, i) => {
        return (
          <div key={g.id} style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '12px 20px',
            borderBottom: i < groups.length - 1 ? '1px solid var(--confire-border)' : undefined,
          }}>
            <div>
              <div style={{ fontSize: 13, color: 'var(--confire-text)', fontWeight: 500 }}>{g.label}</div>
              <div style={{ fontSize: 11, color: 'var(--confire-text-muted)', marginTop: 2 }}>{g.id}</div>
            </div>
            <button
              disabled={!canToggle}
              onClick={() => toggle(g.id, !g.enabled)}
              style={{
                width: 36,
                height: 20,
                borderRadius: 10,
                border: 'none',
                background: g.enabled ? 'var(--confire-accent)' : 'var(--confire-border)',
                cursor: canToggle ? 'pointer' : 'not-allowed',
                position: 'relative',
                transition: 'background 0.15s',
                flexShrink: 0,
              }}
            >
              <span style={{
                position: 'absolute',
                top: 2,
                left: g.enabled ? 18 : 2,
                width: 16,
                height: 16,
                borderRadius: '50%',
                background: '#fff',
                transition: 'left 0.15s',
              }} />
            </button>
          </div>
        )
      })}
      {!canToggle && (
        <div style={{ padding: '10px 20px', borderTop: '1px solid var(--confire-border)', fontSize: 12, color: 'var(--confire-text-muted)' }}>
          Firewall group toggles are available on Dev and Pro plans.{' '}
          <a href="/pricing" style={{ color: 'var(--confire-accent)' }}>Upgrade →</a>
        </div>
      )}
    </Card>
  )
}

// ── main component ────────────────────────────────────────────────────────────

function Dashboard() {
  const { user, session, loading: authLoading, signOut } = useAuth()
  const apiKey = session?.access_token ?? null
  const workerBase = (import.meta as any).env?.PUBLIC_WORKER_URL ?? ''

  const {
    me, recentCalls, allCallBytes, apiKeys, firewall,
    loading, error,
    totalSavedTokens, totalCallsAllTime,
    revokeKey, toggleFirewallGroup,
  } = useDashboard(apiKey, workerBase)

  if (authLoading || loading) {
    return (
      <div style={{ display: 'flex', minHeight: '100svh', alignItems: 'center', justifyContent: 'center' }}>
        <Text variant="secondary" size="sm">Loading…</Text>
      </div>
    )
  }

  if (!user) {
    window.location.href = '/login'
    return null
  }

  const usedPct = me ? Math.min(100, Math.round((me.used / me.effectiveLimit) * 100)) : 0
  const canToggle = !!(me?.features?.firewallGroupToggles)

  const accent = 'var(--confire-accent, #f4811f)'

  return (
    <div style={{ minHeight: '100svh', background: 'var(--confire-bg, #0d1117)', color: 'var(--confire-text, #e6edf3)' }}>
      {/* nav */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 32px',
        height: 56,
        borderBottom: '1px solid var(--confire-border, #21262d)',
        background: 'var(--confire-surface, #161b22)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span style={{ fontSize: 18, fontWeight: 700, color: accent }}>⬡ Confire</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          {me && <PlanBadge plan={me.plan} />}
          <span style={{ fontSize: 13, color: 'var(--confire-text-muted, #8b949e)' }}>{user.email}</span>
          <Button variant="outline" size="sm" onClick={signOut}>Sign out</Button>
        </div>
      </div>

      {/* page header */}
      <div style={{ padding: '32px 32px 0' }}>
        <div style={{ fontSize: 22, fontWeight: 700, marginBottom: 4 }}>Dashboard</div>
        <div style={{ fontSize: 14, color: 'var(--confire-text-muted, #8b949e)' }}>
          {me ? `${me.planName} plan · ${me.subscriptionStatus}` : 'Overview'}
        </div>
      </div>

      <div style={{ padding: '24px 32px', display: 'flex', flexDirection: 'column', gap: 24, maxWidth: 1200 }}>

        {error && (
          <div style={{ background: '#3b1c1c', border: '1px solid #6b2d2d', borderRadius: 8, padding: '12px 16px', fontSize: 13, color: '#f87171' }}>
            {error}
          </div>
        )}

        {/* stat grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
          <StatCard
            label="Tokens saved (all-time)"
            value={formatTokenCount(totalSavedTokens)}
            sub="across all tool calls"
          />
          <StatCard
            label="Total tool calls"
            value={totalCallsAllTime.toLocaleString()}
            sub="processed by Confire"
          />
          <StatCard
            label="Usage this period"
            value={me ? `${me.used.toLocaleString()} / ${me.effectiveLimit.toLocaleString()}` : '—'}
            sub={me ? `${usedPct}% of limit` : undefined}
          />
        </div>

        {/* usage chart */}
        <UsageChart calls={allCallBytes} />

        {/* usage bar + firewall row */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, alignItems: 'start' }}>

          {/* usage bar */}
          <Card>
            <CardHeader title="Usage" />
            <div style={{ padding: '20px 20px 16px' }}>
              {me ? (
                <>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, marginBottom: 8 }}>
                    <span style={{ color: 'var(--confire-text-muted)' }}>Cloud optimizations</span>
                    <span style={{ color: 'var(--confire-text)', fontWeight: 500 }}>
                      {me.used.toLocaleString()} / {me.effectiveLimit.toLocaleString()}
                    </span>
                  </div>
                  <div style={{ height: 6, background: 'var(--confire-border)', borderRadius: 3, overflow: 'hidden' }}>
                    <div style={{
                      height: '100%',
                      width: `${usedPct}%`,
                      background: usedPct >= 80 ? '#f87171' : accent,
                      borderRadius: 3,
                      transition: 'width 0.3s',
                    }} />
                  </div>
                  {me.purchasedCredits > 0 && (
                    <div style={{ fontSize: 12, color: 'var(--confire-text-muted)', marginTop: 8 }}>
                      +{me.purchasedCredits.toLocaleString()} purchased credits available
                    </div>
                  )}
                  {usedPct >= 80 && (
                    <div style={{ marginTop: 12, padding: '10px 12px', background: '#f4811f11', border: '1px solid #f4811f33', borderRadius: 6, fontSize: 12, color: '#f4811f' }}>
                      You're at {usedPct}% of your limit.{' '}
                      <a href="/pricing" style={{ color: accent, fontWeight: 600 }}>Upgrade or buy credits →</a>
                    </div>
                  )}
                </>
              ) : (
                <div style={{ color: 'var(--confire-text-muted)', fontSize: 13 }}>No usage data</div>
              )}
            </div>
          </Card>

          <FirewallSection
            groups={firewall}
            canToggle={canToggle}
            toggle={(id, enabled) => toggleFirewallGroup.mutate({ id, enabled })}
          />

        </div>

        {/* recent calls + sessions row */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, alignItems: 'start' }}>

          {/* recent tool calls */}
          <Card>
            <CardHeader title="Recent Tool Calls" />
            {recentCalls.length === 0 ? (
              <EmptyRow label="No tool calls yet" />
            ) : (
              recentCalls.slice(0, 10).map((c, i) => (
                <div key={c.id} style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '10px 20px',
                  borderBottom: i < Math.min(recentCalls.length, 10) - 1 ? '1px solid var(--confire-border)' : undefined,
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
                    <span style={{
                      width: 8, height: 8, borderRadius: '50%', flexShrink: 0,
                      background: INTEGRATION_COLOR[c.integration] ?? INTEGRATION_COLOR.default,
                    }} />
                    <div style={{ minWidth: 0 }}>
                      <div style={{ fontSize: 13, color: 'var(--confire-text)', fontWeight: 500, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {c.tool_type}
                      </div>
                      <div style={{ fontSize: 11, color: 'var(--confire-text-muted)' }}>
                        {c.integration} · {c.mode}
                      </div>
                    </div>
                  </div>
                  <div style={{ textAlign: 'right', flexShrink: 0, marginLeft: 12 }}>
                    <div style={{ fontSize: 12, color: c.reduction_ratio > 0 ? 'var(--confire-green, #2dd9a0)' : 'var(--confire-text-muted)' }}>
                      {c.reduction_ratio > 0 ? `−${c.reduction_ratio}%` : '—'}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--confire-text-muted)' }}>{timeAgo(c.created_at)}</div>
                  </div>
                </div>
              ))
            )}
          </Card>

          {/* Devices */}
          <DevicesCard apiKeys={apiKeys} revokeKey={revokeKey} />
        </div>

        {/* quick links */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
          {[
            { label: 'Upgrade plan', href: '/pricing', desc: 'Get more optimizations and features' },
            { label: 'Documentation', href: '/docs', desc: 'Learn how to use Confire' },
            { label: 'Billing', href: '/billing', desc: 'Manage your subscription and credits' },
          ].map(link => (
            <a key={link.href} href={link.href} style={{ textDecoration: 'none' }}>
              <div style={{
                background: 'var(--confire-surface)',
                border: '1px solid var(--confire-border)',
                borderRadius: 8,
                padding: '16px 20px',
                transition: 'border-color 0.15s',
              }}
                onMouseEnter={e => (e.currentTarget.style.borderColor = accent)}
                onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--confire-border)')}
              >
                <div style={{ fontSize: 14, fontWeight: 600, color: accent }}>{link.label} →</div>
                <div style={{ fontSize: 12, color: 'var(--confire-text-muted)', marginTop: 4 }}>{link.desc}</div>
              </div>
            </a>
          ))}
        </div>

      </div>
    </div>
  )
}

export function DashboardApp() {
  return (
    <QueryClientProvider client={queryClient}>
      <Dashboard />
    </QueryClientProvider>
  )
}
