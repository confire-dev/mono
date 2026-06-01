import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center gap-1.5 rounded-md border border-dashed font-medium tracking-wide',
  {
    variants: {
      color: {
        green: 'border-confire-green/50 bg-confire-green-dim text-confire-green',
        accent: 'border-confire-accent/55 bg-confire-accent-dim text-confire-accent',
        gray: 'border-confire-ghost-border bg-confire-surface-subtle text-confire-dim',
      },
      size: {
        md: 'px-3.5 py-1.5 text-[13px]',
        sm: 'px-2.5 py-1 text-xs',
      },
    },
    defaultVariants: {
      color: 'green',
      size: 'md',
    },
  },
)

export function Badge({
  className,
  color,
  size,
  icon,
  children,
  ...props
}: React.ComponentProps<'span'> &
  VariantProps<typeof badgeVariants> & { icon?: React.ReactNode }) {
  return (
    <span className={cn(badgeVariants({ color, size }), className)} {...props}>
      {icon && <span className="leading-none">{icon}</span>}
      {children}
    </span>
  )
}
