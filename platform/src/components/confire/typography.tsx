import * as React from 'react'
import { cn } from '@/lib/utils'

export function H1({ className, ...props }: React.ComponentProps<'h1'>) {
  return (
    <h1
      className={cn(
        'm-0 font-sans text-[clamp(2.375rem,5.5vw,4.25rem)] font-extrabold leading-[1.04] tracking-[-0.025em] text-confire-text',
        className,
      )}
      {...props}
    />
  )
}

export function H2({ className, ...props }: React.ComponentProps<'h2'>) {
  return (
    <h2
      className={cn(
        'm-0 font-sans text-[clamp(1.75rem,4vw,3rem)] font-bold leading-[1.08] tracking-[-0.02em] text-confire-text',
        className,
      )}
      {...props}
    />
  )
}

export function H3({ className, ...props }: React.ComponentProps<'h3'>) {
  return (
    <h3
      className={cn(
        'm-0 font-sans text-[clamp(1rem,2vw,1.25rem)] font-semibold leading-snug tracking-[-0.01em] text-confire-text',
        className,
      )}
      {...props}
    />
  )
}

export function Body({ className, ...props }: React.ComponentProps<'p'>) {
  return (
    <p
      className={cn('m-0 text-base leading-[1.65] text-confire-dim', className)}
      {...props}
    />
  )
}

export function BodySm({ className, ...props }: React.ComponentProps<'p'>) {
  return (
    <p
      className={cn('m-0 text-sm leading-relaxed text-confire-dim', className)}
      {...props}
    />
  )
}

export interface SectionTitleProps {
  label?: React.ReactNode
  title: React.ReactNode
  subtitle?: React.ReactNode
  center?: boolean
  className?: string
}

export function SectionTitle({
  label,
  title,
  subtitle,
  center = true,
  className,
}: SectionTitleProps) {
  return (
    <div className={cn('mb-14', center && 'text-center', className)}>
      {label && <div className="mb-5">{label}</div>}
      <H2 className={subtitle ? 'mb-4' : undefined}>{title}</H2>
      {subtitle && (
        <Body className={cn('max-w-[36rem]', center && 'mx-auto')}>{subtitle}</Body>
      )}
    </div>
  )
}
