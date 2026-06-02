import { useState } from 'react'
import {
  Badge,
  Body,
  BodySm,
  BentoGrid,
  Button,
  Card,
  CTASection,
  Container,
  ConfireLogo,
  FeatureCard,
  H1,
  H2,
  H3,
  HorizontalTabs,
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

const iconSm  = 'size-3.5 shrink-0'
const iconMd  = 'size-8 shrink-0'
const iconLg  = 'size-12 shrink-0'

// ── Stats bar ─────────────────────────────────────────────────────────────────

const STATS = [
  { value: '93%',   label: 'average token reduction'  },
  { value: '500ms', label: 'median latency saved'      },
  { value: '13+',   label: 'source-specific optimizers'},
  { value: '∞',     label: 'local optimizations free'  },
]

function StatsBar() {
  return (
    <div className="border-y border-confire-border bg-confire-bg-card">
      <Container className="py-0">
        <div className="grid grid-cols-2 divide-x divide-confire-border sm:grid-cols-4">
          {STATS.map(({ value, label }) => (
            <div key={label} className="px-6 py-6 text-center">
              <div className="mb-1 font-sans text-3xl font-extrabold tracking-tight text-confire-accent">
                {value}
              </div>
              <div className="text-xs text-confire-muted">{label}</div>
            </div>
          ))}
        </div>
      </Container>
    </div>
  )
}

// ── Hero ──────────────────────────────────────────────────────────────────────

const installCode = `# One command. Hooks into every AI coding tool.
$ brew install confire/tap/confire
$ confire setup

✓ Hooked into Claude Code
✓ Hooked into Cursor
✓ Local optimizers active`

function Hero() {
  return (
    <Section className="pt-16 pb-0">
      <Container>
        <div className="grid items-center gap-12 lg:grid-cols-[1fr_480px]">
          {/* left */}
          <div>
            <Badge color="green" icon={<LightningIcon className={iconSm} weight="fill" />} className="mb-6">
              Now in public beta — join 1,200+ developers
            </Badge>

            <H1 className="mb-6">
              Make your AI agent
              <br />
              <span className="text-confire-accent">10× more focused.</span>
            </H1>

            <Body className="mb-8 max-w-lg text-lg">
              Confire sits between your tools and your model. It strips noise, compresses outputs,
              and keeps your context window sharp — so Claude and Cursor stop hallucinating on
              12,000-token bash dumps.
            </Body>

            <div className="mb-10 flex flex-wrap gap-3">
              <Button asChild>
                <a href="/login">
                  Get started free
                  <ArrowRightIcon className="size-4" />
                </a>
              </Button>
              <Button variant="outline" asChild>
                <a href="/docs/install">Read the docs</a>
              </Button>
            </div>

            {/* trust signals */}
            <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-confire-muted">
              {[
                'Free tier — no credit card',
                'Works with Claude Code & Cursor',
                'Local-first — your code stays yours',
              ].map(t => (
                <span key={t} className="flex items-center gap-1.5">
                  <CheckIcon className="size-3.5 text-confire-green" weight="bold" />
                  {t}
                </span>
              ))}
            </div>
          </div>

          {/* right — install snippet */}
          <div className="hidden h-[260px] overflow-hidden rounded-lg border border-confire-border bg-confire-code lg:block">
            <div className="flex h-9 items-center gap-2 border-b border-confire-border-subtle px-4">
              {['#f87171', '#fbbf24', '#4ade80'].map(c => (
                <span key={c} className="size-2.5 rounded-full" style={{ background: c }} />
              ))}
              <span className="ml-2 text-xs text-confire-code-tab-inactive">Terminal</span>
            </div>
            <pre className="overflow-auto p-5 font-mono text-[13px] leading-relaxed text-confire-code-text whitespace-pre">
              {installCode}
            </pre>
          </div>
        </div>
      </Container>
    </Section>
  )
}

// ── Trusted by / integration logos ────────────────────────────────────────────

const INTEGRATIONS = ['Claude Code', 'Cursor', 'VS Code', 'Windsurf', 'GitHub Copilot']

function IntegrationBar() {
  return (
    <Section className="py-10">
      <Container>
        <p className="mb-6 text-center text-xs font-semibold uppercase tracking-widest text-confire-muted">
          Works with the tools you already use
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-4">
          {INTEGRATIONS.map(name => (
            <span key={name} className="text-sm font-semibold text-confire-border-strong hover:text-confire-dim transition-colors">
              {name}
            </span>
          ))}
        </div>
      </Container>
    </Section>
  )
}

// ── Problem / solution split ──────────────────────────────────────────────────

function ProblemSolution() {
  return (
    <Section>
      <Container>
        <SectionLabel number="01">The problem</SectionLabel>
        <SectionTitle
          title="Your model is reading a novel when it needs a Post-it."
          subtitle="Every bash run, every file read, every Figma export floods your context. That noise costs tokens, slows responses, and causes hallucinations."
        />

        <TwoCards
          card1={
            <div>
              <Badge color="accent" className="mb-5">Before Confire</Badge>
              <div className="mb-3 font-mono text-3xl font-extrabold text-confire-text">12,400</div>
              <BodySm className="mb-4">tokens of raw bash output — stack traces, ANSI escapes, duplicate paths — flooding your context window every tool call.</BodySm>
              <div className="flex flex-col gap-2 text-xs text-confire-muted">
                {['Token budget wasted on noise', 'Model loses track of the task', 'Higher costs per session'].map(t => (
                  <span key={t} className="flex items-center gap-2">
                    <span className="size-1.5 rounded-full bg-red-500/60 shrink-0" />
                    {t}
                  </span>
                ))}
              </div>
            </div>
          }
          card2={
            <div>
              <Badge color="green" className="mb-5">After Confire</Badge>
              <div className="mb-3 font-mono text-3xl font-extrabold text-confire-accent">890</div>
              <BodySm className="mb-4">tokens — same signal, 93% less noise. Exit code, stderr summary, changed files. Your model stays sharp and on-task.</BodySm>
              <div className="flex flex-col gap-2 text-xs text-confire-muted">
                {['93% fewer tokens per tool call', 'Model focus stays on the task', '~50% lower session costs'].map(t => (
                  <span key={t} className="flex items-center gap-2">
                    <CheckIcon className="size-3.5 text-confire-green shrink-0" weight="bold" />
                    {t}
                  </span>
                ))}
              </div>
            </div>
          }
        />
      </Container>
    </Section>
  )
}

// ── Core pillars ──────────────────────────────────────────────────────────────

function Pillars() {
  return (
    <Section>
      <Container>
        <SectionLabel number="02">How it works</SectionLabel>
        <SectionTitle
          title="A firewall for your context window."
          subtitle="Confire intercepts every tool response before it reaches the model. Local rules run instantly. Cloud optimizers handle the complex stuff."
        />

        <ThreeCards
          cards={[
            <FeatureCard
              key="local"
              icon={<TerminalWindowIcon className={iconMd} weight="duotone" />}
              title="Local-first, zero latency"
              description="Bash, file reads, search results — optimized on-device in microseconds. No round-trip, no network dependency, no latency added to your session."
            />,
            <FeatureCard
              key="cloud"
              icon={<CloudIcon className={iconMd} weight="duotone" />}
              title="13+ source-specific optimizers"
              description="GitHub PRs, Figma exports, Slack threads, npm audit output — each optimizer knows the exact shape of its data and strips only the noise."
            />,
            <FeatureCard
              key="privacy"
              icon={<LockIcon className={iconMd} weight="duotone" />}
              title="Privacy by default"
              description="Secrets and credentials are redacted before anything leaves your machine. Telemetry is opt-in. Your code never trains our models."
            />,
          ]}
        />
      </Container>
    </Section>
  )
}

// ── Bento — capabilities ──────────────────────────────────────────────────────

function Capabilities() {
  return (
    <Section>
      <Container>
        <SectionLabel number="03">Capabilities</SectionLabel>
        <SectionTitle
          title="Everything you need to run lean, fast agent sessions."
        />

        <BentoGrid
          items={[
            {
              colSpan: 2,
              content: (
                <div>
                  <CpuIcon className="mb-4 size-8 text-confire-accent" weight="duotone" />
                  <H3 className="mb-2">MCP Firewall</H3>
                  <BodySm>Scans every MCP tool response for prompt injection, hidden Unicode, and leaked credentials — before the model ever sees it. Configurable rule groups on Dev and Pro.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <DatabaseIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Session Memory Guard</H3>
                  <BodySm>Watches context growth and warns before you hit the limit. Pro plan adds automatic pre-compact optimization.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <ShieldCheckIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Secret redaction</H3>
                  <BodySm>Detects and strips API keys, tokens, and credentials from tool output before compression. Zero config.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <ChartLineUpIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Usage analytics</H3>
                  <BodySm>Token savings per session, per tool, per integration. Local stats always available — full history in your dashboard on Dev+.</BodySm>
                </div>
              ),
            },
            {
              accent: true,
              content: (
                <div>
                  <BracketsCurlyIcon className="mb-4 size-8 text-white/80" weight="bold" />
                  <H3 className="mb-2 text-white">Open hook API</H3>
                  <BodySm className="text-white/70">Write custom optimizers in TypeScript. Ship them as local rules or private cloud adapters.</BodySm>
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

const setupCode = `$ brew install confire/tap/confire
$ confire setup
$ confire login

✓ Hooked into Claude Code  (v1.8+)
✓ Hooked into Cursor       (v0.44+)
✓ Local optimizers active`

const optimizeCode = `// Confire runs automatically — no code changes needed.
// Every tool call is intercepted and compressed:

Bash output      12,400 tokens → 890  tokens  (93% saved)
GitHub PR diff   8,200  tokens → 620  tokens  (92% saved)
Figma export     15,000 tokens → 1,100 tokens  (93% saved)

// Session cost:  $0.84  →  $0.06`

const statsCode = `$ confire stats

  124,500  tokens saved  (this week)
    2,840  tool calls processed
    ↓ 91%  average reduction

  Top savers:
    bash        47,200 tokens
    file_read   38,100 tokens
    github_pr   21,900 tokens

  Dashboard → confire.dev/dashboard`

function HowItWorks() {
  return (
    <Section>
      <Container>
        <SectionLabel number="04">Get started in 60 seconds</SectionLabel>
        <SectionTitle
          title="Install once. Optimize everything."
          subtitle="Confire hooks into your existing workflow as a Claude Code post-tool hook. No proxy, no port forwarding, no config files."
        />

        <VerticalTabsCode
          tabs={[
            {
              title: 'Install the CLI',
              description: 'One command. Hooks into Claude Code, Cursor, and VS Code automatically.',
              code: setupCode,
            },
            {
              title: 'Optimization runs automatically',
              description: 'Every tool call is intercepted and compressed. You do nothing differently.',
              code: optimizeCode,
            },
            {
              title: 'Track your savings',
              description: 'See exactly where your tokens are going — locally or in your dashboard.',
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
              <span className="text-confire-text font-semibold">stays on task now</span>.
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

function Pricing() {
  const [annual, setAnnual] = useState(false)

  const plans = [
    {
      name: 'Free',
      tagline: 'For individuals getting started',
      price: '$0',
      period: '/mo',
      features: [
        '500 cloud opts / month',
        'Universal optimizer',
        'Local optimizers (unlimited)',
        'Claude Code hook',
        'Basic stats',
      ],
      cta: 'Get started free',
    },
    {
      name: 'Dev',
      tagline: 'For power users and daily drivers',
      price: annual ? '$7.92' : '$10',
      period: '/mo',
      features: [
        '5,000 cloud opts / month',
        '13 source-specific optimizers',
        'Larger input payloads',
        'Optimization history',
        'Remote optimizer updates',
        'Early access to new adapters',
      ],
      cta: 'Start free trial',
      featured: true,
    },
    {
      name: 'Pro',
      tagline: 'For heavy sessions and teams',
      price: annual ? '$16.25' : '$20',
      period: '/mo',
      features: [
        'Unlimited fair-use cloud opts',
        'Session Memory Guard',
        'PreCompact optimizer',
        'Local memory packs',
        'Priority optimizer updates',
        'Data export',
      ],
      cta: 'Upgrade to Pro',
    },
    {
      name: 'Enterprise',
      tagline: 'Self-hosted, custom policies',
      price: 'Custom',
      features: [
        'Self-hosted deployment',
        'Local-only optimizer mode',
        'Team policy controls',
        'Custom optimizers',
        'Audit logging',
        'SSO / SAML (roadmap)',
        'Priority support',
      ],
      cta: 'Contact sales',
    },
  ]

  return (
    <Section>
      <Container>
        <SectionLabel number="05">Pricing</SectionLabel>
        <SectionTitle
          title="Simple pricing. Serious savings."
          subtitle="Start free — no credit card required. Upgrade when you need more cloud optimizations or source-specific adapters."
        />

        <HorizontalTabs
          tabs={[
            {
              label: 'Monthly',
              icon: <CalendarIcon className={iconSm} weight="bold" />,
              content: <PricingSection plans={plans} />,
            },
            {
              label: 'Annual — save 20%',
              icon: <HexagonIcon className={iconSm} weight="bold" />,
              content: <PricingSection plans={plans} />,
            },
          ]}
          defaultTab={annual ? 1 : 0}
        />

        {/* comparison table */}
        <div className="mt-16">
          <SectionTitle
            title="Compare plans"
            subtitle={undefined}
            center
          />
          <ComparisonTable />
        </div>
      </Container>
    </Section>
  )
}

const TABLE_ROWS: { label: string; free: string | boolean; dev: string | boolean; pro: string | boolean; enterprise: string | boolean }[] = [
  { label: 'Cloud optimizations / month', free: '500',        dev: '5,000',      pro: 'Unlimited*', enterprise: 'Unlimited' },
  { label: 'Local optimizations',         free: true,         dev: true,         pro: true,          enterprise: true },
  { label: 'Source-specific optimizers',  free: 'Generic only',dev: '13+',       pro: '13+',         enterprise: 'Custom' },
  { label: 'MCP Firewall',                free: 'Basic',      dev: 'Configurable',pro: 'Configurable',enterprise: 'Custom rules' },
  { label: 'Session Memory Guard',        free: false,        dev: false,        pro: true,          enterprise: true },
  { label: 'PreCompact optimizer',        free: false,        dev: false,        pro: true,          enterprise: true },
  { label: 'Optimization history',        free: false,        dev: true,         pro: true,          enterprise: true },
  { label: 'Data export',                 free: false,        dev: false,        pro: true,          enterprise: true },
  { label: 'Self-hosted deployment',      free: false,        dev: false,        pro: false,         enterprise: true },
  { label: 'Team policy controls',        free: false,        dev: false,        pro: false,         enterprise: true },
  { label: 'SSO / SAML',                  free: false,        dev: false,        pro: false,         enterprise: 'Roadmap' },
  { label: 'Priority support',            free: false,        dev: false,        pro: false,         enterprise: true },
]

function TableCell({ value }: { value: string | boolean }) {
  if (value === true) return <CheckIcon className="mx-auto size-4 text-confire-green" weight="bold" />
  if (value === false) return <span className="text-confire-border-strong">—</span>
  return <span className="text-sm text-confire-dim">{value}</span>
}

function ComparisonTable() {
  const cols = ['', 'Free', 'Dev', 'Pro', 'Enterprise']
  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse text-center text-sm">
        <thead>
          <tr className="border-b border-confire-border">
            {cols.map((c, i) => (
              <th key={c} className={`py-3 font-semibold ${i === 0 ? 'text-left text-confire-muted w-52' : i === 2 ? 'text-confire-accent' : 'text-confire-text'}`}>
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {TABLE_ROWS.map(({ label, free, dev, pro, enterprise }) => (
            <tr key={label} className="border-b border-confire-border-subtle hover:bg-confire-hover-subtle">
              <td className="py-3 text-left text-confire-secondary">{label}</td>
              <td className="py-3"><TableCell value={free} /></td>
              <td className="py-3 bg-confire-accent-dim"><TableCell value={dev} /></td>
              <td className="py-3"><TableCell value={pro} /></td>
              <td className="py-3"><TableCell value={enterprise} /></td>
            </tr>
          ))}
        </tbody>
      </table>
      <p className="mt-4 text-xs text-confire-muted">* Monthly token cap and per-minute rate limits apply.</p>
    </div>
  )
}

// ── Bottom CTA ────────────────────────────────────────────────────────────────

function BottomCTA() {
  return (
    <Section>
      <Container>
        <CTASection
          title="Start saving tokens today."
          subtitle="Install in under a minute. Free tier includes 500 cloud optimizations and unlimited local optimizations every month."
          primaryAction={
            <Button variant="white" asChild>
              <a href="/login">
                Get started free
                <ArrowRightIcon className="size-4" />
              </a>
            </Button>
          }
          secondaryAction={
            <Button variant="white-ghost" asChild>
              <a href="/docs/install">Read the docs</a>
            </Button>
          }
          marqueeItems={[
            { icon: <LightningIcon className={iconSm} weight="fill" />,    text: 'Works with Claude Code & Cursor' },
            { icon: <ChartLineUpIcon className={iconSm} weight="bold" />,  text: '93% average token reduction' },
            { icon: <LockIcon className={iconSm} weight="bold" />,         text: 'Secrets never leave your machine' },
            { icon: <CalendarIcon className={iconSm} weight="bold" />,     text: '500 free cloud opts / month' },
            { icon: <PackageIcon className={iconSm} weight="bold" />,      text: 'Top-up packs when you need more' },
            { icon: <TimerIcon className={iconSm} weight="bold" />,        text: 'Install in under 60 seconds' },
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
          { label: 'Product',  href: '#product'  },
          { label: 'Pricing',  href: '#pricing'  },
          { label: 'Docs',     href: '/docs'      },
          { label: 'Changelog',href: '/changelog' },
        ]}
        ctaLabel="Get started free"
        ctaHref="/login"
      />

      <Hero />
      <StatsBar />
      <IntegrationBar />

      <div id="product">
        <ProblemSolution />
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

          <div className="mb-16 flex flex-wrap gap-3">
            <Button>Primary</Button>
            <Button variant="outline">Outline</Button>
            <Button variant="ghost">Ghost</Button>
            <Button variant="outline-accent">Accent outline</Button>
            <Badge color="green">Green badge</Badge>
            <Badge color="accent">Accent badge</Badge>
          </div>

          <div className="mb-16 grid gap-8 lg:grid-cols-2">
            <Card footer={<Button variant="outline-sm">Action</Button>}>
              <H3 className="mb-2">Single card</H3>
              <BodySm>Corner squares + border. Override --confire-accent in CSS to retheme.</BodySm>
            </Card>
            <BentoGrid
              items={[
                { colSpan: 2, content: <BodySm>Bento wide cell</BodySm> },
                { content: <BodySm>Cell</BodySm> },
                { accent: true, content: <BodySm className="text-white">Accent</BodySm> },
              ]}
            />
          </div>

          <HorizontalTabs
            tabs={[
              { label: 'Tab A', content: <BodySm>Tab A content</BodySm> },
              { label: 'Tab B', content: <BodySm>Tab B content</BodySm> },
            ]}
          />
        </Container>
      </Section>
      <SiteFooter />
    </>
  )
}
