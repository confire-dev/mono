import type { ReactNode } from 'react'
import { cn, LayerCard } from '@cloudflare/kumo'

interface Props {
  children: ReactNode
  className?: string
}

export function AppShell({ children, className }: Props) {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <LayerCard className={cn('w-full max-w-sm rounded-lg p-6', className)}>
        {children}
      </LayerCard>
    </div>
  )
}
