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
        <a href="/" style={{ display: 'flex', alignItems: 'center', gap: 8, textDecoration: 'none' }}>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2C12 2 5 6 5 13C5 17.418 8.134 21 12 21C15.866 21 19 17.418 19 13C19 6 12 2 12 2Z" fill="#f4811f" opacity="0.9"/>
            <path d="M12 7C12 7 8.5 10 8.5 14C8.5 16.485 10.015 18.5 12 18.5C13.985 18.5 15.5 16.485 15.5 14C15.5 10 12 7 12 7Z" fill="#fdb97d" opacity="0.7"/>
          </svg>
          <span style={{ fontSize: 16, fontWeight: 700, color: '#fff', letterSpacing: '-0.01em' }}>
            Confire
          </span>
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
