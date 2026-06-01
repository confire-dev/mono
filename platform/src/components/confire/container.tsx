import * as React from 'react'
import { cn } from '@/lib/utils'

export function Container({
  className,
  children,
  ...props
}: React.ComponentProps<'div'>) {
  return (
    <div
      className={cn('mx-auto w-full max-w-[var(--confire-max-w)] px-[var(--confire-pad)] max-md:px-5', className)}
      {...props}
    >
      {children}
    </div>
  )
}

export function Section({
  className,
  children,
  size = 'default',
  ...props
}: React.ComponentProps<'section'> & { size?: 'default' | 'sm' }) {
  return (
    <section
      className={cn(size === 'sm' ? 'py-12' : 'py-20', className)}
      {...props}
    >
      {children}
    </section>
  )
}

export function SectionLabel({
  className,
  children,
  number,
}: {
  className?: string
  children: React.ReactNode
  number?: string | number
}) {
  return (
    <div
      className={cn(
        'mb-12 flex items-center gap-3.5 text-[10px] font-bold uppercase tracking-[0.18em] text-confire-muted',
        className,
      )}
    >
      {number != null && <span className="text-confire-accent">{number}</span>}
      {children}
      <span className="h-px flex-1 bg-confire-border" />
    </div>
  )
}

export function Separator({ className }: { className?: string }) {
  return <div className={cn('h-px w-full bg-confire-border', className)} />
}
