"use client"

import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { Button, Text, TimeseriesChart } from '@cloudflare/kumo'
import {
  House, Pulse, Shield, Terminal, CreditCard, BookOpen,
  GearSix, SignOut, List, X, CaretRight, Lightning,
  DeviceMobile, ChartBar,
} from '@phosphor-icons/react'
import { useAuth } from '@/hooks/use-auth'
import { useDashboard } from '@/hooks/use-dashboard'
import { formatTokenCount } from '@/lib/types'

echarts.use([LineChart, GridComponent, TooltipComponent, CanvasRenderer])

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


const INTEGRATION_COLOR: Record<string, string> = {
  'claude-code': '#f4811f',
  cursor:        '#7b68ee',
  vscode:        '#007acc',
  default:       '#6b7280',
}

// ── primitives ────────────────────────────────────────────────────────────────

function Card({ children, className, style }: { children: React.ReactNode; className?: string; style?: React.CSSProperties }) {
  return (
    <div
      className={className}
      style={{
        background: 'var(--confire-bg-card)',
        border: '1px solid var(--confire-border)',
        borderRadius: 8,
        overflow: 'hidden',
        ...style,
      }}
    >
      {children}
    </div>
  )
}

function CardHeader({ title, action }: { title: string; action?: React.ReactNode }) {
  return (
    <div style={{
      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      padding: '14px 20px',
      borderBottom: '1px solid var(--confire-border)',
    }}>
      <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--confire-text)', letterSpacing: '0.01em' }}>{title}</span>
      {action}
    </div>
  )
}

function EmptyRow({ label }: { label: string }) {
  return (
    <div style={{ padding: '28px 20px', textAlign: 'center', color: 'var(--confire-text-muted)', fontSize: 13 }}>
      {label}
    </div>
  )
}

function StatCard({ label, value, sub, accent }: { label: string; value: string; sub?: string; accent?: boolean }) {
  return (
    <div style={{
      background: 'var(--confire-bg-card)',
      border: `1px solid ${accent ? '#f4811f44' : 'var(--confire-border)'}`,
      borderRadius: 8,
      padding: '20px 20px 16px',
    }}>
      <div style={{ fontSize: 12, color: 'var(--confire-text-muted)', marginBottom: 10, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</div>
      <div style={{ fontSize: 26, fontWeight: 700, color: 'var(--confire-text)', lineHeight: 1 }}>{value}</div>
      {sub && <div style={{ fontSize: 12, color: 'var(--confire-text-dim)', marginTop: 6 }}>{sub}</div>}
    </div>
  )
}

function PlanBadge({ plan }: { plan: string }) {
  const isFree = plan === 'free'
  const isEnterprise = plan.startsWith('enterprise')
  const bg  = isFree ? 'var(--confire-border)' : isEnterprise ? 'rgba(167,139,250,0.15)' : 'rgba(244,129,31,0.15)'
  const fg  = isFree ? 'var(--confire-text-muted)'   : isEnterprise ? '#a78bfa'                : '#f4811f'
  const bdr = isFree ? 'var(--confire-border)'   : isEnterprise ? 'rgba(167,139,250,0.4)'  : 'rgba(244,129,31,0.4)'
  return (
    <span style={{
      background: bg, color: fg, border: `1px solid ${bdr}`,
      borderRadius: 4, fontSize: 10, fontWeight: 700, padding: '2px 8px',
      textTransform: 'uppercase', letterSpacing: '0.07em',
    }}>{plan.replace(/_/g, ' ')}</span>
  )
}

function Toggle({ on, disabled, onClick }: { on: boolean; disabled?: boolean; onClick: () => void }) {
  return (
    <button
      disabled={disabled}
      onClick={onClick}
      aria-checked={on}
      role="switch"
      style={{
        width: 36, height: 20, borderRadius: 10, border: 'none', flexShrink: 0,
        background: on ? '#f4811f' : 'var(--confire-border)',
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.5 : 1,
        position: 'relative', transition: 'background 0.15s',
      }}
    >
      <span style={{
        position: 'absolute', top: 2, left: on ? 18 : 2,
        width: 16, height: 16, borderRadius: '50%',
        background: '#fff', transition: 'left 0.15s',
      }} />
    </button>
  )
}

// ── nav ───────────────────────────────────────────────────────────────────────

const NAV_ITEMS = [
  { label: 'Overview',    href: '/dashboard',          icon: House },
  { label: 'Activity',    href: '/dashboard/activity', icon: Pulse },
  { label: 'Firewall',    href: '/dashboard/firewall', icon: Shield },
  { label: 'Devices',     href: '/dashboard/devices',  icon: DeviceMobile },
  { label: 'Analytics',   href: '/dashboard/analytics',icon: ChartBar },
  { label: 'CLI & Setup', href: '/cli',                icon: Terminal },
]

const NAV_BOTTOM = [
  { label: 'Billing',       href: '/billing',  icon: CreditCard },
  { label: 'Documentation', href: '/docs',     icon: BookOpen },
  { label: 'Settings',      href: '/settings', icon: GearSix },
]

function NavItem({ label, href, icon: Icon, active }: { label: string; href: string; icon: React.ElementType; active: boolean }) {
  return (
    <a
      href={href}
      style={{
        display: 'flex', alignItems: 'center', gap: 10,
        padding: '8px 12px', borderRadius: 6, textDecoration: 'none',
        fontSize: 13, fontWeight: active ? 600 : 400,
        color: active ? '#f4811f' : 'var(--confire-text-dim)',
        background: active ? 'rgba(244,129,31,0.1)' : 'transparent',
        transition: 'background 0.12s, color 0.12s',
      }}
      onMouseEnter={e => { if (!active) { e.currentTarget.style.background = 'var(--confire-border)'; e.currentTarget.style.color = 'var(--confire-text)' } }}
      onMouseLeave={e => { if (!active) { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--confire-text-dim)' } }}
    >
      <Icon size={16} weight={active ? 'fill' : 'regular'} />
      {label}
    </a>
  )
}

function Sidebar({ currentPath, email, plan, onSignOut, mobile, onClose }: {
  currentPath: string
  email: string
  plan?: string
  onSignOut: () => void
  mobile?: boolean
  onClose?: () => void
}) {
  return (
    <div style={{
      width: 220, flexShrink: 0, height: '100svh', position: mobile ? 'fixed' : 'sticky',
      top: 0, left: 0, zIndex: mobile ? 50 : undefined,
      background: 'var(--confire-bg-card)',
      borderRight: '1px solid var(--confire-border)',
      display: 'flex', flexDirection: 'column',
      overflowY: 'auto',
    }}>
      {/* logo */}
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '0 16px', height: 56,
        borderBottom: '1px solid var(--confire-border)',
        flexShrink: 0,
      }}>
        <a href="/dashboard" style={{ display: 'flex', alignItems: 'center', gap: 8, textDecoration: 'none' }}>
          <Lightning size={20} weight="fill" color="#f4811f" />
          <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--confire-text)' }}>Confire</span>
        </a>
        {mobile && onClose && (
          <button onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--confire-text-dim)', lineHeight: 1 }}>
            <X size={18} />
          </button>
        )}
      </div>

      {/* account chip */}
      <div style={{ padding: '12px 16px', borderBottom: '1px solid var(--confire-border)', flexShrink: 0 }}>
        <div style={{ fontSize: 12, color: 'var(--confire-text-muted)', marginBottom: 4, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{email}</div>
        {plan && <PlanBadge plan={plan} />}
      </div>

      {/* main nav */}
      <nav style={{ padding: '12px 10px', flex: 1 }}>
        <div style={{ fontSize: 10, fontWeight: 600, color: 'var(--confire-text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', padding: '4px 10px 8px' }}>
          Main
        </div>
        {NAV_ITEMS.map(item => (
          <NavItem key={item.href} {...item} active={currentPath === item.href || (item.href !== '/dashboard' && currentPath.startsWith(item.href))} />
        ))}
      </nav>

      {/* bottom nav */}
      <div style={{ padding: '10px 10px 14px', borderTop: '1px solid var(--confire-border)', flexShrink: 0 }}>
        {NAV_BOTTOM.map(item => (
          <NavItem key={item.href} {...item} active={currentPath === item.href} />
        ))}
        <button
          onClick={onSignOut}
          style={{
            display: 'flex', alignItems: 'center', gap: 10,
            padding: '8px 12px', borderRadius: 6, width: '100%',
            fontSize: 13, color: 'var(--confire-text-dim)',
            background: 'none', border: 'none', cursor: 'pointer',
            transition: 'background 0.12s, color 0.12s', textAlign: 'left',
          }}
          onMouseEnter={e => { e.currentTarget.style.background = 'var(--confire-border)'; e.currentTarget.style.color = '#f87171' }}
          onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--confire-text-dim)' }}
        >
          <SignOut size={16} />
          Sign out
        </button>
      </div>
    </div>
  )
}

// ── usage chart ───────────────────────────────────────────────────────────────

function UsageChart({ calls }: { calls: { raw_bytes: number; optimized_bytes: number; created_at: string }[] }) {
  const byDay = new Map<string, number>()
  for (const c of calls) {
    const day = c.created_at.slice(0, 10)
    byDay.set(day, (byDay.get(day) ?? 0) + Math.max(0, Math.round((c.raw_bytes - c.optimized_bytes) / 4)))
  }
  const data: [number, number][] = Array.from(byDay.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([day, tokens]) => [new Date(day).getTime(), tokens])

  if (data.length === 0) return null

  return (
    <Card>
      <CardHeader title="Tokens saved — daily" />
      <div style={{ padding: '12px 8px 4px' }}>
        <TimeseriesChart
          echarts={echarts}
          type="line"
          data={[{ name: 'Tokens saved', data, color: '#f4811f' }]}
          tooltipValueFormat={v => `${Math.round(v).toLocaleString()} tokens`}
          height={180}
        />
      </div>
    </Card>
  )
}

// ── firewall card ─────────────────────────────────────────────────────────────

function FirewallCard({
  groups, canToggle, toggle,
}: {
  groups: { id: string; label: string; enabled: boolean }[]
  canToggle: boolean
  toggle: (id: string, enabled: boolean) => void
}) {
  return (
    <Card>
      <CardHeader title="MCP Firewall" action={
        !canToggle ? (
          <a href="/pricing" style={{ fontSize: 11, color: '#f4811f', textDecoration: 'none' }}>Upgrade to enable →</a>
        ) : undefined
      } />
      {groups.length === 0 && <EmptyRow label="No firewall groups configured" />}
      {groups.map((g, i) => (
        <div key={g.id} style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          padding: '11px 20px',
          borderBottom: i < groups.length - 1 ? '1px solid var(--confire-border)' : undefined,
        }}>
          <div>
            <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--confire-text)' }}>{g.label}</div>
            <div style={{ fontSize: 11, color: 'var(--confire-text-muted)', marginTop: 2 }}>{g.id}</div>
          </div>
          <Toggle on={g.enabled} disabled={!canToggle} onClick={() => toggle(g.id, !g.enabled)} />
        </div>
      ))}
    </Card>
  )
}

// ── devices card ──────────────────────────────────────────────────────────────

function DevicesCard({
  apiKeys, revokeKey,
}: {
  apiKeys: { id: string; key_prefix: string; key_suffix?: string; device_id?: string; last_used_at?: string; created_at: string }[]
  revokeKey: { mutate: (id: string) => void; isPending: boolean; variables?: string }
}) {
  return (
    <Card>
      <CardHeader title="Connected devices" />
      {apiKeys.length === 0
        ? <EmptyRow label="No devices — install the CLI to get started" />
        : apiKeys.map((k, i) => {
            const name    = k.device_id || k.key_prefix
            const keyHint = k.key_suffix ? `••••${k.key_suffix}` : k.key_prefix.slice(0, 8) + '••••'
            return (
              <div key={k.id} style={{
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                padding: '12px 20px',
                borderBottom: i < apiKeys.length - 1 ? '1px solid var(--confire-border)' : undefined,
              }}>
                <div>
                  <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--confire-text)' }}>{name}</div>
                  <div style={{ fontSize: 11, color: 'var(--confire-text-muted)', marginTop: 2, fontFamily: 'monospace' }}>
                    {keyHint} · {k.last_used_at ? `active ${timeAgo(k.last_used_at)}` : `created ${timeAgo(k.created_at)}`}
                  </div>
                </div>
                <button
                  disabled={revokeKey.isPending && revokeKey.variables === k.id}
                  onClick={() => revokeKey.mutate(k.id)}
                  style={{
                    fontSize: 12, color: '#f87171',
                    background: 'transparent', border: '1px solid rgba(248,113,113,0.3)',
                    borderRadius: 4, padding: '3px 10px', cursor: 'pointer',
                    opacity: revokeKey.isPending && revokeKey.variables === k.id ? 0.5 : 1,
                  }}
                >
                  {revokeKey.isPending && revokeKey.variables === k.id ? '…' : 'Revoke'}
                </button>
              </div>
            )
          })
      }
    </Card>
  )
}

// ── recent calls card ─────────────────────────────────────────────────────────

function RecentCallsCard({ calls }: { calls: { id: string; tool_type: string; integration: string; mode: string; reduction_ratio: number; created_at: string }[] }) {
  return (
    <Card>
      <CardHeader title="Recent tool calls" />
      {calls.length === 0
        ? <EmptyRow label="No tool calls yet" />
        : calls.slice(0, 10).map((c, i) => (
            <div key={c.id} style={{
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              padding: '10px 20px',
              borderBottom: i < Math.min(calls.length, 10) - 1 ? '1px solid var(--confire-border)' : undefined,
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
                <span style={{
                  width: 7, height: 7, borderRadius: '50%', flexShrink: 0,
                  background: INTEGRATION_COLOR[c.integration] ?? INTEGRATION_COLOR.default,
                }} />
                <div style={{ minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--confire-text)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {c.tool_type}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--confire-text-muted)' }}>{c.integration} · {c.mode}</div>
                </div>
              </div>
              <div style={{ textAlign: 'right', flexShrink: 0, marginLeft: 12 }}>
                <div style={{ fontSize: 12, color: c.reduction_ratio > 0 ? '#2dd9a0' : 'var(--confire-text-muted)' }}>
                  {c.reduction_ratio > 0 ? `−${c.reduction_ratio}%` : '—'}
                </div>
                <div style={{ fontSize: 11, color: 'var(--confire-text-muted)' }}>{timeAgo(c.created_at)}</div>
              </div>
            </div>
          ))
      }
    </Card>
  )
}

// ── quick links ───────────────────────────────────────────────────────────────

const QUICK_LINKS = [
  { label: 'Upgrade plan',    href: '/pricing',  desc: 'More optimizations & features', icon: Lightning },
  { label: 'Documentation',   href: '/docs',     desc: 'Guides and API reference',      icon: BookOpen },
  { label: 'Billing',         href: '/billing',  desc: 'Manage subscription & credits', icon: CreditCard },
]

function QuickLinks() {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
      {QUICK_LINKS.map(({ label, href, desc, icon: Icon }) => (
        <a key={href} href={href} style={{ textDecoration: 'none' }}>
          <div
            style={{
              background: 'var(--confire-bg-card)', border: '1px solid var(--confire-border)',
              borderRadius: 8, padding: '16px 18px',
              transition: 'border-color 0.15s',
            }}
            onMouseEnter={e => (e.currentTarget.style.borderColor = '#f4811f')}
            onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--confire-border)')}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
              <Icon size={15} color="#f4811f" />
              <span style={{ fontSize: 13, fontWeight: 600, color: '#f4811f' }}>{label}</span>
              <CaretRight size={11} color="#f4811f" style={{ marginLeft: 'auto' }} />
            </div>
            <div style={{ fontSize: 12, color: 'var(--confire-text-muted)' }}>{desc}</div>
          </div>
        </a>
      ))}
    </div>
  )
}

// ── main page ─────────────────────────────────────────────────────────────────

function Dashboard() {
  const { user, session, loading: authLoading, signOut } = useAuth()
  const apiKey    = session?.access_token ?? null
  const workerBase = (import.meta as any).env?.PUBLIC_WORKER_URL ?? ''
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const currentPath = typeof window !== 'undefined' ? window.location.pathname : '/dashboard'

  const {
    me, recentCalls, allCallBytes, apiKeys, firewall,
    loading, error,
    totalSavedTokens, totalCallsAllTime,
    revokeKey, toggleFirewallGroup,
  } = useDashboard(apiKey, workerBase)

  if (authLoading || loading) {
    return (
      <div style={{ display: 'flex', minHeight: '100svh', alignItems: 'center', justifyContent: 'center', background: 'var(--confire-bg)' }}>
        <Text variant="secondary" size="sm">Loading…</Text>
      </div>
    )
  }

  if (!user) {
    window.location.href = '/login'
    return null
  }

  const usedPct  = me ? Math.min(100, Math.round((me.used / me.effectiveLimit) * 100)) : 0
  const canToggle = !!(me?.features?.firewallGroupToggles)

  return (
    <div style={{ display: 'flex', minHeight: '100svh', background: 'var(--confire-bg)', color: 'var(--confire-text)' }}>

      {/* mobile overlay */}
      {sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          style={{ position: 'fixed', inset: 0, zIndex: 40, background: 'rgba(0,0,0,0.5)' }}
        />
      )}

      {/* sidebar — desktop always-visible, mobile slide-in */}
      <div className="sidebar-desktop" style={{ display: 'flex' }}>
        <Sidebar
          currentPath={currentPath}
          email={user.email ?? ''}
          plan={me?.plan}
          onSignOut={signOut}
        />
      </div>

      {sidebarOpen && (
        <Sidebar
          currentPath={currentPath}
          email={user.email ?? ''}
          plan={me?.plan}
          onSignOut={() => { signOut(); setSidebarOpen(false) }}
          mobile
          onClose={() => setSidebarOpen(false)}
        />
      )}

      {/* main content */}
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>

        {/* top bar */}
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          padding: '0 24px', height: 56, flexShrink: 0,
          borderBottom: '1px solid var(--confire-border)',
          background: 'var(--confire-bg-card)',
        }}>
          {/* hamburger — mobile only */}
          <button
            className="hamburger"
            onClick={() => setSidebarOpen(true)}
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--confire-text-dim)', lineHeight: 1, marginRight: 8 }}
          >
            <List size={20} />
          </button>

          {/* breadcrumb */}
          <div style={{ flex: 1, display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--confire-text-muted)' }}>
            <Lightning size={14} color="#f4811f" />
            <span style={{ color: 'var(--confire-text-dim)' }}>Confire</span>
            <CaretRight size={11} />
            <span style={{ color: 'var(--confire-text)', fontWeight: 500 }}>Overview</span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            {me && <PlanBadge plan={me.plan} />}
            <Button variant="outline" size="sm" onClick={signOut}>Sign out</Button>
          </div>
        </div>

        {/* page body */}
        <div style={{ flex: 1, padding: '28px 28px 40px', overflowY: 'auto' }}>
          <div style={{ maxWidth: 1100, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: 24 }}>

            {/* page title */}
            <div>
              <div style={{ fontSize: 20, fontWeight: 700, marginBottom: 4 }}>Overview</div>
              <div style={{ fontSize: 13, color: 'var(--confire-text-dim)' }}>
                {me ? `${me.planName} plan · ${me.subscriptionStatus}` : 'Your Confire account'}
              </div>
            </div>

            {error && (
              <div style={{ background: '#3b1c1c', border: '1px solid #6b2d2d', borderRadius: 8, padding: '12px 16px', fontSize: 13, color: '#f87171' }}>
                {error}
              </div>
            )}

            {/* stat cards */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }} className="stat-grid">
              <StatCard label="Tokens saved" value={formatTokenCount(totalSavedTokens)} sub="all-time" accent />
              <StatCard label="Tool calls processed" value={totalCallsAllTime.toLocaleString()} sub="by Confire" />
              <StatCard
                label="Usage this period"
                value={me ? `${me.used.toLocaleString()} / ${me.effectiveLimit.toLocaleString()}` : '—'}
                sub={me ? `${usedPct}% of limit` : undefined}
              />
            </div>

            {/* usage chart */}
            <UsageChart calls={allCallBytes} />

            {/* usage bar */}
            {me && (
              <Card>
                <CardHeader title="Usage quota" />
                <div style={{ padding: '18px 20px 16px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, marginBottom: 8 }}>
                    <span style={{ color: 'var(--confire-text-dim)' }}>Cloud optimizations</span>
                    <span style={{ fontWeight: 500 }}>{me.used.toLocaleString()} / {me.effectiveLimit.toLocaleString()}</span>
                  </div>
                  <div style={{ height: 5, background: 'var(--confire-border)', borderRadius: 3, overflow: 'hidden' }}>
                    <div style={{
                      height: '100%', width: `${usedPct}%`,
                      background: usedPct >= 80 ? '#f87171' : '#f4811f',
                      borderRadius: 3, transition: 'width 0.3s',
                    }} />
                  </div>
                  {me.purchasedCredits > 0 && (
                    <div style={{ fontSize: 12, color: 'var(--confire-text-muted)', marginTop: 8 }}>
                      +{me.purchasedCredits.toLocaleString()} purchased credits available
                    </div>
                  )}
                  {usedPct >= 80 && (
                    <div style={{ marginTop: 12, padding: '10px 12px', background: 'rgba(244,129,31,0.08)', border: '1px solid rgba(244,129,31,0.25)', borderRadius: 6, fontSize: 12, color: '#f4811f' }}>
                      You're at {usedPct}% of your limit.{' '}
                      <a href="/pricing" style={{ color: '#f4811f', fontWeight: 600 }}>Upgrade or buy credits →</a>
                    </div>
                  )}
                </div>
              </Card>
            )}

            {/* firewall + devices */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }} className="two-col">
              <FirewallCard groups={firewall} canToggle={canToggle} toggle={(id, enabled) => toggleFirewallGroup.mutate({ id, enabled })} />
              <DevicesCard apiKeys={apiKeys} revokeKey={revokeKey} />
            </div>

            {/* recent calls */}
            <RecentCallsCard calls={recentCalls} />

            {/* quick links */}
            <QuickLinks />

          </div>
        </div>
      </div>

      {/* responsive styles injected once */}
      <style>{`
        .hamburger { display: none !important; }
        @media (max-width: 768px) {
          .sidebar-desktop { display: none !important; }
          .hamburger { display: flex !important; }
          .stat-grid { grid-template-columns: 1fr 1fr !important; }
          .two-col { grid-template-columns: 1fr !important; }
        }
        @media (max-width: 480px) {
          .stat-grid { grid-template-columns: 1fr !important; }
        }
      `}</style>

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
