import React, { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import { EarlyAccessForm } from '@/components/marketing/EarlyAccessForm'
import {
  BodySm,
  BentoGrid,
  Button,
  CTASection,
  Container,
  FeatureCard,
  H1,
  H3,
  PricingSection,
  Section,
  SectionLabel,
  SectionTitle,
  SiteFooter,
  SiteNav,
  ThreeCards,
  VerticalTabsCode,
  WithCorners,
} from '@/components/confire'
import {
  ArrowRightIcon,
  BookOpenIcon,
  BracketsCurlyIcon,
  CalendarIcon,
  ChartLineUpIcon,
  CheckIcon,
  CpuIcon,
  DatabaseIcon,
  FingerprintIcon,
  GearSixIcon,
  HexagonIcon,
  LightningIcon,
  LockIcon,
  PackageIcon,
  ProhibitIcon,
  ShieldCheckIcon,
  TerminalWindowIcon,
  TimerIcon,
  UsersIcon,
  WarningIcon,
} from '@phosphor-icons/react'

const iconSm = 'size-3.5 shrink-0'
const iconMd = 'size-8 shrink-0'

// ── Hero ──────────────────────────────────────────────────────────────────────

const DEMO_TABS = [
  { id: 'risky',  label: 'Risky command', Icon: ProhibitIcon },
  { id: 'mcp',    label: 'MCP output',    Icon: CpuIcon },
  { id: 'bash',   label: 'Bash logs',     Icon: TerminalWindowIcon },
  { id: 'figma',  label: 'Figma',         Icon: HexagonIcon },
  { id: 'github', label: 'GitHub PR',     Icon: PackageIcon },
  { id: 'docs',   label: 'Docs fetch',    Icon: BookOpenIcon },
] as const

type DemoTabId = (typeof DEMO_TABS)[number]['id']

const DEMO_HEADER: Record<DemoTabId, string> = {
  risky:  'confire · firewall active',
  mcp:    'confire · context firewall',
  bash:   'confire · context firewall',
  figma:  'confire · context firewall',
  github: 'confire · context firewall',
  docs:   'confire · context firewall',
}

const DEMO_OUTPUTS: Record<DemoTabId, string> = {
  risky: `pretool: bash_execute
command: git push --force-with-lease origin main

CONFIRE REVIEW REQUIRED

rule:    Review force pushes
risk:    rewrites remote branch history
         and can affect open PRs
action:  run \`confire bypass-next\` to allow
         once, then retry`,

  mcp: `posttool: mcp__figma__get_node

  before   228,906 tokens
  after      4,717 tokens
  98% saved

  sanitized:
    secrets redacted:           0
    hidden instructions removed: 0
  kept:
    layout, spacing, colors, typography,
    component states`,

  bash: `posttool: bash_execute

  before   12,400 tokens  ████████████████████
  after       890 tokens  █▌

  93% saved

  stripped: node_modules listing    (8,200 tok)
  stripped: repeated stack traces   (2,100 tok)
  stripped: env dump headers        (1,210 tok)
  kept:     actual command output`,

  figma: `posttool: mcp__figma__get_file

  before   22,000 tokens  ████████████████████
  after     1,600 tokens  █▌

  93% saved

  stripped: SVG path metadata      (12,000 tok)
  stripped: redundant style rules   (6,200 tok)
  stripped: hidden/locked layers    (2,200 tok)
  kept:     component tree and tokens`,

  github: `posttool: github_get_pull_request

  before   15,000 tokens  ████████████████████
  after     1,100 tokens  █▌

  93% saved

  stripped: CI check run logs       (9,400 tok)
  stripped: generated file diffs    (3,800 tok)
  stripped: bot comment threads       (700 tok)
  kept:     code changes and reviews`,

  docs: `posttool: web_fetch

  before    9,800 tokens  ████████████████████
  after       720 tokens  █▌

  93% saved

  stripped: nav, footer, sidebar    (4,200 tok)
  stripped: repeated boilerplate    (3,100 tok)
  stripped: duplicate code examples (1,780 tok)
  kept:     main content, headings`,
}

function Hero() {
  const [activeTab, setActiveTab] = useState<DemoTabId>('risky')
  const [displayed, setDisplayed] = useState('')
  const [cursorOn, setCursorOn]   = useState(true)

  useEffect(() => {
    const full = DEMO_OUTPUTS[activeTab]
    setDisplayed('')
    let i = 0
    const id = setInterval(() => {
      i++
      setDisplayed(full.slice(0, i))
      if (i >= full.length) clearInterval(id)
    }, 11)
    return () => clearInterval(id)
  }, [activeTab])

  useEffect(() => {
    const id = setInterval(() => setCursorOn(c => !c), 520)
    return () => clearInterval(id)
  }, [])

  return (
    <Section className="confire-dot-region px-0 pt-24 pb-0">
      <div className="px-4 text-center sm:px-8">
        <H1 className="mb-6">
          Keep AI coding agents<br className="hidden sm:block" />
          cleaner and safer.
        </H1>
        <p className="mx-auto mb-4 max-w-[40rem] text-base leading-relaxed text-confire-muted">
          Confire reviews risky tool calls before they run and sanitizes noisy MCP,
          Bash, Figma, GitHub, docs, and API output before it enters context.
        </p>
        <div className="mb-10 flex flex-wrap justify-center gap-3">
          <Button variant="outline" asChild>
            <a href="/login">
              Start free
              <ArrowRightIcon className="size-4" weight="bold" />
            </a>
          </Button>
          <Button variant="ghost-sm" asChild>
            <a href="#how-it-works">See how it works</a>
          </Button>
        </div>
        <p className="mb-1 text-sm text-confire-muted">
          Works with Claude Code, Cursor, and VS Code.
        </p>
        <p className="mx-auto max-w-[36rem] text-xs text-confire-border-strong">
          Claude Code supports full hook-based firewall mode. Cursor and VS Code use MCP
          gateway mode for tools routed through Confire.
        </p>
      </div>

      <div className="mt-14 flex flex-wrap justify-center gap-2 px-4">
        {DEMO_TABS.map(({ id, label, Icon }) => (
          <button
            key={id}
            onClick={() => setActiveTab(id)}
            className={cn(
              'flex cursor-pointer items-center gap-2 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-200',
              activeTab === id
                ? 'border-confire-border-strong bg-confire-card text-confire-text'
                : 'border-confire-border bg-transparent text-confire-muted hover:border-confire-border-mid hover:text-confire-text',
            )}
          >
            <Icon className="size-3.5" weight={activeTab === id ? 'duotone' : 'regular'} />
            {label}
          </button>
        ))}
      </div>

      <div className="mx-auto mt-6 w-full max-w-[var(--confire-max-w)] px-4 sm:px-8">
        <div className="relative overflow-hidden rounded-t-2xl border border-b-0 border-confire-border bg-confire-card">
          <div className="flex items-center gap-1.5 border-b border-confire-border px-5 py-3">
            <span className="size-2.5 rounded-full bg-confire-border-strong" />
            <span className="size-2.5 rounded-full bg-confire-border-strong" />
            <span className="size-2.5 rounded-full bg-confire-border-strong" />
            <span className="ml-3 font-mono text-[11px] text-confire-muted">
              {DEMO_HEADER[activeTab]}
            </span>
            <span className="ml-auto flex items-center gap-1.5 text-[11px] text-confire-accent">
              <span className="size-1.5 animate-pulse rounded-full bg-confire-accent" />
              live
            </span>
          </div>

          <div className="relative z-10 min-h-[180px] p-7 pb-3">
            <pre className="whitespace-pre-wrap font-mono text-[13px] leading-relaxed text-confire-text-dim">
              {displayed}
              <span
                className="ml-px inline-block w-[6px] translate-y-[1px] bg-confire-accent align-text-top"
                style={{ height: '1em', opacity: cursorOn ? 1 : 0, transition: 'opacity 0.08s' }}
              />
            </pre>
          </div>

          <div className="relative h-52 overflow-hidden">
            <div
              className="pointer-events-none absolute left-1/2 -translate-x-1/2 animate-pulse"
              style={{
                bottom: '-80px',
                width: '600px',
                height: '320px',
                background: 'radial-gradient(ellipse at 50% 80%, rgba(244,129,31,0.16) 0%, rgba(244,129,31,0.04) 45%, transparent 68%)',
                animationDuration: '3s',
              }}
            />
            {[0, 1, 2, 3, 4, 5, 6, 7].map(i => {
              const size = 120 + i * 130
              return (
                <div
                  key={i}
                  className="pointer-events-none absolute left-1/2 -translate-x-1/2 rounded-full"
                  style={{
                    width: size,
                    height: size,
                    bottom: -(size * 0.56),
                    border: `1px solid rgba(244,129,31,${Math.max(0.025, 0.22 - i * 0.026)})`,
                  }}
                />
              )
            })}
            <div
              className="pointer-events-none absolute inset-x-0 top-0 h-12"
              style={{ background: 'linear-gradient(to bottom, var(--confire-bg-card), transparent)' }}
            />
          </div>
        </div>
      </div>
    </Section>
  )
}

// ── Stats bar ─────────────────────────────────────────────────────────────────

const STATS = [
  { value: 'Built-in',    label: 'risky action guardrails'    },
  { value: '5 layers',    label: 'of output protection'           },
  { value: 'Local-first', label: 'policy evaluation'          },
  { value: 'Sanitized',   label: 'before context reaches the model'  },
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
                <div className="mb-1.5 font-sans text-[1.6rem] font-extrabold tracking-tight text-confire-accent leading-tight">
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

// ── Problem ───────────────────────────────────────────────────────────────────

const CODE_ITEMS: Array<{ label: string; items: string[] }> = [
  {
    label: 'Risky tool calls',
    items: [
      'git push --force',
      'gh pr close',
      'supabase db reset',
      'mcp__stripe__create_refund',
    ],
  },
  {
    label: 'Noisy tool output',
    items: [
      'Figma trees',
      'GitHub API blobs',
      'Bash logs',
      'WebFetch pages',
      'MCP JSON dumps',
    ],
  },
  {
    label: 'Untrusted context',
    items: [
      'hidden instructions',
      'prompt-injection-like text',
      'secret-looking values',
      'credential-lure content',
    ],
  },
]

const PROBLEM_DESCRIPTIONS = [
  'Confire reviews or blocks risky actions before they run.',
  'Confire sanitizes and filters output before it reaches context.',
  'Confire treats tool output as data, not instructions.',
]

function Problem() {
  return (
    <Section className="confire-dot-region">
      <Container>
        <SectionLabel number="01">The agent tool problem</SectionLabel>
        <SectionTitle
          title="AI agents can call tools. But tools create risk and noise."
          subtitle="Modern coding agents can read files, run shell commands, call MCP servers, fetch docs, inspect Figma, review PRs, and touch external systems. That power creates two problems: risky actions and context pollution. Confire sits between the agent and its tools."
        />

        <ThreeCards
          cards={CODE_ITEMS.map(({ label, items }, i) => (
            <div key={label}>
              <div className="mb-4 text-[11px] font-bold uppercase tracking-widest text-confire-muted">
                {label}
              </div>
              <div className="mb-5 space-y-1.5">
                {items.map(item => (
                  <div
                    key={item}
                    className="rounded border border-confire-border bg-confire-bg px-3 py-1.5 font-mono text-xs text-confire-text-dim"
                  >
                    {item}
                  </div>
                ))}
              </div>
              <p className="text-xs text-confire-muted">{PROBLEM_DESCRIPTIONS[i]}</p>
            </div>
          ))}
        />
      </Container>
    </Section>
  )
}

// ── How It Works ──────────────────────────────────────────────────────────────

const HOW_STEPS = [
  {
    step: '1',
    title: 'Before tools run',
    icon: <ShieldCheckIcon className="size-8" weight="duotone" />,
    body: 'Confire evaluates the tool call before execution. It can allow, warn, review, or block based on built-in rules and your custom policy.',
    examples: ['Review force pushes', 'Review mutating MCP tools', 'Review database resets', 'Block repository deletion'],
  },
  {
    step: '2',
    title: 'After tools return',
    icon: <DatabaseIcon className="size-8" weight="duotone" />,
    body: 'Confire sanitizes the tool result before it enters the agent working context. It can redact common secrets, remove suspicious hidden instructions, and normalize noisy output.',
    examples: ['Redact API keys', 'Remove hidden prompt-injection-like text', 'Pack huge MCP JSON', 'Clean repeated Bash logs'],
  },
  {
    step: '3',
    title: 'The agent gets clean context',
    icon: <LightningIcon className="size-8" weight="duotone" />,
    body: 'The model receives only the sanitized, filtered, task-ready result. Less noise. Fewer wasted tokens. Safer tool use.',
    examples: [],
  },
]

function HowItWorks() {
  return (
    <Section id="how-it-works">
      <Container>
        <SectionLabel number="02">How Confire works</SectionLabel>
        <SectionTitle
          title="A firewall around every tool call."
          subtitle="Confire intercepts at two points: before a tool runs and after it returns."
        />

        <ThreeCards
          cards={HOW_STEPS.map(({ step, title, icon, body, examples }) => (
            <div key={step}>
              <div className="mb-3 flex items-center gap-3">
                <div className="text-confire-dim">{icon}</div>
                <div className="font-mono text-xs text-confire-muted">Step {step}</div>
              </div>
              <H3 className="mb-3">{title}</H3>
              <p className="mb-4 text-sm leading-relaxed text-confire-muted">{body}</p>
              {examples.length > 0 && (
                <div className="space-y-1">
                  {examples.map(ex => (
                    <div key={ex} className="flex items-center gap-2 text-xs text-confire-border-strong">
                      <span className="size-1 rounded-full bg-confire-accent shrink-0" />
                      {ex}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        />
      </Container>
    </Section>
  )
}

// ── Capabilities ──────────────────────────────────────────────────────────────

function Capabilities() {
  return (
    <Section>
      <Container>
        <SectionLabel number="03">Capabilities</SectionLabel>
        <SectionTitle
          title="Context control and tool safety in one local layer."
          subtitle="Every capability runs on your machine. The cloud receives only sanitized, redacted metadata — never raw tool output."
        />

        <BentoGrid
          items={[
            {
              colSpan: 2,
              content: (
                <div>
                  <ShieldCheckIcon className="mb-4 size-8 text-confire-accent" weight="duotone" />
                  <H3 className="mb-2">Tool Firewall</H3>
                  <BodySm>Review risky commands before they run. Built-in rules cover common Git, GitHub, shell, database, deploy, package publish, and mutating MCP actions.</BodySm>
                  <div className="mt-4 flex flex-wrap gap-1.5">
                    {['git push --force', 'git reset --hard', 'gh pr merge', 'terraform destroy', 'mcp__*__delete_*'].map(ex => (
                      <span key={ex} className="rounded border border-confire-border bg-confire-bg px-2 py-0.5 font-mono text-[11px] text-confire-text-dim">{ex}</span>
                    ))}
                  </div>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <CpuIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">MCP Firewall</H3>
                  <BodySm>Apply generic risk scoring, sanitization, and context budgeting to unknown MCP servers. Works even when there is no source-specific handler yet.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <ShieldCheckIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Context Firewall</H3>
                  <BodySm>Redacts secrets, removes injections, and trims noise from every tool output before it enters context. Works on Bash logs, WebFetch pages, GitHub PRs, Figma outputs, and MCP responses.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <LockIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Secret Redaction</H3>
                  <BodySm>Redact common secret-looking values before any downstream processing. All subsequent context passes receive only sanitized, redacted content.</BodySm>
                </div>
              ),
            },
            {
              content: (
                <div>
                  <WarningIcon className="mb-4 size-8 text-confire-dim" weight="duotone" />
                  <H3 className="mb-2">Prompt-Injection Sanitization</H3>
                  <BodySm>Detect and sandbox common instruction-like patterns in untrusted tool output. Fetched pages, MCP results, and external content are treated as data, not commands.</BodySm>
                </div>
              ),
            },
            {
              accent: true,
              content: (
                <div>
                  <GearSixIcon className="mb-4 size-8 text-white/80" weight="bold" />
                  <H3 className="text-white mb-2">Custom Guardrails</H3>
                  <BodySm className="text-white/70">Dev users can define custom rules from the dashboard and sync them locally. Rules are evaluated on your machine.</BodySm>
                </div>
              ),
            },
          ]}
        />
      </Container>
    </Section>
  )
}

// ── Client cards ──────────────────────────────────────────────────────────────

const CLIENTS = [
  {
    name: 'Claude Code',
    mode: 'Full firewall mode',
    Icon: TerminalWindowIcon,
    features: [
      'PreToolUse risky action review',
      'PostToolUse output sanitization',
      'Context firewall: redaction, injection removal, noise trimming',
      'Built-in and custom rules',
      'Local policy evaluation',
    ],
  },
  {
    name: 'Cursor',
    mode: 'MCP gateway mode',
    Icon: HexagonIcon,
    features: [
      'Protects tools routed through Confire',
      'Sanitizes and filters MCP output',
      'Supports Confire policy rules on routed tools',
      'Advisory context where supported',
    ],
  },
  {
    name: 'VS Code',
    mode: 'MCP gateway mode',
    Icon: PackageIcon,
    features: [
      'Protects tools routed through Confire',
      'Sanitizes and filters MCP output',
      'Supports Confire policy rules on routed tools',
      'Advisory context where supported',
    ],
  },
]

function Clients() {
  return (
    <Section>
      <Container>
        <SectionLabel number="04">Works with your agent workflow</SectionLabel>
        <SectionTitle
          title="Start with Claude Code. Extend through MCP."
        />

        <ThreeCards
          cards={CLIENTS.map(({ name, mode, Icon, features }) => (
            <div key={name}>
              <div className="mb-1 flex items-center gap-3">
                <Icon className="size-6 text-confire-dim" weight="duotone" />
                <div className="text-base font-bold text-confire-text">{name}</div>
              </div>
              <div className="mb-5 text-xs font-semibold text-confire-accent">{mode}</div>
              <ul className="space-y-2">
                {features.map(f => (
                  <li key={f} className="flex items-start gap-2 text-xs text-confire-muted">
                    <CheckIcon className="mt-0.5 size-3 shrink-0 text-confire-accent" weight="bold" />
                    {f}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        />

        <p className="mt-6 text-center text-xs text-confire-border-strong">
          Claude Code supports full hook-based enforcement. Cursor and VS Code support
          Confire-routed MCP tools and advisory firewall behavior where supported.
        </p>
      </Container>
    </Section>
  )
}

// ── Results ───────────────────────────────────────────────────────────────────

const RESULTS = [
  { value: '45.2%',    label: 'less tool-result context in an early real Claude Code benchmark.' },
  { value: '24.8%',    label: 'lower measured Claude API cost in an early Figma workflow with the same number of API turns.' },
  { value: '98%',      label: 'reduction on a Figma tool output in local testing.' },
  { value: 'Built-in', label: 'guardrails for risky Git, MCP, database, deploy, and shell actions.' },
]

function Results() {
  return (
    <Section className="confire-dot-region">
      <Container>
        <SectionLabel number="05">Early results</SectionLabel>
        <SectionTitle
          title="Less noisy context. More controlled tool use."
          subtitle="Confire reduces the parts of agent workflows that waste context: repeated logs, giant API responses, verbose MCP output, full design trees, and fetched pages with irrelevant boilerplate."
        />

        <WithCorners cols={4} rows={1}>
          <div className="grid grid-cols-2 border border-confire-border xl:grid-cols-4">
            {RESULTS.map(({ value, label }, i) => (
              <div
                key={value + i}
                className={`px-6 py-8${i < RESULTS.length - 1 ? ' border-b border-confire-border xl:border-b-0 xl:border-r' : ''}`}
              >
                <div className="mb-2 font-sans text-[2rem] font-extrabold tracking-tight text-confire-accent leading-tight">
                  {value}
                </div>
                <div className="text-xs leading-relaxed text-confire-muted">{label}</div>
              </div>
            ))}
          </div>
        </WithCorners>

        <p className="mt-6 text-center text-xs text-confire-border-strong">
          Benchmarks vary by tool, model, and workflow. Confire is most effective in
          tool-heavy sessions with noisy MCP, Bash, Figma, GitHub, docs, API, and log output.
        </p>
      </Container>
    </Section>
  )
}

// ── Privacy ───────────────────────────────────────────────────────────────────

const TRUST_BULLETS = [
  'Tool inputs are evaluated locally.',
  'Built-in rules work offline.',
  'Secret redaction runs locally before any data leaves your machine.',
  'Prompt-injection sanitization runs locally on every MCP response.',
  'All context passes receive only sanitized, redacted content.',
  'Raw tool inputs and outputs are not sent as telemetry.',
  'Aggregate usage metadata powers your dashboard.',
]

function Privacy() {
  return (
    <Section>
      <Container>
        <SectionLabel number="06">Local-first by default</SectionLabel>
        <SectionTitle
          title="Your policy runs locally. Raw tool inputs do not go to Confire Cloud."
          subtitle="Confire is built for developer trust. Rule evaluation happens on your machine. Built-in guardrails work without an account. Custom rules are synced locally and evaluated locally."
        />

        <WithCorners cols={1} rows={1}>
          <div className="border border-confire-border bg-confire-card p-8">
            <div className="grid gap-3 sm:grid-cols-2">
              {TRUST_BULLETS.map(bullet => (
                <div key={bullet} className="flex items-start gap-3">
                  <CheckIcon className="mt-0.5 size-4 shrink-0 text-confire-accent" weight="bold" />
                  <span className="text-sm text-confire-muted">{bullet}</span>
                </div>
              ))}
            </div>
            <p className="mt-6 border-t border-confire-border pt-5 text-xs text-confire-border-strong">
              Confire is a guardrail layer, not a perfect security boundary. It detects
              common risky actions, secret patterns, and suspicious context patterns, and
              gives you control before agents act.
            </p>
          </div>
        </WithCorners>
      </Container>
    </Section>
  )
}

// ── Setup ─────────────────────────────────────────────────────────────────────

const installCode = `curl -fsSL https://confire.dev/install.sh | sh`

const connectCode = `confire login
confire install claude
confire install cursor
confire install vscode
confire doctor`

const testCode = `confire policy test 'git push --force'

Action:   review
Rule:     Review force push
Severity: high
Reason:   Force push can rewrite remote branch history
          and affect open PRs.
Source:   builtin`

function Setup() {
  return (
    <Section>
      <Container>
        <SectionLabel number="07">Get started</SectionLabel>
        <SectionTitle
          title="Install once. Protect every supported session."
        />

        <VerticalTabsCode
          tabs={[
            {
              title: 'Install the CLI',
              description: 'One command. macOS and Linux supported at launch.',
              lang: 'bash',
              code: installCode,
            },
            {
              title: 'Connect your agents',
              description: 'Log in and hook Confire into each supported client.',
              lang: 'bash',
              code: connectCode,
            },
            {
              title: 'Test a policy',
              description: 'Verify that firewall rules are active.',
              lang: 'bash',
              code: testCode,
            },
          ]}
        />
      </Container>
    </Section>
  )
}

// ── Pricing ───────────────────────────────────────────────────────────────────

const PLANS = [
  {
    name: 'Free',
    tagline: 'Available now — try Confire’s local context and tool firewall for Claude Code, Cursor, and VS Code.',
    price: '$0',
    period: '/month',
    features: [
      'Claude Code full firewall',
      'Cursor + VS Code MCP gateway',
      'Tool Firewall + Context Firewall',
      'Secret redaction',
      'Injection guard',
      'Local context passes',
      'Basic security event stats',
    ],
    cta: 'Start free',
  },
  {
    name: 'Dev',
    tagline: 'Early access — for daily AI coding with higher limits, custom dashboard guardrails, policy sync, and full history.',
    price: '$10',
    period: '/mo soon',
    features: [
      'Everything in Free',
      'Custom firewall rules',
      'Custom dashboard guardrails',
      'Policy sync',
      'Full security event history',
    ],
    cta: 'Request early access',
    featured: true,
  },
  {
    name: 'Team',
    tagline: 'Planned — shared policies, audit logs, team dashboard, and centralized control for agent-using engineering teams.',
    price: 'Planned',
    features: [
      'Everything in Dev',
      'Shared policy management',
      'Team usage dashboard',
      'Audit controls',
      'SSO / SAML',
    ],
    cta: 'Join waitlist',
  },
]

function Pricing() {
  const [earlyAccessOpen, setEarlyAccessOpen] = useState(false)
  const [earlyAccessPlan, setEarlyAccessPlan] = useState<'dev' | 'team'>('dev')
  const [planError, setPlanError] = useState(false)

  useEffect(() => {
    setPlanError(new URLSearchParams(window.location.search).get('error') === 'invalid_plan')
  }, [])

  function handlePlanCta(planSlug: string) {
    if (planSlug === 'free') {
      window.location.href = '/login'
      return
    }
    setEarlyAccessPlan(planSlug === 'team' ? 'team' : 'dev')
    setEarlyAccessOpen(true)
  }

  const plansWithHandlers = PLANS.map(plan => {
    const slug = plan.name === 'Dev' ? 'dev' : plan.name === 'Team' ? 'team' : 'free'
    return {
      ...plan,
      onCta: () => handlePlanCta(slug),
    }
  })

  return (
    <Section>
      <Container>
        <SectionTitle
          title="Free local firewall for AI coding agents. Dev opens soon."
        />

        {planError && (
          <p className="mx-auto mb-6 max-w-lg rounded-md border border-red-500/50 bg-red-500/10 px-4 py-2 text-center text-sm text-red-400">
            That plan is no longer available. Continue on Free or request early access below.
          </p>
        )}

        <PricingSection plans={plansWithHandlers} />

        <p className="mx-auto mt-6 max-w-lg text-center text-xs text-confire-border-strong">
          Confire is free during the public validation phase. Dev early access is
          opening for power users who want custom rules, higher limits, and full history.
        </p>

        <EarlyAccessForm
          open={earlyAccessOpen}
          onOpenChange={setEarlyAccessOpen}
          planInterest={earlyAccessPlan}
        />
      </Container>
    </Section>
  )
}

// ── FAQ ───────────────────────────────────────────────────────────────────────

const FAQ_ITEMS = [
  {
    q: 'Is Confire only a token optimizer?',
    a: 'No. Confire is a context and tool firewall. It reviews risky tool calls, redacts common secrets, sanitizes suspicious tool output, and trims context noise — in that order.',
  },
  {
    q: 'Does Confire replace Claude Code?',
    a: 'No. Confire runs around your existing agent workflow. Claude Code, Cursor, and VS Code remain your coding tools.',
  },
  {
    q: 'What is full firewall mode?',
    a: 'Full firewall mode means Confire can review or block tool calls before they run and replace/sanitize tool output before it enters context. This is available for Claude Code through hooks.',
  },
  {
    q: 'What is MCP gateway mode?',
    a: 'MCP gateway mode protects tools routed through Confire. It applies policies, sanitizes outputs, and filters MCP responses for Cursor and VS Code workflows.',
  },
  {
    q: 'Does Confire send my code to the cloud?',
    a: 'Tool inputs are evaluated locally. Secret redaction and prompt-injection sanitization run locally on your machine. Only structured telemetry events (risk level, action taken, session counts) are sent to the cloud — never raw tool output.',
  },
  {
    q: 'Can Confire prevent every unsafe agent action?',
    a: 'No. Confire is a guardrail layer, not a perfect security boundary. It reviews common risky actions and suspicious tool flows, but you should still review important operations.',
  },
  {
    q: 'What happens if Confire fails?',
    a: 'Confire is designed to fail safely. If the daemon is unavailable, your agent continues normally — tool calls pass through unmodified.',
  },
]

function FAQItem({ q, a }: { q: string; a: string }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="border-b border-confire-border">
      <button
        onClick={() => setOpen(o => !o)}
        className="flex w-full items-center justify-between gap-4 py-5 text-left text-sm font-medium text-confire-text transition-colors hover:text-confire-accent"
      >
        {q}
        <span
          className={cn(
            'shrink-0 text-confire-muted transition-transform duration-200',
            open && 'rotate-45',
          )}
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden>
            <path d="M7 1v12M1 7h12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </span>
      </button>
      {open && (
        <p className="pb-5 text-sm leading-relaxed text-confire-muted">{a}</p>
      )}
    </div>
  )
}

function FAQ() {
  return (
    <Section>
      <Container>
        <SectionTitle
          title="Questions developers ask before installing a firewall."
        />
        <WithCorners cols={1} rows={1}>
          <div className="border border-confire-border bg-confire-card px-8">
            {FAQ_ITEMS.map(item => (
              <FAQItem key={item.q} {...item} />
            ))}
          </div>
        </WithCorners>
      </Container>
    </Section>
  )
}

// ── Bottom CTA ────────────────────────────────────────────────────────────────

function BottomCTA() {
  return (
    <Section className="confire-dot-region px-4 pb-6 sm:px-8">
      <Container>
        <CTASection
          title="Give your AI coding agent a firewall."
          subtitle="Review risky tool calls before they run. Sanitize and filter tool output before it enters context. Start free with Claude Code, Cursor, or VS Code."
          primaryAction={
            <Button variant="white" asChild>
              <a href="/login">Start free, no card required</a>
            </Button>
          }
          secondaryAction={
            <Button variant="white-ghost" asChild>
              <a href="/docs">Read the docs</a>
            </Button>
          }
          marqueeItems={[
            { icon: <ShieldCheckIcon className={iconSm} weight="fill" />,  text: 'Claude Code full firewall' },
            { icon: <CpuIcon className={iconSm} weight="bold" />,          text: 'Cursor + VS Code MCP gateway' },
            { icon: <LockIcon className={iconSm} weight="bold" />,         text: 'Local policy evaluation' },
            { icon: <FingerprintIcon className={iconSm} weight="bold" />,  text: 'Sanitized before context reaches the model' },
            { icon: <TimerIcon className={iconSm} weight="bold" />,        text: 'Takes about 30 seconds to install' },
            { icon: <LightningIcon className={iconSm} weight="fill" />,    text: 'Works locally by default' },
          ]}
        />
        <p className="mt-4 text-center text-xs text-confire-muted">
          Takes about 30 seconds to install. Works locally by default.
        </p>
      </Container>
    </Section>
  )
}

// ── Footer columns ────────────────────────────────────────────────────────────

const FOOTER_COLUMNS = [
  {
    heading: 'Product',
    links: [
      { label: 'Overview',   href: '/'           },
      { label: 'Pricing',    href: '/#pricing'    },
      { label: 'Dashboard',  href: '/dashboard'   },
      { label: 'CLI',        href: '/docs/cli'    },
    ],
  },
  {
    heading: 'Developers',
    links: [
      { label: 'Docs',       href: '/docs'                                    },
      { label: 'GitHub',     href: 'https://github.com/confire-ai/confire'   },
      { label: 'Changelog',  href: '/changelog'                               },
      { label: 'Security',   href: '/security'                                },
    ],
  },
  {
    heading: 'Company',
    links: [
      { label: 'Blog',     href: '/blog'    },
      { label: 'Contact',  href: '/contact' },
      { label: 'Status',   href: 'https://status.confire.dev' },
    ],
  },
  {
    heading: 'Legal',
    links: [
      { label: 'Privacy',  href: '/privacy' },
      { label: 'Terms',    href: '/terms'   },
    ],
  },
]

// ── Main export ───────────────────────────────────────────────────────────────

export function MarketingHome() {
  return (
    <>
      <SiteNav
        items={[
          { label: 'Product', href: '#how-it-works' },
          { label: 'Pricing', href: '#pricing'       },
          { label: 'Docs',    href: '/docs'           },
        ]}
        ctaLabel="Start free"
        ctaHref="/login"
      />

      <Hero />
      <StatsBar />

      <div id="product">
        <Problem />
        <HowItWorks />
        <Capabilities />
        <Clients />
      </div>

      <Results />
      <Privacy />
      <Setup />

      <div id="pricing">
        <Pricing />
      </div>

      <FAQ />
      <BottomCTA />

      <SiteFooter columns={FOOTER_COLUMNS} secondaryColumns={[]} />
    </>
  )
}

export function DesignSystemShowcase() {
  return (
    <>
      <SiteNav />
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
