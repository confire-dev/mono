import { useState } from 'react'
import { Button, LayerCard, Text } from '@cloudflare/kumo'
import { CaretDown, CaretRight, BookOpen, Terminal, ArrowRight } from '@phosphor-icons/react'

interface Props {
  user: { email: string; id: string }
  device: { id: string; name?: string; cliVersion?: string }
  callbackURL: string
}

const PERMISSIONS = [
  {
    label: 'Account & Analytics',
    count: 2,
    items: ['Read account usage data', 'View token savings and statistics'],
  },
  {
    label: 'Optimization Engine',
    count: 3,
    items: ['Process tool calls for optimization', 'Apply context compression rules', 'Read optimization configuration'],
  },
  {
    label: 'Billing & Subscription',
    count: 1,
    items: ['Read plan status and quota limits'],
  },
]

function PermissionRow({ label, count, items }: { label: string; count: number; items: string[] }) {
  const [expanded, setExpanded] = useState(false)
  return (
    <div style={{ borderBottom: '1px solid rgba(255,255,255,0.07)' }}>
      <button
        onClick={() => setExpanded(v => !v)}
        style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          width: '100%', padding: '11px 0', background: 'none', border: 'none', cursor: 'pointer',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span style={{ fontSize: 13, color: '#e5e7eb', fontWeight: 500 }}>{label}</span>
          <span style={{
            fontSize: 11, fontWeight: 600, color: '#6b7280',
            background: 'rgba(255,255,255,0.06)', borderRadius: 4,
            padding: '1px 6px',
          }}>{count}</span>
        </div>
        {expanded
          ? <CaretDown size={13} color="#6b7280" />
          : <CaretRight size={13} color="#6b7280" />
        }
      </button>
      {expanded && (
        <div style={{ paddingBottom: 10, display: 'flex', flexDirection: 'column', gap: 6 }}>
          {items.map(item => (
            <div key={item} style={{ display: 'flex', alignItems: 'center', gap: 8, paddingLeft: 4 }}>
              <div style={{ width: 4, height: 4, borderRadius: '50%', background: '#4b5563', flexShrink: 0 }} />
              <span style={{ fontSize: 12, color: '#9ca3af' }}>{item}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export function CliAuthorizeCard({ user, device, callbackURL }: Props) {
  const [state, setState] = useState<'idle' | 'loading' | 'done' | 'error'>('idle')
  const [error, setError] = useState('')

  async function authorize() {
    setState('loading')
    setError('')
    try {
      const res = await fetch('/api/cli/authorize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ deviceId: device.id, deviceName: device.name, callbackURL }),
      })
      if (!res.ok) {
        const { error: e } = await res.json() as { error: string }
        throw new Error(e ?? 'Authorization failed')
      }
      const { apiKey, email, message } = await res.json() as {
        apiKey: string
        email: string
        message: string
      }
      setState('done')
      // Navigate browser to CLI callback server — fetch() is blocked as mixed content
      // (HTTPS page → HTTP localhost), but navigation is not subject to that restriction.
      const params = new URLSearchParams({ api_key: apiKey, email, message })
      window.location.href = `${callbackURL}?${params}`
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Authorization failed')
      setState('error')
    }
  }

  const initials = user.email.slice(0, 2).toUpperCase()
  const totalPermissions = PERMISSIONS.reduce((sum, p) => sum + p.count, 0)

  // Success state
  if (state === 'done') {
    return (
      <div style={{
        minHeight: '100svh', display: 'flex', alignItems: 'center', justifyContent: 'center',
        background: '#000', padding: '24px', colorScheme: 'dark',
      }}>
        <div style={{ width: '100%', maxWidth: 480 }}>
          <LayerCard className="rounded-2xl p-8">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
              {/* logo */}
              <div>
                <img src="/brand/confire-bg-transparent-logo-white.svg" alt="Confire" height={30} style={{ height: 30, width: 'auto' }} />
              </div>

              {/* success heading */}
              <div>
                <div style={{ fontSize: 17, fontWeight: 600, color: '#4ade80', marginBottom: 8 }}>
                  Authorization granted to Confire CLI
                </div>
                <p style={{ fontSize: 14, color: '#9ca3af', lineHeight: 1.6, margin: 0 }}>
                  Confire is now authenticated. You can continue in your terminal.
                </p>
              </div>

              {/* links */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {[
                  { icon: Terminal, label: 'Learn more about Confire CLI', href: '/docs' },
                  { icon: BookOpen, label: 'Check out Docs', href: '/docs' },
                  { icon: ArrowRight, label: 'See Detailed Usage', href: '/dashboard' },
                ].map(({ icon: Icon, label, href }) => (
                  <a key={href + label} href={href} style={{
                    display: 'flex', alignItems: 'center', gap: 8,
                    fontSize: 13, color: '#60a5fa', textDecoration: 'none',
                    padding: '4px 0',
                  }}
                    onMouseEnter={e => e.currentTarget.style.color = '#93c5fd'}
                    onMouseLeave={e => e.currentTarget.style.color = '#60a5fa'}
                  >
                    <Icon size={14} />
                    {label}
                  </a>
                ))}
              </div>

              <p style={{ fontSize: 12, color: '#4b5563', margin: 0 }}>You can close this window.</p>
            </div>
          </LayerCard>
        </div>
      </div>
    )
  }

  return (
    <div style={{
      minHeight: '100svh', display: 'flex', alignItems: 'center', justifyContent: 'center',
      background: '#000', padding: '24px', colorScheme: 'dark',
    }}>
      <div style={{ width: '100%', maxWidth: 480 }}>
        {/* visual: CLI → Confire */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 16, marginBottom: 24 }}>
          {/* terminal icon */}
          <div style={{
            width: 52, height: 52, borderRadius: '50%',
            background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.12)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <svg width="22" height="22" viewBox="0 0 22 22" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M4 7l5 4-5 4" stroke="#e5e7eb" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M11 15h6" stroke="#e5e7eb" strokeWidth="1.6" strokeLinecap="round"/>
            </svg>
          </div>

          {/* dashed arrow */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 3 }}>
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} style={{ width: 4, height: 1, background: 'rgba(255,255,255,0.25)', borderRadius: 1 }} />
            ))}
            <div style={{ width: 0, height: 0, borderTop: '4px solid transparent', borderBottom: '4px solid transparent', borderLeft: '6px solid rgba(255,255,255,0.25)', marginLeft: 1 }} />
          </div>

          {/* Confire icon */}
          <div style={{
            width: 52, height: 52, borderRadius: '50%',
            background: 'rgba(235,90,24,0.12)', border: '1px solid rgba(235,90,24,0.3)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <img src="/brand/confire-bg-transparent-logo-orange.svg" alt="Confire" width={28} height={28} style={{ width: 28, height: 28 }} />
          </div>
        </div>

        {/* title */}
        <h1 style={{
          fontSize: 20, fontWeight: 700, color: '#fff', textAlign: 'center',
          margin: '0 0 20px', letterSpacing: '-0.01em',
        }}>
          Confire CLI wants to access your account
        </h1>

        {/* main card */}
        <LayerCard className="rounded-2xl">
          <div style={{ padding: '0 20px' }}>

            {/* signed in as */}
            <div style={{
              display: 'flex', alignItems: 'center', gap: 12,
              padding: '16px 0',
              borderBottom: '1px solid rgba(255,255,255,0.07)',
            }}>
              <div style={{
                width: 36, height: 36, borderRadius: 8, flexShrink: 0,
                background: 'rgba(244,129,31,0.15)', border: '1px solid rgba(244,129,31,0.25)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: 13, fontWeight: 700, color: '#f4811f',
              }}>
                {initials}
              </div>
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 2 }}>Signed in as</div>
                <div style={{ fontSize: 13, fontWeight: 500, color: '#e5e7eb', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{user.email}</div>
              </div>
            </div>

            {/* account row */}
            <div style={{
              display: 'flex', alignItems: 'center', gap: 12,
              padding: '16px 0',
              borderBottom: '1px solid rgba(255,255,255,0.07)',
            }}>
              <div style={{
                width: 36, height: 36, borderRadius: 8, flexShrink: 0,
                background: 'rgba(235,90,24,0.08)', border: '1px solid rgba(235,90,24,0.15)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <img src="/brand/confire-bg-transparent-logo-orange.svg" alt="Confire" width={20} height={20} style={{ width: 20, height: 20 }} />
              </div>
              <div style={{ minWidth: 0, flex: 1 }}>
                <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 2 }}>Account</div>
                <div style={{ fontSize: 13, fontWeight: 500, color: '#e5e7eb' }}>
                  {user.email.split('@')[0]}'s Account
                </div>
              </div>
            </div>

            {/* device info */}
            {(device.name || device.cliVersion) && (
              <div style={{
                padding: '14px 0',
                borderBottom: '1px solid rgba(255,255,255,0.07)',
              }}>
                <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 8 }}>Device info</div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px 16px' }}>
                  {device.name && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                      <span style={{ fontSize: 11, color: '#6b7280' }}>Name</span>
                      <span style={{ fontSize: 12, fontWeight: 500, color: '#9ca3af', fontFamily: 'monospace' }}>{device.name}</span>
                    </div>
                  )}
                  {device.cliVersion && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                      <span style={{ fontSize: 11, color: '#6b7280' }}>Version</span>
                      <span style={{ fontSize: 12, fontWeight: 500, color: '#9ca3af', fontFamily: 'monospace' }}>{device.cliVersion}</span>
                    </div>
                  )}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                    <span style={{ fontSize: 11, color: '#6b7280' }}>ID</span>
                    <span style={{ fontSize: 12, fontWeight: 500, color: '#9ca3af', fontFamily: 'monospace' }}>{device.id.slice(0, 8)}…</span>
                  </div>
                </div>
              </div>
            )}

            {/* permissions */}
            <div style={{ padding: '14px 0' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                <span style={{ fontSize: 12, color: '#6b7280' }}>This will allow Confire CLI to</span>
                <span style={{ fontSize: 11, color: '#4b5563' }}>{totalPermissions} total permissions</span>
              </div>
              {PERMISSIONS.map(p => (
                <PermissionRow key={p.label} {...p} />
              ))}
            </div>

          </div>

          {/* actions */}
          <div style={{
            padding: '16px 20px 20px',
            display: 'flex', flexDirection: 'column', gap: 10,
            borderTop: '1px solid rgba(255,255,255,0.07)',
          }}>
            {state === 'error' && (
              <Text variant="error" size="sm" as="p" DANGEROUS_className="text-center">
                {error}
              </Text>
            )}
            <Button
              onClick={authorize}
              variant="primary"
              loading={state === 'loading'}
              className="w-full"
            >
              Authorize
            </Button>
            <Button
              onClick={() => window.history.back()}
              variant="outline"
              className="w-full"
              disabled={state === 'loading'}
            >
              Cancel
            </Button>
          </div>
        </LayerCard>

        <p style={{ textAlign: 'center', fontSize: 12, color: '#4b5563', marginTop: 16 }}>
          Only authorize access if you trust this application.
        </p>
      </div>
    </div>
  )
}
