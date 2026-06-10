import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'
import starlightThemeRapide from 'starlight-theme-rapide'

export default defineConfig({
  site: process.env.SITE_URL ?? 'https://docs.confire.dev',
  base: process.env.DOCS_BASE ?? '/',
  integrations: [
    starlight({
      title: 'Confire',
      tagline: 'Local tool and context firewall for AI coding agents.',
      logo: {
        src: './src/assets/logo-light.svg',
        replacesTitle: false,
        alt: 'Confire flame',
      },
      favicon: '/favicon.svg',
      head: [
        { tag: 'script', content: "document.documentElement.dataset.theme = 'dark'" },
      ],
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/confire-dev' },
      ],
      editLink: {
        baseUrl: 'https://github.com/confire-dev/mono/edit/main/docs-site/',
      },
      plugins: [starlightThemeRapide()],
      customCss: ['./src/styles/custom.css'],
      expressiveCode: {
        themes: ['github-dark'],
      },
      sidebar: [
        {
          label: 'Getting started',
          items: [
            { label: 'Install',              slug: 'getting-started/install' },
            { label: 'Connect your agent',   slug: 'getting-started/connect' },
            { label: 'Verify your setup',    slug: 'getting-started/verify' },
            { label: 'Uninstall',            slug: 'getting-started/uninstall' },
          ],
        },
        {
          label: 'Core concepts',
          items: [
            { label: 'What is Confire?',     slug: 'core-concepts/what-is-confire' },
            { label: 'Hook phases',          slug: 'core-concepts/hook-phases' },
            { label: 'Client support',       slug: 'core-concepts/client-support-modes' },
          ],
        },
        {
          label: 'Tool Result Firewall',
          items: [
            { label: 'Overview',           slug: 'context-firewall/overview' },
            { label: 'Tool result checks', slug: 'context-firewall/tool-result-checks' },
            { label: 'Secret warnings',    slug: 'context-firewall/secret-warnings' },
            { label: 'Injection guard',    slug: 'context-firewall/injection-guard' },
            { label: 'MCP security',       slug: 'context-firewall/mcp-security' },
            { label: 'Provenance',         slug: 'context-firewall/provenance' },
          ],
        },
        {
          label: 'Tool Firewall',
          items: [
            { label: 'Overview',             slug: 'tool-firewall/overview' },
            { label: 'Built-in rules',       slug: 'tool-firewall/built-in-rules' },
            { label: 'Policy modes',         slug: 'tool-firewall/policy-modes' },
            { label: 'MCP risk classifier',  slug: 'tool-firewall/mcp-risk-classifier' },
            { label: 'Custom guardrails',    slug: 'tool-firewall/custom-rules' },
            { label: 'Bypass and approvals', slug: 'tool-firewall/bypass-and-approvals' },
          ],
        },
        {
          label: 'Configuration',
          items: [
            { label: 'Config file',                    slug: 'configuration/config-file' },
            { label: 'Client settings',                slug: 'configuration/per-agent-settings' },
            { label: 'Privacy',                        slug: 'configuration/privacy-and-cloud-optimization' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'CLI commands',     slug: 'reference/cli' },
            { label: 'Plans and limits', slug: 'reference/plans-and-limits' },
          ],
        },
      ],
    }),
  ],
})
