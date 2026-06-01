import * as React from 'react'
import { cn } from '@/lib/utils'

export function Input({ className, type, ...props }: React.ComponentProps<'input'>) {
  return (
    <input
      type={type}
      className={cn(
        'h-10 w-full rounded-lg border border-confire-border-mid bg-confire-card px-3 py-2 text-sm text-confire-text outline-none transition-colors placeholder:text-confire-dim focus-visible:border-confire-accent focus-visible:ring-2 focus-visible:ring-confire-accent/30 disabled:cursor-not-allowed disabled:opacity-50',
        className,
      )}
      {...props}
    />
  )
}
