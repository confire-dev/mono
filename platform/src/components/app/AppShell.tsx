import type { ReactNode } from 'react'
import { cn, LayerCard } from '@cloudflare/kumo'

interface Props {
  children: ReactNode
  className?: string
}

export function AppShell({ children, className }: Props) {
  return (
    <div
      className="min-h-svh"
      style={{ background: '#000', colorScheme: 'dark' }}
    >
      {/* logo top-left */}
      <header
        style={{
          position: 'fixed', top: 0, left: 0, right: 0, zIndex: 10,
          height: 56, display: 'flex', alignItems: 'center',
          padding: '0 24px',
          background: 'transparent',
        }}
      >
        <a href="/" style={{ display: 'flex', alignItems: 'center', textDecoration: 'none' }}>
          <img src="/brand/confire-bg-transparent-logo-white.svg" alt="Confire" height={28} style={{ height: 28, width: 'auto' }} />
        </a>
      </header>

      {/* centered content */}
      <div
        style={{
          display: 'flex', minHeight: '100svh',
          alignItems: 'center', justifyContent: 'center',
          padding: '80px 24px 48px',
        }}
      >
        <div style={{ width: '100%', maxWidth: 400 }}>
          <LayerCard className={cn('w-full rounded-2xl p-7', className)}>
            {children}
          </LayerCard>
        </div>
      </div>
    </div>
  )
}
