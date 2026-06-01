import * as React from 'react'
import { cn } from '@/lib/utils'
import { WithCorners } from './corner-squares'
import { BodySm } from './typography'

export function Card({
  children,
  className,
  accent,
  footer,
}: {
  children: React.ReactNode
  className?: string
  accent?: string
  footer?: React.ReactNode
}) {
  return (
    <WithCorners cols={1} rows={1}>
      <div
        className={cn(
          'relative overflow-hidden border border-confire-border bg-confire-card p-8',
          className,
        )}
        style={accent ? { borderTopWidth: 2, borderTopColor: accent } : undefined}
      >
        {children}
        {footer && (
          <div className="mt-7 border-t border-confire-border-subtle pt-5">{footer}</div>
        )}
      </div>
    </WithCorners>
  )
}

export function TwoCards({
  card1,
  card2,
  layout = 'horizontal',
}: {
  card1: React.ReactNode
  card2: React.ReactNode
  layout?: 'horizontal' | 'vertical'
}) {
  const horizontal = layout === 'horizontal'
  return (
    <WithCorners cols={horizontal ? 2 : 1} rows={horizontal ? 1 : 2}>
      <div
        className={cn(
          'grid border border-confire-border',
          horizontal ? 'grid-cols-1 md:grid-cols-2' : 'grid-cols-1',
        )}
      >
        <div
          className={cn(
            'p-8',
            horizontal && 'border-b border-confire-border md:border-b-0 md:border-r',
            !horizontal && 'border-b border-confire-border',
          )}
        >
          {card1}
        </div>
        <div className="p-8">{card2}</div>
      </div>
    </WithCorners>
  )
}

export function ThreeCards({ cards }: { cards: React.ReactNode[] }) {
  return (
    <WithCorners cols={3} rows={1}>
      <div className="grid grid-cols-1 border border-confire-border lg:grid-cols-3">
        {cards.map((card, i) => (
          <div
            key={i}
            className={cn(
              'p-8',
              i < cards.length - 1 && 'border-b border-confire-border lg:border-b-0 lg:border-r',
            )}
          >
            {card}
          </div>
        ))}
      </div>
    </WithCorners>
  )
}

export interface BentoItem {
  colSpan?: number
  rowSpan?: number
  content: React.ReactNode
  accent?: boolean
  dark?: boolean
}

export function BentoGrid({ items }: { items: BentoItem[] }) {
  return (
    <div className="relative grid grid-cols-1 border border-confire-border sm:grid-cols-2 lg:grid-cols-4">
      {items.map((item, i) => (
        <div
          key={i}
          className={cn(
            '-m-px min-h-[180px] overflow-hidden border border-confire-border p-8',
            item.accent && 'bg-confire-accent',
            item.dark && 'bg-confire-code',
            !item.accent && !item.dark && 'bg-confire-card',
          )}
          style={{
            gridColumn: `span ${Math.min(item.colSpan ?? 1, 4)}`,
            gridRow: `span ${item.rowSpan ?? 1}`,
          }}
        >
          {item.content}
        </div>
      ))}
    </div>
  )
}

export function FeatureCard({
  icon,
  title,
  description,
  accentColor,
}: {
  icon: React.ReactNode
  title: string
  description: string
  accentColor?: string
}) {
  return (
    <div>
      <div className={cn('mb-4', !accentColor && 'text-confire-dim')} style={accentColor ? { color: accentColor } : undefined}>
        {icon}
      </div>
      <div className="mb-2.5 text-[15px] font-semibold text-confire-text">{title}</div>
      <BodySm>{description}</BodySm>
    </div>
  )
}

export function QuoteCard({
  quote,
  author,
}: {
  quote: React.ReactNode
  author: React.ReactNode
}) {
  return (
    <WithCorners cols={1} rows={1}>
      <div className="grid items-center gap-6 border border-confire-border p-10 md:grid-cols-[1fr_auto] md:gap-15 md:p-12 lg:gap-16">
        <p className="m-0 text-[clamp(1.125rem,2.5vw,1.625rem)] font-normal leading-normal text-confire-quote">
          {quote}
        </p>
        <div className="whitespace-nowrap text-right text-[13px] text-confire-muted">
          <div className="mb-1 font-semibold text-confire-tertiary">{author}</div>
        </div>
      </div>
    </WithCorners>
  )
}
