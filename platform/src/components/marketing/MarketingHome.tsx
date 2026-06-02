import React, { useState } from 'react'
import {
  BodySm,
  BentoGrid,
  Button,
  CTASection,
  Container,
  ConfireLogo,
  FeatureCard,
  H1,
  H2,
  H3,
  PricingSection,
  QuoteCard,
  Section,
  SectionLabel,
  SectionTitle,
  SiteFooter,
  SiteNav,
  ThreeCards,
  TwoCards,
  VerticalTabsCode,
  WithCorners,
} from '@/components/confire'
import {
  ArrowRightIcon,
  BracketsCurlyIcon,
  CalendarIcon,
  ChartLineUpIcon,
  CheckIcon,
  CloudIcon,
  CpuIcon,
  DatabaseIcon,
  HexagonIcon,
  LightningIcon,
  LockIcon,
  PackageIcon,
  ShieldCheckIcon,
  TerminalWindowIcon,
  TimerIcon,
} from '@phosphor-icons/react'

const iconSm = 'size-3.5 shrink-0'
const iconMd = 'size-8 shrink-0'

// ── Hero — full-bleed orange card, centered, exactly like CF ──────────────────

function Hero() {
  // Re-use CTASection's exact orange card treatment but with H1 sizing
  const tiles = [
    { style: { top: '16%', left: '7%',   rotate: '-14deg', delay: '0s'   } },
    { style: { top: '12%', left: '19%',  rotate: '10deg',  delay: '0.8s' } },
    { style: { top: '52%', left: '9%',   rotate: '-8deg',  delay: '0.4s' } },
    { style: { top: '10%', right: '13%', rotate: '12deg',  delay: '1s'   } },
    { style: { top: '18%', right: '5%',  rotate: '-6deg',  delay: '0.6s' } },
    { style: { bottom: '16%', right: '9%', rotate: '15deg', delay: '0.2s' } },
  ]

  return (
    <Section className="confire-dot-region px-4 pt-6 pb-0 sm:px-8">
      {/* orange card */}
      <div className="confire-cta-surface relative overflow-hidden rounded-2xl">

        {/* floating dashed tiles */}
        {tiles.map(({ style: { rotate, delay, ...pos } }, i) => (
          <div
            key={i}
            className="absolute flex size-14 animate-confire-float items-center justify-center rounded-[10px] border border-dashed border-confire-on-accent-border-dashed text-confire-on-accent-faint"
            style={{ ...pos, animationDelay: delay, transform: `rotate(${rotate})` } as React.CSSProperties}
          >
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
          </div>
        ))}

        {/* glow */}
        <div className="confire-cta-glow pointer-events-none absolute bottom-0 left-1/2 h-[320px] w-[600px] -translate-x-1/2" />

        {/* content */}
        <div className="relative z-10 px-6 py-20 pb-16 text-center sm:px-10 sm:py-24 sm:pb-20">
          <H1 className="mb-6 text-confire-white">
            Everything we learned from
            <br className="hidden sm:block" />
            running AI agents — yours by default.
          </H1>
          <p className="mx-auto mb-10 max-w-[36rem] text-base leading-relaxed text-confire-on-accent">
            One optimizer for every tool call your AI makes.
            <br />
            Cheaper sessions, sharper context, zero config.
          </p>
          <Button variant="white" asChild>
            <a href="/login">Start building for free</a>
          </Button>
        </div>
      </div>
    </Section>
  )
}

// ── Stats bar ─────────────────────────────────────────────────────────────────

const STATS = [
  { value: '93%',  label: 'average token reduction'  },
  { value: '5+',   label: 'source-specific optimizers' },
  { value: '∞',    label: 'local optimizations free'  },
  { value: '<1ms', label: 'local optimizer latency'   },
]

function StatsBar() {
  return (
    <Section className="confire-dot-region py-0">
      <Container>
        <WithCorners cols={4} rows={1}>
          <div className="grid grid-cols-2 border border-confire-border sm:grid-cols-4">
            {STATS.map(({ value, label }, i) => (
              <div
                key={label}
                className={`px-6 py-8 text-center${i < STATS.length - 1 ? ' border-b border-confire-border sm:border-b-0 sm:border-r' : ''}`}
              >
                <div className="mb-1.5 font-sans text-[2.25rem] font-extrabold tracking-tight text-confire-accent">
                  {value}
                </div>
                <div className="text-xs text-confire-muted">{label}</div>
              </div>
            ))}
          </div>
        </WithCorners>
      </Container>
    </Section>
  )
}

// ── Integration bar ───────────────────────────────────────────────────────────

const INTEGRATIONS = [
  { name: 'Claude Code', live: true  },
  { name: 'Cursor',      live: true  },
  { name: 'VS Code',     live: false },
  { name: 'Windsurf',    live: false },
  { name: 'Zed',         live: false },
]

function IntegrationBar() {
  return (
    <Section className="py-12">
      <Container>
        <p className="mb-8 text-center text-[11px] font-semibold uppercase tracking-[0.16em] text-confire-muted">
          Works with your AI coding tool
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-3">
          {INTEGRATIONS.map(({ name, live }) => (
            <span key={name} className="flex items-center gap-1.5 text-sm font-semibold text-confire-border-strong">
              {name}
              {!live && <span className="rounded bg-confire-border px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider text-confire-muted">soon</span>}
            </span>
          ))}
        </div>
      </Container>
    </Section>
  )
}

// ── Why Confire — before/after split card ─────────────────────────────────────

function WhyConfire() {
  return (
    <Section className="confire-dot-region">
      <Container>
        <SectionTitle
          title="Why choose Confire"
          subtitle={
            <>Everything needed to{' '}
              <span className="text-confire-accent">run lean, focused AI sessions</span>
            </>
          }
        />

        <TwoCards
          card1={
            <div>
              <div className="mb-3 text-[11px] font-bold uppercase tracking-widest text-red-400/70">
                ⚠ Status: unresolved
              </div>
              <H3 className="mb-5">
                Drowning in<br />tool output noise
              </H3>
              <div className="mb-5 space-y-2 text-xs text-confire-muted">
                {[
                  '"Model lost context again after the bash run"',
                  '"Why is our Claude bill $400 this month?"',
                  '"Agent keeps hallucinating — context too big"',
                ].map(q => (
                  <div key={q} className="rounded-md border border-confire-border bg-confire-bg px-3 py-2 font-mono">
                    {q}
                  </div>
                ))}
              </div>
              <div className="flex flex-col gap-1.5 text-xs text-red-400/80">
                {['12,400 tokens per bash call', '3–4 context compacts / session', '$0.84 per session'].map(t => (
                  <span key={t} className="flex items-center gap-2">
                    <span className="size-1.5 shrink-0 rounded-full bg-red-500/60" />{t}
                  </span>
                ))}
              </div>
            </div>
          }
          card2={
            <div className="confire-cta-surface flex h-full flex-col items-center justify-center rounded-xl p-10 text-center">
              <div className="confire-cta-glow pointer-events-none absolute bottom-0 left-1/2 h-48 w-80 -translate-x-1/2" />
              <H3 className="relative z-10 mb-6 text-[1.6rem] text-white">
                Shipping with<br />Confire
              </H3>
              <div className="relative z-10 flex items-center gap-2 rounded-full bg-white/15 px-5 py-2.5 text-sm font-semibold text-white backdrop-blur-sm">
                <CheckIcon className="size-4" weight="bold" />
                890 tokens · session cost $0.06
              </div>
            </div>
          }
        />
      </Container>
    </Section>
  )
}

// ── Three pillars ─────────────────────────────────────────────────────────────

function Pillars() {
  return (
    <Section>
      <Container>
        <SectionLabel number="01">How it works</SectionLabel>
        <SectionTitle
          title="A firewall for your context window."
          subtitle="Confire intercepts every tool response before it reaches the model. Local rules run instantly. Cloud optimizers handle the complex sources."
        />

        <ThreeCards
          cards={[
            <FeatureCard
              key="local"
              icon={<TerminalWindowIcon className={iconMd} weight="duotone" />}
              title="Run everywhere"
              description="Local optimizer hooks into Claude Code, Cursor, and VS Code. Zero latency — runs in microseconds on your machine before any token is sent."
            />,
            <FeatureCard
              key="cloud"
              icon={<CloudIcon className={iconMd} weight="duotone" />}
              title="Run with any source"
              description="GitHub PRs, Figma exports, and more — cloud optimizers that know the exact shape of each tool's output and strip only the noise."
            />,
            <FeatureCard
              key="privacy"
              icon={<LockIcon className={iconMd} weight="duotone" />}
              title="Run at zero risk"
              description="Secrets and credentials redacted before anything leaves your machine. Telemetry opt-in. Your code never trains our models. Ever."
            />,
          ]}
        />
      </Container>
    </Section>
  )
}

// ── Capabilities bento ────────────────────────────────────────────────────────

function Capabilities() {
  return (
    <Section>
      <Container>
        <SectionLabel number="02">Capabilities</SectionLabel>
        <SectionTitle
          title="One smart optimizer for every tool call."
          subtitle="Close to your model, close to your data — every optimization runs before a single token hits the context window."
        />

        <BentoGrid
          items={[
            {
              colSpan: 2,
              content: (
                <div>
                  <CpuIcon className="mb-4 size-8 text-confire-accent" weight="duotone" />
                  <H3 className="mb-2">MCP Firewall</H3>
                  <BodySm>Scans every MCP tool response for prompt injection, hidden Unicode, and leaked credentials before the model ever sees it. Configurable rule groups on Dev and Pro.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <DatabaseIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <div className="mb-2 flex items-center gap-2">
                    <H3>Session Memory Guard</H3>
                    <span className="rounded bg-confire-border px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider text-confire-muted">soon</span>
                  </div>
                  <BodySm>Watches context growth and warns before you hit the limit. Automatic pre-compact optimization coming on Pro.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <ShieldCheckIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Secret redaction</H3>
                  <BodySm>Detects and strips API keys, tokens, and credentials from every tool output. Zero configuration.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <ChartLineUpIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Usage analytics</H3>
                  <BodySm>Tokens saved per session, per tool, per integration. Local stats always — full history on Dev+.</BodySm>
                </div>
              ),
            },
            {
              accent: true,
              content: (
                <div>
                  <BracketsCurlyIcon className="mb-4 size-8 text-white/80" weight="bold" />
                  <div className="mb-2 flex items-center gap-2">
                    <H3 className="text-white">Open hook API</H3>
                    <span className="rounded bg-white/20 px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider text-white/70">soon</span>
                  </div>
                  <BodySm className="text-white/70">Write custom optimizers in TypeScript. Ship as local rules or private cloud adapters.</BodySm>
                </div>
              ),
            },
          ]}
        />
      </Container>
    </Section>
  )
}

// ── How to get started ────────────────────────────────────────────────────────

const setupCode = `# One command. Hooks into Claude Code and Cursor.
$ brew install confire/tap/confire
$ confire setup

✓ Hooked into Claude Code  (v1.8+)
✓ Hooked into Cursor       (v0.44+)
✓ Local optimizers active
  More integrations coming soon`

const optimizeCode = `// No code changes needed — Confire intercepts automatically.

Bash output      12,400 tokens → 890  tokens  (93% saved)
GitHub PR diff    8,200 tokens → 620  tokens  (92% saved)
Figma export     15,000 tokens → 1,100 tokens  (93% saved)

// Avg session cost before:  $0.84
// Avg session cost after:   $0.06`

const statsCode = `$ confire stats

  124,500  tokens saved  (this week)
    2,840  tool calls processed
    ↓ 91%  average reduction

  Top savers:
    bash        47,200 tokens
    file_read   38,100 tokens
    github_pr   21,900 tokens

  Full history → confire.dev/dashboard`

function HowItWorks() {
  return (
    <Section>
      <Container>
        <SectionLabel number="03">Get started in 60 seconds</SectionLabel>
        <SectionTitle
          title="Install once. Optimize everything."
          subtitle="Hooks into your existing workflow as a Claude Code post-tool hook. No proxy, no port forwarding, no config files to maintain."
        />
        <VerticalTabsCode
          tabs={[
            {
              title: 'Install the CLI',
              description: 'One command. Auto-hooks into Claude Code, Cursor, and VS Code.',
              code: setupCode,
            },
            {
              title: 'Optimization runs automatically',
              description: 'Every tool call is intercepted and compressed. You change nothing.',
              code: optimizeCode,
            },
            {
              title: 'Track your savings',
              description: 'See exactly where your tokens go — locally or in the dashboard.',
              code: statsCode,
            },
          ]}
        />
      </Container>
    </Section>
  )
}

// ── Social proof ──────────────────────────────────────────────────────────────

function SocialProof() {
  return (
    <Section>
      <Container>
        <QuoteCard
          quote={
            <>
              Confire cut our Claude Code bill in half overnight. The agent stopped getting confused
              by tool output noise — it just{' '}
              <span className="font-semibold text-confire-text">stays on task now</span>.
              We went from 3–4 compacts per session to zero.
            </>
          }
          author={
            <div>
              <div className="font-semibold text-confire-dim">Senior AI engineer</div>
              <div className="text-confire-muted">Series B startup, 40-person eng team</div>
            </div>
          }
        />
      </Container>
    </Section>
  )
}

// ── Pricing ───────────────────────────────────────────────────────────────────

const PLANS = [
  {
    name: 'Free',
    tagline: 'for hobby projects',
    price: '$0',
    period: '/month',
    features: ['500 cloud opts / mo', 'Local optimizations unlimited', 'Claude Code + Cursor hook', 'Usage dashboard'],
    cta: 'Start for free',
  },
  {
    name: 'Developer',
    tagline: 'for power users',
    price: '$10',
    period: '/month',
    features: ['5,000 cloud opts / mo', 'All source optimizers', 'Full history & analytics', 'Firewall group controls'],
    cta: 'Get Developer',
    featured: true,
  },
  {
    name: 'Pro',
    tagline: 'for heavy sessions',
    price: 'Coming soon',
    features: ['Higher limits', 'Session Memory Guard', 'PreCompact optimizer', 'Data export'],
    cta: 'Join waitlist',
    disabled: true,
  },
  {
    name: 'Enterprise',
    tagline: 'for teams',
    price: 'Coming soon',
    features: ['Team dashboard', 'Centralized billing', 'Custom optimizers', 'SSO / SAML', 'Audit logging'],
    cta: 'Talk to us',
    disabled: true,
  },
]

function Pricing() {
  return (
    <Section>
      <Container>
        <SectionTitle
          title={<>Pay only for<br />useful context.</>}
          subtitle="(Not to pad token counts.)"
        />

        <PricingSection plans={PLANS} />

        <p className="mt-4 text-center text-xs text-confire-muted">
          Annual billing and top-up packs coming soon.
        </p>
      </Container>
    </Section>
  )
}

// ── Bottom CTA — same orange card as hero ─────────────────────────────────────

function BottomCTA() {
  return (
    <Section className="confire-dot-region px-4 pb-6 sm:px-8">
      <Container>
        <CTASection
          title="Build without context limits."
          subtitle="Eliminate token noise and ship faster with Confire. Start building for free — no credit card required."
          primaryAction={
            <Button variant="white" asChild>
              <a href="/login">Start building for free</a>
            </Button>
          }
          secondaryAction={
            <Button variant="white-ghost" asChild>
              <a href="/docs">View docs</a>
            </Button>
          }
          marqueeItems={[
            { icon: <LightningIcon className={iconSm} weight="fill" />,   text: 'Claude Code + Cursor support' },
            { icon: <ChartLineUpIcon className={iconSm} weight="bold" />, text: 'Up to 93% token reduction' },
            { icon: <LockIcon className={iconSm} weight="bold" />,        text: 'Secrets never leave your machine' },
            { icon: <CalendarIcon className={iconSm} weight="bold" />,    text: '500 free cloud opts / month' },
            { icon: <PackageIcon className={iconSm} weight="bold" />,     text: 'Top-up packs when you need more' },
            { icon: <TimerIcon className={iconSm} weight="bold" />,       text: 'Install in under 60 seconds' },
          ]}
        />
      </Container>
    </Section>
  )
}

// ── Main export ───────────────────────────────────────────────────────────────

export function MarketingHome() {
  return (
    <>
      <SiteNav
        items={[
          { label: 'Products',  href: '#product'   },
          { label: 'Solutions', href: '#why'        },
          { label: 'Pricing',   href: '#pricing'   },
          { label: 'Docs',      href: '/docs'       },
        ]}
        ctaLabel="Get started free"
        ctaHref="/login"
      />

      <Hero />
      <StatsBar />
      <IntegrationBar />

      <div id="why">
        <WhyConfire />
      </div>

      <div id="product">
        <Pillars />
        <Capabilities />
        <HowItWorks />
      </div>

      <SocialProof />

      <div id="pricing">
        <Pricing />
      </div>

      <BottomCTA />
      <SiteFooter />
    </>
  )
}

/** Design system reference — all Confire UI primitives in one scroll. */
export function DesignSystemShowcase() {
  return (
    <>
      <SiteNav logo={<ConfireLogo wordmark="CONFIRE DS" />} />
      <Section>
        <Container>
          <SectionLabel number="DS">Component library</SectionLabel>
          <SectionTitle title="Confire design system" subtitle="Tailwind + CSS variable theme tokens." />
        </Container>
      </Section>
      <SiteFooter />
    </>
  )
}
