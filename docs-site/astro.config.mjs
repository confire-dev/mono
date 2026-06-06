import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'

export default defineConfig({
  site: process.env.SITE_URL ?? 'https://docs.confire.dev',
  base: process.env.DOCS_BASE ?? '/',
  integrations: [
    starlight({
      title: 'Confire',
      tagline: 'Context and tool firewall for AI coding agents.',
      logo: {
        light: './src/assets/logo-light.svg',
        dark:  './src/assets/logo-dark.svg',
        replacesTitle: false,
      },
      favicon: '/favicon.svg',
      social: [
        { icon: 'github',  label: 'GitHub',  href: 'https://github.com/confire-dev' },
      ],
      editLink: {
        baseUrl: 'https://github.com/confire-dev/mono/edit/main/docs-site/',
      },
      customCss: ['./src/styles/custom.css'],
      expressiveCode: {
        themes: ['github-dark'],
      },
      sidebar: [
        {
          label: 'Getting started',
          items: [
            { label: 'What is Confire?', link: '/' },
            { label: 'Install the CLI',   slug: 'getting-started/install' },
            { label: 'Connect your agents', slug: 'getting-started/connect' },
            { label: 'Test a policy',     slug: 'getting-started/first-policy' },
          ],
        },
        {
          label: 'How it works',
          items: [
            { label: 'Architecture',          slug: 'how-it-works/architecture' },
            { label: 'Tool Firewall',          slug: 'how-it-works/tool-firewall' },
            { label: 'MCP Firewall',           slug: 'how-it-works/mcp-firewall' },
            { label: 'Context Optimizer',      slug: 'how-it-works/context-optimizer' },
            { label: 'Secret Redaction',       slug: 'how-it-works/secret-redaction' },
            { label: 'Prompt-Injection Guard', slug: 'how-it-works/injection-guard' },
          ],
        },
        {
          label: 'Clients',
          items: [
            { label: 'Claude Code', slug: 'clients/claude-code' },
            { label: 'Cursor',      slug: 'clients/cursor' },
            { label: 'VS Code',     slug: 'clients/vscode' },
          ],
        },
        {
          label: 'Configuration',
          items: [
            { label: 'Policy rules',       slug: 'configuration/policy-rules' },
            { label: 'Custom guardrails',  slug: 'configuration/custom-guardrails' },
            { label: 'Remote policy sync', slug: 'configuration/remote-sync' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'CLI commands', slug: 'reference/cli' },
            { label: 'Plans',        slug: 'reference/plans' },
          ],
        },
      ],
    }),
  ],
})
