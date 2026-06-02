import type { ReactNode } from 'react'
import { cn, LayerCard } from '@cloudflare/kumo'

interface Props {
  children: ReactNode
  className?: string
}

export function AppShell({ children, className }: Props) {
  return (
    <div
      className="relative flex min-h-svh flex-col items-center justify-center p-6 md:p-10"
      style={{
        background: 'radial-gradient(ellipse 80% 60% at 50% -10%, rgba(244,129,31,0.07) 0%, transparent 60%)',
      }}
    >
      {/* subtle dot grid */}
      <div
        className="pointer-events-none absolute inset-0"
        style={{
          backgroundImage: 'radial-gradient(circle, rgba(255,255,255,0.04) 1px, transparent 1px)',
          backgroundSize: '24px 24px',
        }}
      />
      <div className="relative z-10 flex w-full max-w-sm flex-col items-center gap-6">
        {/* logo mark */}
        <a href="/" className="flex items-center gap-2.5 no-underline">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2C12 2 5 6 5 13C5 17.418 8.134 21 12 21C15.866 21 19 17.418 19 13C19 6 12 2 12 2Z" fill="#f4811f" opacity="0.9"/>
            <path d="M12 7C12 7 8.5 10 8.5 14C8.5 16.485 10.015 18.5 12 18.5C13.985 18.5 15.5 16.485 15.5 14C15.5 10 12 7 12 7Z" fill="#fdb97d" opacity="0.7"/>
          </svg>
          <span style={{ fontSize: 18, fontWeight: 700, color: 'var(--kumo-text-heading, #fff)', letterSpacing: '-0.01em' }}>
            Confire
          </span>
        </a>
        <LayerCard className={cn('w-full rounded-xl p-7', className)}>
          {children}
        </LayerCard>
      </div>
    </div>
  )
}
