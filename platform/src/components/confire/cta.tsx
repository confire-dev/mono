import * as React from 'react'
import { cn } from '@/lib/utils'
import { Button } from './button'
import { H2 } from './typography'

export interface MarqueeItem {
  icon: React.ReactNode
  text: string
}

export function Marquee({
  items,
  className,
}: {
  items: MarqueeItem[]
  className?: string
}) {
  const doubled = [...items, ...items]
  return (
    <div
      className={cn(
        'overflow-hidden border-t border-confire-on-accent-border bg-confire-marquee py-3',
        className,
      )}
    >
      <div className="flex w-max animate-confire-marquee gap-0">
        {doubled.map((item, i) => (
          <div
            key={i}
            className="flex shrink-0 items-center gap-2 px-9 text-[13px] font-medium whitespace-nowrap text-confire-on-accent-dim"
          >
            <span className="opacity-60">{item.icon}</span>
            {item.text}
          </div>
        ))}
      </div>
    </div>
  )
}

export interface CTASectionProps {
  title: React.ReactNode
  subtitle?: React.ReactNode
  primaryAction?: React.ReactNode
  secondaryAction?: React.ReactNode
  marqueeItems?: MarqueeItem[]
  className?: string
}

export function CTASection({
  title,
  subtitle,
  primaryAction,
  secondaryAction,
  marqueeItems,
  className,
}: CTASectionProps) {
  const tiles = [
    { style: { top: '18%', left: '8%', ['--confire-rot' as string]: 'rotate(-14deg)' } },
    { style: { top: '14%', left: '18%', ['--confire-rot' as string]: 'rotate(10deg)', animationDelay: '0.8s' } },
    { style: { top: '55%', left: '10%', ['--confire-rot' as string]: 'rotate(-8deg)', animationDelay: '0.4s' } },
    { style: { top: '12%', right: '14%', ['--confire-rot' as string]: 'rotate(12deg)', animationDelay: '1s' } },
    { style: { top: '20%', right: '6%', ['--confire-rot' as string]: 'rotate(-6deg)', animationDelay: '0.6s' } },
    { style: { bottom: '18%', right: '10%', ['--confire-rot' as string]: 'rotate(15deg)', animationDelay: '0.2s' } },
  ]

  return (
    <div className={cn('confire-cta-surface relative overflow-hidden rounded-2xl', className)}>
      {tiles.map((t, i) => (
        <div
          key={i}
          className="absolute flex size-14 animate-confire-float items-center justify-center rounded-[10px] border border-dashed border-confire-on-accent-border-dashed text-confire-on-accent-faint"
          style={t.style as React.CSSProperties}
        >
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
        </div>
      ))}

      <div className="confire-cta-glow pointer-events-none absolute bottom-0 left-1/2 h-[260px] w-[500px] -translate-x-1/2" />

      <div className="relative z-[2] px-10 py-[72px] pb-14 text-center">
        <H2 className="mb-5 text-[clamp(1.875rem,4.5vw,3.25rem)] text-confire-white">{title}</H2>
        {subtitle && (
          <p className="mx-auto mb-9 max-w-[32rem] text-base leading-relaxed text-confire-on-accent">
            {subtitle}
          </p>
        )}
        <div className="flex flex-wrap justify-center gap-3">
          {primaryAction ?? <Button variant="white">Get started</Button>}
          {secondaryAction ?? <Button variant="white-ghost">View docs</Button>}
        </div>
      </div>

      {marqueeItems && marqueeItems.length > 0 && <Marquee items={marqueeItems} />}
    </div>
  )
}
