import * as React from 'react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/hooks/use-auth'
import { Button } from './button'
import { Container } from './container'

export interface NavItem {
  label: string
  href?: string
  onClick?: () => void
}

export interface SiteNavProps {
  logo?: React.ReactNode
  items?: NavItem[]
  loginHref?: string
  ctaLabel?: string
  ctaHref?: string
  className?: string
}

type LogoBg      = 'transparent' | 'black' | 'white'
type LogoColor   = 'orange' | 'white' | 'black'
type LogoSize    = 'sm' | 'md' | 'lg'

const LOGO_SIZES: Record<LogoSize, number> = { sm: 20, md: 24, lg: 32 }

/**
 * Confire flame mark using the real brand SVG assets in /public/brand/.
 *
 * Defaults: bg=transparent, color=orange — the primary logo for dark surfaces.
 * All variants are static image references so only the src string is in the bundle.
 */
export function ConfireMark({
  bg = 'transparent',
  color = 'orange',
  size = 'md',
  className,
}: {
  bg?: LogoBg
  color?: LogoColor
  size?: LogoSize
  className?: string
}) {
  const px = LOGO_SIZES[size]
  const base = `/brand/confire-bg-${bg}-logo-${color}`
  return (
    <img
      src={`${base}.png`}
      srcSet={`${base}.png 1x, ${base}@2x.png 2x`}
      alt="Confire"
      width={px}
      height={px}
      className={cn('shrink-0', className)}
      draggable={false}
    />
  )
}

export function ConfireLogo({
  className,
  wordmark = 'CONFIRE',
  markBg = 'transparent',
  markColor = 'orange',
  markSize = 'md',
}: {
  className?: string
  wordmark?: string | false
  markBg?: LogoBg
  markColor?: LogoColor
  markSize?: LogoSize
}) {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <ConfireMark bg={markBg} color={markColor} size={markSize} />
      {wordmark !== false && (
        <span className="text-[13px] font-extrabold tracking-[0.13em] text-confire-text">
          {wordmark}
        </span>
      )}
    </div>
  )
}

export function SiteNav({
  logo,
  items = [],
  loginHref = '/login',
  ctaLabel = 'Get started',
  ctaHref = '/login',
  className,
}: SiteNavProps) {
  const { user, loading } = useAuth()

  return (
    <nav
      className={cn(
        'sticky top-0 z-50 border-b border-confire-border-subtle bg-confire-bg-nav backdrop-blur-md',
        className,
      )}
    >
      <Container className="flex h-[58px] items-center gap-2">
        <a href="/" className="mr-10 shrink-0 no-underline">
          {logo ?? <ConfireLogo />}
        </a>

        <div className="hidden flex-1 items-center gap-1 md:flex">
          {items.map((item) => (
            <NavLink key={item.label} {...item} />
          ))}
        </div>

        <div className="ml-auto flex items-center gap-2">
          {!loading && user ? (
            <Button variant="outline-sm" asChild>
              <a href="/dashboard">Dashboard</a>
            </Button>
          ) : (
            <>
              <Button variant="ghost-sm" asChild>
                <a href={loginHref}>Login</a>
              </Button>
              <Button variant="outline-sm" asChild>
                <a href={ctaHref}>{ctaLabel}</a>
              </Button>
            </>
          )}
        </div>
      </Container>
    </nav>
  )
}

function NavLink({ label, href = '#', onClick }: NavItem) {
  return (
    <a
      href={href}
      onClick={onClick}
      className="flex items-center gap-1 rounded-md px-3.5 py-1.5 text-sm font-medium text-confire-dim no-underline transition-colors hover:text-confire-text"
    >
      {label}
      <svg width="8" height="8" viewBox="0 0 8 8" fill="currentColor" aria-hidden>
        <path d="M4 6L1 2.5h6L4 6z" />
      </svg>
    </a>
  )
}

export interface FooterColumn {
  heading: string
  links: { label: string; href: string }[]
}

export interface SiteFooterProps {
  columns?: FooterColumn[]
  secondaryColumns?: FooterColumn[]
  className?: string
}

export function SiteFooter({
  columns = defaultFooterColumns,
  secondaryColumns = defaultFooterSecondary,
  className,
}: SiteFooterProps) {
  return (
    <footer className={cn('border-t border-confire-border-footer bg-confire-bg-footer', className)}>
      <Container className="py-16 pb-12">
        <div className="mb-13">
          <ConfireLogo wordmark="CONFIRE" />
        </div>

        <div className="mb-12 grid grid-cols-1 gap-8 border-b border-confire-border-footer pb-12 sm:grid-cols-2 lg:grid-cols-4">
          {columns.map((col) => (
            <FooterColumn key={col.heading} {...col} />
          ))}
        </div>

        <div className="mb-12 grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
          {secondaryColumns.map((col) => (
            <FooterColumn key={col.heading} {...col} />
          ))}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-confire-border-footer pt-6">
          <p className="m-0 text-xs text-confire-border-strong">© {new Date().getFullYear()} Confire</p>
          <div className="flex flex-wrap gap-6">
            {['Privacy', 'Terms', 'Security'].map((l) => (
              <a
                key={l}
                href="#"
                className="text-xs text-confire-border-strong no-underline transition-colors hover:text-confire-secondary"
              >
                {l}
              </a>
            ))}
          </div>
        </div>
      </Container>
    </footer>
  )
}

function FooterColumn({ heading, links }: FooterColumn) {
  return (
    <div className="mb-8">
      <p className="mb-4 text-xs font-semibold tracking-wider text-confire-caption uppercase">{heading}</p>
      <ul className="m-0 flex list-none flex-col gap-2.5 p-0">
        {links.map((link) => (
          <li key={link.label}>
            <a
              href={link.href}
              className="text-sm text-confire-secondary no-underline transition-colors hover:text-confire-hover-text"
            >
              {link.label}
            </a>
          </li>
        ))}
      </ul>
    </div>
  )
}

const defaultFooterColumns: FooterColumn[] = [
  {
    heading: 'Product',
    links: [
      { label: 'Pricing', href: '/#pricing' },
      { label: 'Dashboard', href: '/dashboard' },
      { label: 'CLI', href: '/docs/cli' },
    ],
  },
  {
    heading: 'Developers',
    links: [
      { label: 'Documentation', href: '/docs' },
      { label: 'GitHub', href: 'https://github.com/confire-dev/confire' },
      { label: 'Changelog', href: '/changelog' },
    ],
  },
  {
    heading: 'Company',
    links: [
      { label: 'About', href: '/about' },
      { label: 'Blog', href: '/blog' },
      { label: 'Contact', href: '/contact' },
    ],
  },
  {
    heading: 'Legal',
    links: [
      { label: 'Privacy', href: '/privacy' },
      { label: 'Terms', href: '/terms' },
    ],
  },
]

const defaultFooterSecondary: FooterColumn[] = [
  {
    heading: 'Support',
    links: [
      { label: 'Help center', href: '/help' },
      { label: 'Status', href: 'https://status.confire.dev' },
    ],
  },
  {
    heading: 'Community',
    links: [
      { label: 'Discord', href: '#' },
      { label: 'Twitter', href: '#' },
    ],
  },
]
