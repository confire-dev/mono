"use client"

import { useEffect, useRef, useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { Button, Sidebar, Text, TimeseriesChart } from '@cloudflare/kumo'
import {
  House, Pulse, Shield, Terminal, CreditCard,
  GearSix, SignOut, Lightning, DeviceMobile, User,
  ArrowSquareOut, CheckCircle, Warning, ArrowRight,
} from '@phosphor-icons/react'
import { useAuth } from '@/hooks/use-auth'
import { useDashboard, type MeData } from '@/hooks/use-dashboard'
import { formatTokenCount } from '@/lib/types'
import { EarlyAccessForm } from '@/components/marketing/EarlyAccessForm'

echarts.use([LineChart, GridComponent, TooltipComponent, CanvasRenderer])

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

const ACTION_LABELS: Record<string, { label: string; color: string }> = {
  optimized:  { label: 'optimized',  color: '#4ade80' },
  reviewed:   { label: 'reviewed',   color: '#60a5fa' },
  blocked:    { label: 'blocked',    color: '#f87171' },
  sanitized:  { label: 'sanitized',  color: '#a78bfa' },
  redacted:   { label: 'redacted',   color: '#fbbf24' },
  passthrough:{ label: 'passed',     color: '#6b7280' },
  local:      { label: 'local',      color: '#6b7280' },
}

// ── primitives ────────────────────────────────────────────────────────────────

function Card({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <div style={{
      background: 'var(--confire-bg-card)',
      border: '1px solid var(--confire-border)',
      borderRadius: 8, overflow: 'hidden', ...style,
    }}>
      {children}
    </div>
  )
}

function CardHeader({ title, action }: { title: string; action?: React.ReactNode }) {
  return (
    <div style={{
      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      padding: '14px 20px', borderBottom: '1px solid var(--confire-border)',
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

function PlanBadge({ plan }: { plan: string }) {
  const isFree = plan === 'free'
  const isEnterprise = plan.startsWith('enterprise')
  const bg  = isFree ? 'var(--confire-border)' : isEnterprise ? 'rgba(167,139,250,0.15)' : 'rgba(244,129,31,0.15)'
  const fg  = isFree ? 'var(--confire-text-muted)' : isEnterprise ? '#a78bfa' : '#f4811f'
  const bdr = isFree ? 'var(--confire-border)' : isEnterprise ? 'rgba(167,139,250,0.4)' : 'rgba(244,129,31,0.4)'
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
    <button disabled={disabled} onClick={onClick} role="switch" aria-checked={on}
      style={{
        width: 36, height: 20, borderRadius: 10, border: 'none', flexShrink: 0,
        background: on ? '#f4811f' : 'var(--confire-border)',
        cursor: disabled ? 'not-allowed' : 'pointer', opacity: disabled ? 0.5 : 1,
        position: 'relative', transition: 'background 0.15s',
      }}
    >
      <span style={{
        position: 'absolute', top: 2, left: on ? 18 : 2,
        width: 16, height: 16, borderRadius: '50%', background: '#fff', transition: 'left 0.15s',
      }} />
    </button>
  )
}

// ── nav ───────────────────────────────────────────────────────────────────────

const NAV = [
  {
    label: null,
    items: [{ label: 'Overview', href: '/dashboard', icon: House }],
  },
  {
    label: 'Observe',
    items: [
      { label: 'Activity', href: '/dashboard/activity', icon: Pulse },
    ],
  },
  {
    label: 'Configure',
    items: [
      { label: 'Firewall', href: '/dashboard/firewall', icon: Shield },
      { label: 'Devices',  href: '/dashboard/devices',  icon: DeviceMobile },
    ],
  },
  {
    label: 'Resources',
    items: [
      { label: 'Setup', href: '/dashboard/setup', icon: Terminal },
    ],
  },
]

const NAV_BOTTOM = [
  { label: 'Billing',  href: '/dashboard/billing',  icon: CreditCard },
  { label: 'Settings', href: '/dashboard/settings', icon: GearSix },
]

function navigate(href: string) {
  history.pushState({}, '', href)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

// ── sidebar ───────────────────────────────────────────────────────────────────

function AppSidebar({ currentPath }: { currentPath: string }) {
  function click(href: string) {
    return (e: React.MouseEvent) => { e.preventDefault(); navigate(href) }
  }

  return (
    <Sidebar>
      <Sidebar.Header>
        <a href="/dashboard" onClick={click('/dashboard')}
          style={{ display: 'flex', alignItems: 'center', gap: 8, textDecoration: 'none', padding: '2px 0', overflow: 'hidden' }}>
          {/* icon: shown always; full logo: hidden when sidebar collapsed */}
          <img
            src="/brand/confire-bg-transparent-logo-orange.svg"
            alt="Confire"
            className="group-data-[state=collapsed]/sidebar:hidden"
            style={{ height: 22, width: 'auto', flexShrink: 0 }}
          />
          <img
            src="/brand/confire-bg-transparent-logo-orange-1.png"
            alt="Confire"
            className="group-data-[state=collapsed]/sidebar:block hidden"
            style={{ height: 22, width: 22, flexShrink: 0, objectFit: 'contain' }}
          />
        </a>
      </Sidebar.Header>

      <Sidebar.Content>
        {NAV.map(({ label, items }, i) => (
          <Sidebar.Group key={i}>
            {label && <Sidebar.GroupLabel>{label}</Sidebar.GroupLabel>}
            <Sidebar.Menu>
              {items.map(({ label: l, href, icon }) => {
                const active = href === '/dashboard' ? currentPath === href : currentPath.startsWith(href)
                return (
                  <Sidebar.MenuButton key={href} icon={icon} href={href} active={active} tooltip={l} onClick={click(href) as any}>
                    {l}
                  </Sidebar.MenuButton>
                )
              })}
            </Sidebar.Menu>
          </Sidebar.Group>
        ))}

        {/* spacer — pushes bottom items to the foot of the content area */}
        <div style={{ flex: 1 }} />

        {/* bottom nav — sits above the footer line */}
        <Sidebar.Group>
          <Sidebar.Menu>
            {NAV_BOTTOM.map(({ label: l, href, icon }) => (
              <Sidebar.MenuButton key={href} icon={icon} href={href} active={currentPath === href} tooltip={l} onClick={click(href) as any}>
                {l}
              </Sidebar.MenuButton>
            ))}
            <Sidebar.MenuButton icon={ArrowSquareOut} href="https://docs.confire.dev" tooltip="Docs">
              Docs
            </Sidebar.MenuButton>
          </Sidebar.Menu>
        </Sidebar.Group>
      </Sidebar.Content>

      {/* footer line — collapse trigger only */}
      <Sidebar.Footer>
        <Sidebar.Trigger />
      </Sidebar.Footer>
    </Sidebar>
  )
}

// ── user menu ─────────────────────────────────────────────────────────────────

function UserMenu({ email, plan, onSignOut }: { email: string; plan?: string; onSignOut: () => void }) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const h = (e: MouseEvent) => { if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false) }
    document.addEventListener('mousedown', h)
    return () => document.removeEventListener('mousedown', h)
  }, [open])

  const row: React.CSSProperties = {
    display: 'flex', alignItems: 'center', gap: 8, padding: '7px 12px', borderRadius: 5,
    fontSize: 13, color: 'var(--confire-text-dim)', textDecoration: 'none',
    background: 'transparent', border: 'none', cursor: 'pointer', width: '100%', textAlign: 'left',
  }

  return (
    <div ref={ref} style={{ position: 'relative' }}>
      <button onClick={() => setOpen(v => !v)} style={{
        width: 32, height: 32, borderRadius: '50%',
        background: open ? 'rgba(244,129,31,0.25)' : 'rgba(244,129,31,0.15)',
        border: `1px solid ${open ? 'rgba(244,129,31,0.5)' : 'rgba(244,129,31,0.3)'}`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontSize: 12, fontWeight: 600, color: '#f4811f', cursor: 'pointer',
      }}>
        {email[0]?.toUpperCase() ?? 'U'}
      </button>
      {open && (
        <div style={{
          position: 'absolute', top: 'calc(100% + 6px)', right: 0,
          background: 'var(--confire-bg-card)', border: '1px solid var(--confire-border)',
          borderRadius: 8, padding: '4px', minWidth: 210, zIndex: 200,
          boxShadow: '0 8px 32px rgba(0,0,0,0.5)',
        }}>
          <div style={{ padding: '8px 12px 10px', borderBottom: '1px solid var(--confire-border)', marginBottom: 4 }}>
            <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--confire-text)', marginBottom: 2 }}>{email.split('@')[0]}</div>
            <div style={{ fontSize: 11, color: 'var(--confire-text-muted)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{email}</div>
            {plan && <div style={{ marginTop: 6 }}><PlanBadge plan={plan} /></div>}
          </div>
          <a href="/dashboard/settings" onClick={() => setOpen(false)} style={row}
            onMouseEnter={e => { e.currentTarget.style.background = 'rgba(255,255,255,0.05)'; e.currentTarget.style.color = 'var(--confire-text)' }}
            onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--confire-text-dim)' }}>
            <User size={14} /> Profile
          </a>
          <a href="/dashboard/billing" onClick={() => setOpen(false)} style={row}
            onMouseEnter={e => { e.currentTarget.style.background = 'rgba(255,255,255,0.05)'; e.currentTarget.style.color = 'var(--confire-text)' }}
            onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--confire-text-dim)' }}>
            <CreditCard size={14} /> Billing
          </a>
          <div style={{ height: 1, background: 'var(--confire-border)', margin: '4px 0' }} />
          <button onClick={() => { setOpen(false); onSignOut() }} style={{ ...row, color: '#f87171' }}
            onMouseEnter={e => { e.currentTarget.style.background = 'rgba(248,113,113,0.08)' }}
            onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}>
            <SignOut size={14} /> Log out
          </button>
        </div>
      )}
    </div>
  )
}

// ── overview page ─────────────────────────────────────────────────────────────

function StatusChip({ label, value, ok }: { label: string; value: string; ok?: boolean }) {
  return (
    <div style={{
      display: 'flex', flexDirection: 'column', gap: 4,
      padding: '10px 16px', borderRadius: 7,
      background: 'var(--confire-bg-card)', border: '1px solid var(--confire-border)',
      flex: '1 1 0', minWidth: 120,
    }}>
      <span style={{ fontSize: 10, fontWeight: 600, color: 'var(--confire-text-muted)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</span>
      <span style={{ fontSize: 12, fontWeight: 500, color: ok === false ? '#f87171' : ok === true ? '#4ade80' : '#f4811f' }}>{value}</span>
    </div>
  )
}

function BigStatCard({ label, value, sub, accent }: { label: string; value: string; sub?: string; accent?: boolean }) {
  return (
    <div style={{
      background: 'var(--confire-bg-card)',
      border: `1px solid ${accent ? '#f4811f44' : 'var(--confire-border)'}`,
      borderRadius: 8, padding: '20px 20px 16px',
    }}>
      <div style={{ fontSize: 11, color: 'var(--confire-text-muted)', marginBottom: 10, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</div>
      <div style={{ fontSize: 28, fontWeight: 700, color: accent ? '#f4811f' : 'var(--confire-text)', lineHeight: 1, fontVariantNumeric: 'tabular-nums' }}>{value}</div>
      {sub && <div style={{ fontSize: 12, color: 'var(--confire-text-dim)', marginTop: 6 }}>{sub}</div>}
    </div>
  )
}

function OverviewPage({
  me, recentCalls, allCallBytes, apiKeys, firewall, canToggle, totalSavedTokens, totalCallsAllTime,
  revokeKey, toggleFirewallGroup, onRequestEarlyAccess,
}: any) {
  // context reduction from byte data
  const totalRaw = allCallBytes.reduce((s: number, c: any) => s + c.raw_bytes, 0)
  const totalOpt = allCallBytes.reduce((s: number, c: any) => s + c.optimized_bytes, 0)
  const contextReduction = totalRaw > 0 ? Math.round((1 - totalOpt / totalRaw) * 100) : 0

  // protected = calls with a meaningful action (not just passthrough/local)
  const protectedCalls = recentCalls.filter((c: any) => c.mode && c.mode !== 'passthrough' && c.mode !== 'local')
  const secretsRedacted = recentCalls.filter((c: any) => c.mode === 'redacted').length

  // top noisy tools
  const byTool: Record<string, { total: number; count: number }> = {}
  for (const c of recentCalls) {
    const t = c.tool_type || 'unknown'
    if (!byTool[t]) byTool[t] = { total: 0, count: 0 }
    byTool[t].total += c.reduction_ratio || 0
    byTool[t].count += 1
  }
  const topTools = Object.entries(byTool)
    .map(([tool, { total, count }]) => ({ tool, avg: Math.round(total / count) }))
    .sort((a, b) => b.avg - a.avg)
    .slice(0, 4)

  // usage chart data
  const byDay = new Map<string, number>()
  for (const c of allCallBytes) {
    const day = c.created_at.slice(0, 10)
    byDay.set(day, (byDay.get(day) ?? 0) + Math.max(0, Math.round((c.raw_bytes - c.optimized_bytes) / 4)))
  }
  const chartData: [number, number][] = Array.from(byDay.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([day, tokens]) => [new Date(day).getTime(), tokens])

  const usedPct = me ? Math.min(100, Math.round((me.used / me.effectiveLimit) * 100)) : 0

  // active integrations from recent calls
  const seenIntegrations = [...new Set(recentCalls.slice(0, 20).map((c: any) => c.integration).filter(Boolean))] as string[]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      {/* page heading */}
      <div>
        <h1 style={{ fontSize: 22, fontWeight: 700, margin: 0, marginBottom: 4 }}>Overview</h1>
        <p style={{ fontSize: 13, color: 'var(--confire-text-dim)', margin: 0 }}>
          {me ? `${me.planName} plan · ${me.subscriptionStatus}` : 'Your Confire account'}
        </p>
      </div>

      {/* status row */}
      <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
        <StatusChip label="Mode" value="Balanced" ok={true} />
        {seenIntegrations.includes('claude-code') && <StatusChip label="Claude Code" value="Full firewall" ok={true} />}
        {seenIntegrations.includes('cursor') && <StatusChip label="Cursor" value="MCP gateway" ok={true} />}
        {seenIntegrations.includes('vscode') && <StatusChip label="VS Code" value="MCP gateway" ok={true} />}
        <StatusChip label="Cloud optimizer" value={me?.used > 0 ? 'Enabled' : 'Enabled'} ok={true} />
      </div>

      {/* 4 stat cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12 }} className="stat-grid">
        <BigStatCard label="Tokens saved" value={formatTokenCount(totalSavedTokens)} sub="all-time" accent />
        <BigStatCard label="Context reduction" value={contextReduction > 0 ? `${contextReduction}%` : '—'} sub="avg across calls" />
        <BigStatCard label="Protected actions" value={protectedCalls.length.toLocaleString()} sub="reviewed / blocked" />
        <BigStatCard label="Secrets redacted" value={secretsRedacted > 0 ? secretsRedacted.toString() : '0'} sub="all-time" />
      </div>

      {/* recent activity + top tools */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: 16 }} className="activity-col">

        {/* recent activity */}
        <Card>
          <CardHeader title="Recent activity" />
          {recentCalls.length === 0
            ? <EmptyRow label="No activity yet — install the CLI to get started" />
            : recentCalls.slice(0, 8).map((c: any, i: number) => {
                const action = ACTION_LABELS[c.mode] ?? ACTION_LABELS.passthrough
                return (
                  <div key={c.id} style={{
                    display: 'flex', alignItems: 'center', gap: 12,
                    padding: '9px 20px',
                    borderBottom: i < Math.min(recentCalls.length, 8) - 1 ? '1px solid var(--confire-border)' : undefined,
                  }}>
                    <span style={{
                      width: 6, height: 6, borderRadius: '50%', flexShrink: 0,
                      background: INTEGRATION_COLOR[c.integration] ?? INTEGRATION_COLOR.default,
                    }} />
                    <div style={{ minWidth: 0, flex: 1 }}>
                      <span style={{ fontSize: 12, fontWeight: 500, color: 'var(--confire-text)', fontFamily: 'monospace' }}>
                        {c.tool_type}
                      </span>
                      <span style={{ fontSize: 11, color: 'var(--confire-text-muted)', marginLeft: 8 }}>{c.integration}</span>
                    </div>
                    <span style={{
                      fontSize: 11, fontWeight: 600, color: action.color,
                      background: `${action.color}18`, borderRadius: 4, padding: '2px 6px',
                    }}>
                      {action.label}
                    </span>
                    {c.reduction_ratio > 0 && (
                      <span style={{ fontSize: 11, color: '#4ade80', fontVariantNumeric: 'tabular-nums', minWidth: 40, textAlign: 'right' }}>
                        −{c.reduction_ratio}%
                      </span>
                    )}
                    <span style={{ fontSize: 11, color: 'var(--confire-text-muted)', minWidth: 52, textAlign: 'right' }}>
                      {timeAgo(c.created_at)}
                    </span>
                  </div>
                )
              })
          }
        </Card>

        {/* top noisy tools */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Card>
            <CardHeader title="Top noisy tools" />
            {topTools.length === 0
              ? <EmptyRow label="No data yet" />
              : topTools.map((t, i) => (
                  <div key={t.tool} style={{
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    padding: '9px 20px',
                    borderBottom: i < topTools.length - 1 ? '1px solid var(--confire-border)' : undefined,
                  }}>
                    <span style={{ fontSize: 12, color: 'var(--confire-text)', fontFamily: 'monospace', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: 160 }}>
                      {t.tool}
                    </span>
                    <span style={{ fontSize: 12, fontWeight: 600, color: '#4ade80', fontVariantNumeric: 'tabular-nums', flexShrink: 0 }}>
                      {t.avg}% avg
                    </span>
                  </div>
                ))
            }
          </Card>

          {/* plan usage */}
          {me && (
            <Card>
              <CardHeader title="Plan usage" />
              <div style={{ padding: '14px 20px 16px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 8 }}>
                  <span style={{ color: 'var(--confire-text-dim)' }}>Remote optimizations</span>
                  <span style={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
                    {me.used.toLocaleString()} / {me.effectiveLimit.toLocaleString()}
                  </span>
                </div>
                <div style={{ height: 4, background: 'var(--confire-border)', borderRadius: 3, overflow: 'hidden', marginBottom: 10 }}>
                  <div style={{
                    height: '100%', width: `${usedPct}%`,
                    background: usedPct >= 80 ? '#f87171' : '#f4811f',
                    borderRadius: 3, transition: 'width 0.3s',
                  }} />
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12 }}>
                  <span style={{ color: 'var(--confire-text-muted)' }}>Custom rules</span>
                  <span style={{ color: me.plan === 'free' ? '#6b7280' : '#4ade80', fontWeight: 500 }}>
                    {me.plan === 'free' ? 'locked' : 'enabled'}
                  </span>
                </div>
                {me.plan === 'free' && (
                  <div style={{ marginTop: 12 }}>
                    <button
                      onClick={onRequestEarlyAccess}
                      style={{
                        display: 'block', width: '100%', padding: '7px 12px', textAlign: 'center',
                        background: 'rgba(244,129,31,0.1)', border: '1px solid rgba(244,129,31,0.3)',
                        borderRadius: 6, fontSize: 12, fontWeight: 600, color: '#f4811f', cursor: 'pointer',
                      }}
                    >
                      Request Dev early access →
                    </button>
                  </div>
                )}
              </div>
            </Card>
          )}
        </div>
      </div>

      {/* token savings chart */}
      {chartData.length > 0 && (
        <Card>
          <CardHeader title="Tokens saved — daily" />
          <div style={{ padding: '12px 8px 4px' }}>
            <TimeseriesChart
              echarts={echarts}
              type="line"
              data={[{ name: 'Tokens saved', data: chartData, color: '#f4811f' }]}
              tooltipValueFormat={(v: number) => `${Math.round(v).toLocaleString()} tokens`}
              height={160}
            />
          </div>
        </Card>
      )}

      {/* firewall + devices */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }} className="two-col">
        {/* firewall */}
        <Card>
          <CardHeader title="MCP Firewall" action={
            !canToggle ? <a href="/dashboard/billing" style={{ fontSize: 11, color: '#f4811f', textDecoration: 'none' }}>Upgrade to enable →</a> : undefined
          } />
          {firewall.length === 0
            ? <EmptyRow label="No firewall groups configured" />
            : firewall.map((g: any, i: number) => (
                <div key={g.id} style={{
                  display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                  padding: '11px 20px',
                  borderBottom: i < firewall.length - 1 ? '1px solid var(--confire-border)' : undefined,
                }}>
                  <div>
                    <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--confire-text)' }}>{g.label}</div>
                    <div style={{ fontSize: 11, color: 'var(--confire-text-muted)', marginTop: 2 }}>{g.id}</div>
                  </div>
                  <Toggle on={g.enabled} disabled={!canToggle} onClick={() => toggleFirewallGroup.mutate({ id: g.id, enabled: !g.enabled })} />
                </div>
              ))
          }
        </Card>

        {/* devices */}
        <Card>
          <CardHeader title="Connected devices" />
          {apiKeys.length === 0
            ? <EmptyRow label="No devices — install the CLI to get started" />
            : apiKeys.map((k: any, i: number) => {
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
                      onClick={() => revokeKey.mutate(k.id)}
                      disabled={revokeKey.isPending && revokeKey.variables === k.id}
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
      </div>
    </div>
  )
}

// ── known routes ─────────────────────────────────────────────────────────────

const KNOWN_PATHS = [
  '/dashboard',
  '/dashboard/activity',
  '/dashboard/firewall',
  '/dashboard/devices',
  '/dashboard/setup',
  '/dashboard/billing',
  '/dashboard/settings',
]

// ── dashboard 404 ─────────────────────────────────────────────────────────────

function Dashboard404({ path }: { path: string }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '60vh', gap: 16, textAlign: 'center' }}>
      <div style={{
        fontSize: 11, fontWeight: 700, color: 'var(--confire-text-muted)',
        textTransform: 'uppercase', letterSpacing: '0.1em',
        background: 'rgba(244,129,31,0.08)', border: '1px solid rgba(244,129,31,0.2)',
        borderRadius: 4, padding: '3px 10px',
      }}>
        404
      </div>
      <h1 style={{ fontSize: 22, fontWeight: 700, margin: 0, color: 'var(--confire-text)' }}>
        Page not found
      </h1>
      <p style={{ fontSize: 13, color: 'var(--confire-text-dim)', margin: 0, maxWidth: 320 }}>
        <code style={{ fontFamily: 'monospace', color: 'var(--confire-text-muted)', fontSize: 12 }}>{path}</code>
        {' '}doesn't exist. It may have been moved or removed.
      </p>
      <a
        href="/dashboard"
        onClick={(e) => { e.preventDefault(); navigate('/dashboard') }}
        style={{
          marginTop: 8, padding: '8px 20px', borderRadius: 6, fontSize: 13, fontWeight: 600,
          background: '#f4811f', color: '#fff', textDecoration: 'none', border: 'none',
          cursor: 'pointer', display: 'inline-block',
        }}
      >
        Back to Overview
      </a>
    </div>
  )
}

// ── billing page ─────────────────────────────────────────────────────────────

const DEV_FEATURES = [
  '5,000 remote optimizations/month',
  'Custom dashboard guardrails',
  'Remote policy sync to local CLI',
  'Optimization history',
  'Larger input payloads',
  'Tool-use guidance',
  'Early access to new clients',
]

function fmtDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' })
}

function BillingPage({ me, apiKey, workerBase, onRequestEarlyAccess }: {
  me: MeData | null
  apiKey: string | null
  workerBase: string
  onRequestEarlyAccess: () => void
}) {
  const [actionError,     setActionError]     = useState<string | null>(null)
  const [cancelStep,      setCancelStep]      = useState<'idle' | 'confirm' | 'done'>('idle')
  const [cancelling,      setCancelling]      = useState(false)
  const [cancelledUntil,  setCancelledUntil]  = useState<string | null>(null)

  async function confirmCancel() {
    if (!apiKey) return
    setCancelling(true)
    setActionError(null)
    try {
      const res  = await fetch(`${workerBase}/api/subscription/cancel`, {
        method:  'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${apiKey}` },
      })
      const data = await res.json() as { ok?: boolean; periodEnd?: string; error?: string; message?: string }
      if (!res.ok || !data.ok) {
        setActionError(data.message ?? data.error ?? 'Cancel failed — please try again.')
        setCancelStep('idle')
        return
      }
      setCancelledUntil(data.periodEnd ?? null)
      setCancelStep('done')
    } catch {
      setActionError('Network error — please try again.')
      setCancelStep('idle')
    } finally {
      setCancelling(false)
    }
  }

  const isFree        = !me || me.plan === 'free'
  const isAnnual      = me?.plan?.includes('annual') ?? false
  const usedPct       = me ? Math.min(100, Math.round((me.used / me.effectiveLimit) * 100)) : 0
  const isPendingCancel = me?.cancelAtPeriodEnd || cancelStep === 'done'
  const periodEndLabel  = cancelledUntil ?? me?.periodEnd ?? null
  const canCancel       = !isFree && me
    && (me.subscriptionStatus === 'active' || me.subscriptionStatus === 'trialing')
    && !isPendingCancel

  const rowStyle = (last?: boolean): React.CSSProperties => ({
    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
    padding: '10px 20px',
    borderBottom: last ? undefined : '1px solid var(--confire-border)',
    fontSize: 13,
  })

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h1 style={{ fontSize: 22, fontWeight: 700, margin: 0, marginBottom: 4 }}>Billing</h1>
        <p style={{ fontSize: 13, color: 'var(--confire-text-dim)', margin: 0 }}>
          Plan, usage, and upgrade options
        </p>
      </div>

      {actionError && (
        <div style={{
          background: '#3b1c1c', border: '1px solid #6b2d2d', borderRadius: 8,
          padding: '12px 16px', fontSize: 13, color: '#f87171',
        }}>
          {actionError}
        </div>
      )}

      {/* Current plan */}
      <Card>
        <CardHeader title="Current plan" />
        {me ? (
          <>
            <div style={rowStyle()}>
              <span style={{ color: 'var(--confire-text-dim)' }}>Plan</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span style={{ fontWeight: 600 }}>{me.planName}</span>
                {isAnnual && <span style={{ fontSize: 11, color: 'var(--confire-text-muted)' }}>annual</span>}
                <PlanBadge plan={me.plan} />
                {isPendingCancel && (
                  <span style={{
                    fontSize: 10, fontWeight: 700, padding: '2px 7px', borderRadius: 4,
                    background: 'rgba(251,191,36,0.12)', color: '#fbbf24',
                    border: '1px solid rgba(251,191,36,0.3)', letterSpacing: '0.06em', textTransform: 'uppercase',
                  }}>Cancels at period end</span>
                )}
              </div>
            </div>
            <div style={rowStyle()}>
              <span style={{ color: 'var(--confire-text-dim)' }}>Status</span>
              <span style={{
                fontWeight: 600,
                color: me.subscriptionStatus === 'active'   ? '#4ade80' :
                       me.subscriptionStatus === 'trialing' ? '#60a5fa' :
                       'var(--confire-text-muted)',
              }}>
                {me.subscriptionStatus === 'none' ? 'free tier' : me.subscriptionStatus}
              </span>
            </div>
            {periodEndLabel && (
              <div style={rowStyle(true)}>
                <span style={{ color: 'var(--confire-text-dim)' }}>
                  {isPendingCancel ? 'Access until' : isAnnual ? 'Renews' : 'Next billing date'}
                </span>
                <span style={{ fontWeight: 600, color: isPendingCancel ? '#fbbf24' : undefined }}>
                  {fmtDate(periodEndLabel)}
                </span>
              </div>
            )}
          </>
        ) : (
          <div style={{ padding: '14px 20px', fontSize: 13, color: 'var(--confire-text-muted)' }}>Loading…</div>
        )}
      </Card>

      {/* Usage */}
      {me && (
        <Card>
          <CardHeader title="Usage this period" />
          <div style={{ padding: '14px 20px 18px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 8 }}>
              <span style={{ color: 'var(--confire-text-dim)' }}>Remote optimizations</span>
              <span style={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
                {me.used.toLocaleString()} / {me.effectiveLimit.toLocaleString()}
              </span>
            </div>
            <div style={{ height: 6, background: 'var(--confire-border)', borderRadius: 3, overflow: 'hidden', marginBottom: 8 }}>
              <div style={{
                height: '100%', width: `${usedPct}%`,
                background: usedPct >= 80 ? '#f87171' : '#f4811f',
                borderRadius: 3, transition: 'width 0.3s',
              }} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: 'var(--confire-text-muted)' }}>
              <span>{usedPct}% used</span>
              {me.purchasedCredits > 0 && (
                <span>+{me.purchasedCredits.toLocaleString()} purchased credits</span>
              )}
            </div>
          </div>
        </Card>
      )}

      {/* Dev early access (free plan) */}
      {isFree && (
        <Card>
          <CardHeader title="Dev — $10/month soon, early access now" />
          <div style={{ padding: '16px 20px 20px' }}>
            <p style={{ fontSize: 13, color: 'var(--confire-text-dim)', margin: '0 0 16px', lineHeight: 1.6 }}>
              Dev is opening soon for power users who want custom guardrails, higher remote limits, and full history.
              Request access and we'll email you when it's ready.
            </p>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px 20px', marginBottom: 18 }}>
              {DEV_FEATURES.map(f => (
                <div key={f} style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, color: 'var(--confire-text-dim)' }}>
                  <CheckCircle size={13} color="#4ade80" weight="fill" />
                  {f}
                </div>
              ))}
            </div>
            <button
              onClick={onRequestEarlyAccess}
              style={{
                padding: '9px 22px', borderRadius: 6, fontSize: 13, fontWeight: 600,
                background: '#f4811f', color: '#fff', border: 'none', cursor: 'pointer',
              }}
            >
              Request early access
            </button>
          </div>
        </Card>
      )}

      {/* Subscription management (paid plans) */}
      {!isFree && me && (
        <Card>
          <CardHeader title="Subscription" />
          <div style={{ padding: '14px 20px' }}>
            {cancelStep === 'done' ? (
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10, fontSize: 13 }}>
                <Warning size={16} color="#fbbf24" weight="fill" style={{ flexShrink: 0, marginTop: 1 }} />
                <span style={{ color: 'var(--confire-text-dim)', lineHeight: 1.6 }}>
                  Your subscription has been cancelled.
                  {cancelledUntil && (
                    <> You have full access until <strong style={{ color: 'var(--confire-text)' }}>{fmtDate(cancelledUntil)}</strong>.</>
                  )}
                </span>
              </div>
            ) : isPendingCancel && periodEndLabel ? (
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10, fontSize: 13 }}>
                <Warning size={16} color="#fbbf24" weight="fill" style={{ flexShrink: 0, marginTop: 1 }} />
                <span style={{ color: 'var(--confire-text-dim)', lineHeight: 1.6 }}>
                  Your subscription is scheduled to cancel. You have full access until{' '}
                  <strong style={{ color: 'var(--confire-text)' }}>{fmtDate(periodEndLabel)}</strong>.
                </span>
              </div>
            ) : cancelStep === 'confirm' ? (
              <div>
                <p style={{ fontSize: 13, color: 'var(--confire-text-dim)', margin: '0 0 14px', lineHeight: 1.6 }}>
                  Cancel your subscription?{periodEndLabel && (
                    <> You'll keep full access until <strong style={{ color: 'var(--confire-text)' }}>{fmtDate(periodEndLabel)}</strong>. No refund is issued for the remaining period.</>
                  )}
                </p>
                <div style={{ display: 'flex', gap: 10 }}>
                  <button
                    onClick={confirmCancel}
                    disabled={cancelling}
                    style={{
                      padding: '7px 18px', borderRadius: 6, fontSize: 13, fontWeight: 600,
                      background: '#f87171', color: '#fff', border: 'none', cursor: 'pointer',
                      opacity: cancelling ? 0.7 : 1,
                    }}
                  >
                    {cancelling ? 'Cancelling…' : 'Yes, cancel subscription'}
                  </button>
                  <button
                    onClick={() => setCancelStep('idle')}
                    style={{
                      padding: '7px 18px', borderRadius: 6, fontSize: 13, fontWeight: 600,
                      background: 'transparent', color: 'var(--confire-text-dim)',
                      border: '1px solid var(--confire-border)', cursor: 'pointer',
                    }}
                  >
                    Keep subscription
                  </button>
                </div>
              </div>
            ) : (
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 12 }}>
                <div>
                  {me.subscriptionStatus === 'active' && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, color: '#4ade80', marginBottom: periodEndLabel ? 4 : 0 }}>
                      <CheckCircle size={13} weight="fill" />
                      Subscription active
                    </div>
                  )}
                  {periodEndLabel && (
                    <div style={{ fontSize: 12, color: 'var(--confire-text-muted)' }}>
                      Renews {fmtDate(periodEndLabel)}
                    </div>
                  )}
                </div>
                {canCancel && (
                  <button
                    onClick={() => setCancelStep('confirm')}
                    style={{
                      padding: '6px 14px', borderRadius: 6, fontSize: 12, fontWeight: 500,
                      background: 'transparent', color: '#f87171',
                      border: '1px solid rgba(248,113,113,0.3)', cursor: 'pointer',
                    }}
                  >
                    Cancel subscription
                  </button>
                )}
              </div>
            )}
          </div>
        </Card>
      )}
    </div>
  )
}

// ── stub pages ────────────────────────────────────────────────────────────────

function StubPage({ title, subtitle, icon: Icon }: { title: string; subtitle: string; icon: React.ElementType }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h1 style={{ fontSize: 22, fontWeight: 700, margin: 0, marginBottom: 4 }}>{title}</h1>
        <p style={{ fontSize: 13, color: 'var(--confire-text-dim)', margin: 0 }}>{subtitle}</p>
      </div>
      <Card>
        <div style={{ padding: '60px 20px', textAlign: 'center' }}>
          <Icon size={32} color="var(--confire-border-strong)" weight="duotone" style={{ marginBottom: 12 }} />
          <div style={{ fontSize: 14, fontWeight: 500, color: 'var(--confire-text-dim)', marginBottom: 6 }}>Coming soon</div>
          <div style={{ fontSize: 13, color: 'var(--confire-text-muted)' }}>This page is under construction.</div>
        </div>
      </Card>
    </div>
  )
}

// ── main dashboard ────────────────────────────────────────────────────────────

function Dashboard() {
  const { user, session, loading: authLoading, signOut } = useAuth()
  const apiKey     = session?.access_token ?? null
  const workerBase = (import.meta as any).env?.PUBLIC_WORKER_URL ?? ''
  const [currentPath, setCurrentPath] = useState(() =>
    typeof window !== 'undefined' ? window.location.pathname : '/dashboard'
  )

  useEffect(() => {
    const h = () => setCurrentPath(window.location.pathname)
    window.addEventListener('popstate', h)
    return () => window.removeEventListener('popstate', h)
  }, [])

  const {
    me, recentCalls, allCallBytes, apiKeys, firewall,
    loading, error,
    totalSavedTokens, totalCallsAllTime,
    revokeKey, toggleFirewallGroup,
  } = useDashboard(apiKey, workerBase)

  if (authLoading || loading) {
    return (
      <div style={{ display: 'flex', height: '100svh', alignItems: 'center', justifyContent: 'center', background: 'var(--confire-bg)' }}>
        <Text variant="secondary" size="sm">Loading…</Text>
      </div>
    )
  }

  if (!user) { window.location.href = '/login'; return null }

  const canToggle = !!(me?.features?.firewallGroupToggles)

  const [earlyAccessOpen, setEarlyAccessOpen] = useState(false)
  function openEarlyAccess() {
    if (!apiKey) { window.location.href = '/login'; return }
    setEarlyAccessOpen(true)
  }

  // page title for topbar
  const pageLabel: Record<string, string> = {
    '/dashboard': 'Overview',
    '/dashboard/activity': 'Activity',
    '/dashboard/firewall': 'Firewall',
    '/dashboard/devices': 'Devices',
    '/dashboard/setup': 'Setup',
    '/dashboard/billing': 'Billing',
    '/dashboard/settings': 'Settings',
  }
  const topLabel = Object.entries(pageLabel).find(([p]) => p !== '/dashboard' && currentPath.startsWith(p))?.[1]
    ?? (currentPath === '/dashboard' ? 'Overview' : 'Dashboard')

  return (
    <Sidebar.Provider
      defaultOpen
      defaultWidth={260}
      style={{ height: '100svh', background: 'var(--confire-bg)', color: 'var(--confire-text)' } as React.CSSProperties}
    >
      <AppSidebar currentPath={currentPath} />

      {/* main area */}
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>

        {/* top bar */}
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          padding: '0 20px', height: 52, flexShrink: 0,
          borderBottom: '1px solid var(--confire-border)',
          background: 'var(--confire-bg-footer)',
        }}>
          <Sidebar.Trigger className="sidebar-mobile-trigger" />
          <span style={{ flex: 1, fontSize: 13, fontWeight: 500, color: 'var(--confire-text-muted)', marginLeft: 4 }}>
            {topLabel}
          </span>
          <UserMenu email={user.email ?? ''} plan={me?.plan} onSignOut={signOut} />
        </div>

        {/* scrollable content */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '28px 28px 48px' }}>
          {error && (
            <div style={{ background: '#3b1c1c', border: '1px solid #6b2d2d', borderRadius: 8, padding: '12px 16px', fontSize: 13, color: '#f87171', marginBottom: 20 }}>
              {error}
            </div>
          )}
          <div style={{ maxWidth: 1100, margin: '0 auto' }}>
            {currentPath === '/dashboard' && (
              <OverviewPage
                me={me} recentCalls={recentCalls} allCallBytes={allCallBytes}
                apiKeys={apiKeys} firewall={firewall} canToggle={canToggle}
                totalSavedTokens={totalSavedTokens} totalCallsAllTime={totalCallsAllTime}
                revokeKey={revokeKey} toggleFirewallGroup={toggleFirewallGroup}
                onRequestEarlyAccess={openEarlyAccess}
              />
            )}
            {currentPath.startsWith('/dashboard/activity') && (
              <StubPage title="Activity" subtitle="Every tool call Confire processed, in real time" icon={Pulse} />
            )}
            {currentPath.startsWith('/dashboard/firewall') && (
              <StubPage title="Firewall" subtitle="Configure mode, built-in rules, and custom policies" icon={Shield} />
            )}
            {currentPath.startsWith('/dashboard/devices') && (
              <StubPage title="Devices" subtitle="Installed clients, versions, and connection status" icon={DeviceMobile} />
            )}
            {currentPath.startsWith('/dashboard/setup') && (
              <StubPage title="Setup" subtitle="Install, configure, and verify Confire on your machine" icon={Terminal} />
            )}
            {currentPath.startsWith('/dashboard/billing') && (
              <BillingPage me={me} apiKey={apiKey} workerBase={workerBase} onRequestEarlyAccess={openEarlyAccess} />
            )}
            {currentPath.startsWith('/dashboard/settings') && (
              <StubPage title="Settings" subtitle="Account preferences, telemetry, and API keys" icon={GearSix} />
            )}
            {!KNOWN_PATHS.some(p => p === currentPath || (p !== '/dashboard' && currentPath.startsWith(p))) && (
              <Dashboard404 path={currentPath} />
            )}
          </div>
        </div>
      </div>

      <EarlyAccessForm open={earlyAccessOpen} onOpenChange={setEarlyAccessOpen} planInterest="dev" />

      <style>{`
        .sidebar-mobile-trigger { display: none !important; }
        @media (max-width: 768px) {
          .sidebar-mobile-trigger { display: flex !important; }
          .stat-grid { grid-template-columns: 1fr 1fr !important; }
          .two-col { grid-template-columns: 1fr !important; }
          .activity-col { grid-template-columns: 1fr !important; }
        }
        @media (max-width: 480px) {
          .stat-grid { grid-template-columns: 1fr 1fr !important; }
        }
      `}</style>
    </Sidebar.Provider>
  )
}

export function DashboardApp() {
  return (
    <QueryClientProvider client={queryClient}>
      <Dashboard />
    </QueryClientProvider>
  )
}

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 30_000 } },
})
