import * as React from 'react'
import { cn } from '@/lib/utils'
import { WithCorners } from './corner-squares'
import { Button } from './button'

export interface PricingPlan {
  name: string
  tagline: string
  price: string | 'Custom'
  period?: string
  features?: string[]
  cta: string
  featured?: boolean
  onCta?: () => void
}

export interface PricingTab {
  label: string
  icon?: React.ReactNode
}

export function PricingSection({
  plans,
  tabs,
}: {
  plans: PricingPlan[]
  tabs?: PricingTab[]
}) {
  return (
    <div>
      {tabs && tabs.length > 0 && (
        <div className="mb-10 flex justify-center">
          <div className="inline-flex gap-0.5 rounded-full border border-confire-border bg-confire-card p-1">
            {tabs.map((tab, i) => (
              <span
                key={i}
                className={cn(
                  'flex items-center gap-1.5 rounded-full px-5 py-2 font-sans text-sm font-semibold',
                  i === 0 ? 'bg-confire-accent text-confire-white' : 'text-confire-secondary',
                )}
              >
                {tab.icon && <span className="text-[13px]">{tab.icon}</span>}
                {tab.label}
              </span>
            ))}
          </div>
        </div>
      )}

      <WithCorners cols={plans.length} rows={1}>
        <div
          className={cn(
            'grid border border-confire-border',
            plans.length === 2 && 'grid-cols-1 sm:grid-cols-2',
            plans.length === 3 && 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
            plans.length >= 4 && 'grid-cols-1 sm:grid-cols-2 xl:grid-cols-4',
          )}
        >
          {plans.map((plan, i) => (
            <PricingCard key={i} plan={plan} last={i === plans.length - 1} />
          ))}
        </div>
      </WithCorners>

    </div>
  )
}

function PricingCard({ plan, last }: { plan: PricingPlan; last?: boolean }) {
  return (
    <div
      className={cn(
        'flex flex-col gap-0 p-7',
        !last && 'border-b border-confire-border sm:border-b-0 sm:border-r',
        plan.featured && 'bg-confire-card-featured',
      )}
    >
      <div className="mb-1 text-base font-bold text-confire-text">{plan.name}</div>
      <div className="mb-5 text-xs text-confire-muted">{plan.tagline}</div>

      {plan.price === 'Custom' ? (
        <div className="mb-1 text-4xl font-extrabold tracking-tight text-confire-text">Custom</div>
      ) : (
        <div className="mb-1 flex items-baseline gap-0.5">
          <span className="text-[38px] font-extrabold tracking-tight text-confire-text">
            {plan.price}
          </span>
          {plan.period && (
            <span className="text-[13px] text-confire-muted">{plan.period}</span>
          )}
        </div>
      )}

      {plan.features && plan.features.length > 0 && (
        <div className="mt-5 flex flex-wrap gap-1.5">
          {plan.features.map((f) => (
            <span
              key={f}
              className="rounded-full border border-confire-border-soft px-2.5 py-1 text-xs text-confire-tertiary"
            >
              {f}
            </span>
          ))}
        </div>
      )}

      <div className="mt-auto pt-6">
        <Button
          variant="outline"
          className="w-full justify-center"
          onClick={plan.onCta}
        >
          {plan.cta}
        </Button>
      </div>
    </div>
  )
}
