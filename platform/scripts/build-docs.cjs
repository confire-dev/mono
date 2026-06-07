#!/usr/bin/env node
// Builds docs-site with base=/docs and copies output into public/docs/
// so Astro serves them as static assets at /docs/* on the platform domain.
const { execSync } = require('child_process')
const path = require('path')
const fs = require('fs')

const root     = path.resolve(__dirname, '../../')
const docsDir  = path.join(root, 'docs-site')
const docsDist = path.join(docsDir, 'dist')
const dest     = path.join(__dirname, '../public/docs')

console.log('Building docs-site...')
execSync('npm install', { cwd: docsDir, stdio: 'inherit' })
execSync('npm run build', {
  cwd: docsDir,
  stdio: 'inherit',
  env: {
    ...process.env,
    DOCS_BASE: '/docs',
    SITE_URL: process.env.DOCS_SITE_URL ?? 'https://confire.dev',
  },
})

if (fs.existsSync(dest)) fs.rmSync(dest, { recursive: true })
fs.cpSync(docsDist, dest, { recursive: true })
console.log('Docs copied to public/docs/')
