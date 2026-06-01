import * as React from 'react'
import { cn } from '@/lib/utils'
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

export function ConfireLogo({ className, wordmark = 'CONFIRE' }: { className?: string; wordmark?: string }) {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <svg width="36" height="24" viewBox="0 0 44 30" fill="none" aria-hidden>
        <path
          d="M36 22H8C5.8 22 4 20.2 4 18s1.8-4 4-4h.4C8.8 10.5 12 8 16 8c2.4 0 4.5 1 6 2.6C23.1 9 25 8 27.2 8c4.9 0 8.8 3.9 8.8 8.7v.1C38.2 17.3 40 19.4 40 22c0 0-1.3 0-4 0z"
          fill="var(--confire-accent)"
        />
        <path
          d="M10 22h24c1.1 0 2 .9 2 2s-.9 2-2 2H10c-1.1 0-2-.9-2-2s.9-2 2-2z"
          fill="var(--confire-accent)"
          opacity="0.65"
        />
        <path
          d="M12 26h20c.6 0 1 .4 1 1s-.4 1-1 1H12c-.6 0-1-.4-1-1s.4-1 1-1z"
          fill="var(--confire-accent)"
          opacity="0.45"
        />
      </svg>
      <span className="text-[13px] font-extrabold tracking-[0.13em] text-confire-text">{wordmark}</span>
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
          <Button variant="ghost-sm" asChild>
            <a href={loginHref}>Login</a>
          </Button>
          <Button variant="outline-sm" asChild>
            <a href={ctaHref}>{ctaLabel}</a>
          </Button>
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
      { label: 'Pricing', href: '/pricing' },
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
