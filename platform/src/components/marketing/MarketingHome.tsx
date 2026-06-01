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

const sampleCode = `const result = await confire.optimize({
  tool: 'Bash',
  output: rawShellOutput,
})
// 847 tokens → 124 tokens (85% saved)`

export function MarketingHome() {
  return (
    <>
      <SiteNav
        items={[
          { label: 'Product', href: '#product' },
          { label: 'Pricing', href: '/pricing' },
          { label: 'Docs', href: '/docs' },
        ]}
        ctaHref="/login"
      />

      <Section className="pt-16 pb-10">
        <Container>
          <div className="mx-auto max-w-3xl text-center">
            <Badge color="green" icon="⚡" className="mb-6">
              Universal AI agent optimizer
            </Badge>
            <H1 className="mb-6">
              Less noise in context.
              <br />
              <span className="text-confire-accent">Sharper agents.</span>
            </H1>
            <Body className="mx-auto mb-8 max-w-xl text-lg">
              Confire optimizes tool outputs before they reach the model — cheaper, faster,
              and more focused AI coding sessions.
            </Body>
            <div className="flex flex-wrap justify-center gap-3">
              <Button asChild>
                <a href="/login">Start for free</a>
              </Button>
              <Button variant="outline" asChild>
                <a href="/pricing">View pricing</a>
              </Button>
            </div>
          </div>
        </Container>
      </Section>

      <Section id="product">
        <Container>
          <SectionLabel number="01">Core patterns</SectionLabel>
          <SectionTitle
            title="Built for developers"
            subtitle="Corner-square cards, bento grids, and tabbed code — all themeable via CSS variables."
          />

          <ThreeCards
            cards={[
              <FeatureCard
                key="local"
                icon={<span className="text-2xl">⌘</span>}
                title="Local-first"
                description="Optimize on-device with zero latency. Cloud when you need the heavy hitters."
              />,
              <FeatureCard
                key="cloud"
                icon={<span className="text-2xl">☁</span>}
                title="Cloud optimizers"
                description="Figma, GitHub, Slack, and 13+ remote optimizers on Dev and Pro plans."
              />,
              <FeatureCard
                key="privacy"
                icon={<span className="text-2xl">🔒</span>}
                title="Privacy by default"
                description="Stats stay local. Telemetry is opt-in. Your code never trains our models."
              />,
            ]}
          />
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionLabel number="02">How it works</SectionLabel>
          <VerticalTabsCode
            tabs={[
              {
                title: 'Install the CLI',
                description: 'One command hooks into Claude Code, Cursor, and more.',
                code: `$ brew install confire\n$ confire setup\n$ confire login`,
              },
              {
                title: 'Optimize automatically',
                description: 'Every tool call is compressed before hitting the model context.',
                code: sampleCode,
              },
              {
                title: 'Track savings',
                description: 'See tokens saved locally, full history in your dashboard.',
                code: `$ confire stats\n\n  124,500 tokens saved (all time)\n  Full history → confire.dev/dashboard`,
              },
            ]}
          />
        </Container>
      </Section>

      <Section>
        <Container>
          <TwoCards
            card1={
              <div>
                <Badge color="accent" className="mb-4">
                  Before
                </Badge>
                <BodySm>12,400 tokens of raw bash output flooding your context window.</BodySm>
              </div>
            }
            card2={
              <div>
                <Badge color="green" className="mb-4">
                  After
                </Badge>
                <BodySm>890 tokens — same signal, 93% less noise. Your model stays sharp.</BodySm>
              </div>
            }
          />
        </Container>
      </Section>

      <Section>
        <Container>
          <QuoteCard
            quote="Confire cut our average session cost in half. The agent stopped drowning in tool output."
            author="Engineering team"
          />
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionLabel number="03">Pricing</SectionLabel>
          <PricingSection
            tabs={[
              { label: 'Monthly', icon: '◎' },
              { label: 'Annual', icon: '⬡' },
            ]}
            plans={[
              {
                name: 'Free',
                tagline: 'For individuals',
                price: '$0',
                period: '/mo',
                features: ['500 cloud opts', 'Generic optimizer', 'Local unlimited'],
                cta: 'Get started',
              },
              {
                name: 'Dev',
                tagline: 'For power users',
                price: '$10',
                period: '/mo',
                features: ['5,000 opts', '13 optimizers', 'Stats history'],
                cta: 'Upgrade',
                featured: true,
              },
              {
                name: 'Pro',
                tagline: 'For teams',
                price: '$20',
                period: '/mo',
                features: ['20,000 opts', 'Priority support', 'Team dashboard'],
                cta: 'Contact sales',
              },
            ]}
          />
        </Container>
      </Section>

      <Section>
        <Container>
          <CTASection
            title="Start saving tokens today"
            subtitle="Install Confire in under a minute. Free tier includes 500 cloud optimizations per month."
            primaryAction={
              <Button variant="white" asChild>
                <a href="/login">Install Confire</a>
              </Button>
            }
            secondaryAction={
              <Button variant="white-ghost" asChild>
                <a href="/docs">Read the docs</a>
              </Button>
            }
            marqueeItems={[
              { icon: '⚡', text: 'Works with Claude Code & Cursor' },
              { icon: '◈', text: 'Local stats, cloud dashboard' },
              { icon: '◎', text: '500 free optimizations / month' },
              { icon: '⬡', text: 'Top-up packs when you need more' },
            ]}
          />
        </Container>
      </Section>

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
