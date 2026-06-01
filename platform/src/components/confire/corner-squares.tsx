import * as React from 'react'
import { cn } from '@/lib/utils'

function CornerSquare({ className, style }: { className?: string; style?: React.CSSProperties }) {
  return (
    <span
      className={cn('pointer-events-none absolute z-10 block size-1.5 bg-confire-square', className)}
      style={style}
      aria-hidden
    />
  )
}

export interface WithCornersProps {
  children: React.ReactNode
  className?: string
  cols?: number
  rows?: number
}

/** Corner-square frame used on cards and section groups. */
export function WithCorners({ children, className, cols = 1, rows = 1 }: WithCornersProps) {
  const marks: React.ReactNode[] = [
    <CornerSquare key="tl" className="-top-[3px] -left-[3px]" />,
    <CornerSquare key="tr" className="-top-[3px] -right-[3px]" />,
    <CornerSquare key="bl" className="-bottom-[3px] -left-[3px]" />,
    <CornerSquare key="br" className="-bottom-[3px] -right-[3px]" />,
  ]

  for (let c = 1; c < cols; c++) {
    const left = `calc(${(c / cols) * 100}% - 3px)`
    marks.push(<CornerSquare key={`tc${c}`} style={{ top: -3, left }} />)
    marks.push(<CornerSquare key={`bc${c}`} style={{ bottom: -3, left }} />)
  }
  for (let r = 1; r < rows; r++) {
    const top = `calc(${(r / rows) * 100}% - 3px)`
    marks.push(<CornerSquare key={`lr${r}`} style={{ top, left: -3 }} />)
    marks.push(<CornerSquare key={`rr${r}`} style={{ top, right: -3 }} />)
  }

  return (
    <div className={cn('relative', className)}>
      {marks}
      {children}
    </div>
  )
}
