import * as React from 'react'
import { Slot } from 'radix-ui'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex shrink-0 cursor-pointer items-center justify-center gap-1.5 rounded-full border-none font-sans font-semibold whitespace-nowrap transition-all duration-150 outline-none focus-visible:ring-2 focus-visible:ring-confire-accent/50 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        primary:
          'bg-confire-accent px-6 py-2.5 text-[15px] text-confire-white hover:brightness-110 hover:-translate-y-px',
        ghost:
          'border border-confire-border-mid bg-transparent px-[18px] py-2 text-sm text-confire-text hover:bg-confire-hover',
        outline:
          'border border-confire-border-strong bg-transparent px-6 py-2.5 text-[15px] text-confire-text hover:bg-confire-hover',
        'outline-accent':
          'border border-dashed border-confire-accent/60 bg-transparent px-6 py-2.5 text-[15px] text-confire-accent hover:bg-confire-accent-dim',
        'ghost-sm':
          'border border-transparent bg-transparent px-3.5 py-1.5 text-[13px] text-confire-dim hover:text-confire-text',
        'outline-sm':
          'border border-confire-border-mid bg-transparent px-3.5 py-1.5 text-[13px] text-confire-text hover:bg-confire-hover',
        white:
          'bg-confire-white px-6 py-2.5 text-[15px] text-confire-inverse hover:bg-confire-white-hover',
        'white-ghost':
          'border border-confire-ghost-border-strong bg-transparent px-6 py-2.5 text-[15px] text-confire-white hover:bg-confire-hover',
      },
      size: {
        default: '',
        sm: 'px-4 py-2 text-sm',
        lg: 'px-8 py-3 text-base',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'default',
    },
  },
)

export function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: React.ComponentProps<'button'> &
  VariantProps<typeof buttonVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot.Root : 'button'
  return (
    <Comp className={cn(buttonVariants({ variant, size, className }))} {...props} />
  )
}

export { buttonVariants }
