import * as React from 'react'
import { cn } from '@/lib/utils'

/** Visible corner mark size — keep offset at half so squares sit on the border. */
const CORNER_PX = 7
const CORNER_OFFSET = CORNER_PX / 2

function CornerSquare({ className, style }: { className?: string; style?: React.CSSProperties }) {
  return (
    <span
      className={cn('pointer-events-none absolute z-10 block bg-confire-square', className)}
      style={{ width: CORNER_PX, height: CORNER_PX, ...style }}
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
    <CornerSquare key="tl" style={{ top: -CORNER_OFFSET, left: -CORNER_OFFSET }} />,
    <CornerSquare key="tr" style={{ top: -CORNER_OFFSET, right: -CORNER_OFFSET }} />,
    <CornerSquare key="bl" style={{ bottom: -CORNER_OFFSET, left: -CORNER_OFFSET }} />,
    <CornerSquare key="br" style={{ bottom: -CORNER_OFFSET, right: -CORNER_OFFSET }} />,
  ]

  for (let c = 1; c < cols; c++) {
    const left = `calc(${(c / cols) * 100}% - ${CORNER_OFFSET}px)`
    marks.push(<CornerSquare key={`tc${c}`} className="hidden sm:block" style={{ top: -CORNER_OFFSET, left }} />)
    marks.push(<CornerSquare key={`bc${c}`} className="hidden sm:block" style={{ bottom: -CORNER_OFFSET, left }} />)
  }
  for (let r = 1; r < rows; r++) {
    const top = `calc(${(r / rows) * 100}% - ${CORNER_OFFSET}px)`
    marks.push(<CornerSquare key={`lr${r}`} className="hidden sm:block" style={{ top, left: -CORNER_OFFSET }} />)
    marks.push(<CornerSquare key={`rr${r}`} className="hidden sm:block" style={{ top, right: -CORNER_OFFSET }} />)
  }

  return (
    <div className={cn('relative', className)}>
      {marks}
      {children}
    </div>
  )
}
