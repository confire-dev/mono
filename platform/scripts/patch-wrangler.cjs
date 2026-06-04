// Removes the ASSETS binding name from the Astro-generated dist/server/wrangler.json.
// Cloudflare Pages reserves "ASSETS" and injects it automatically at runtime.
// The @astrojs/cloudflare adapter unconditionally adds { binding: 'ASSETS' } in
// cloudflareConfigCustomizer() (packages/integrations/cloudflare/src/wrangler.ts:54-59)
// without checking for Pages projects (pages_build_output_dir). This is a known
// upstream bug — remove this script once the adapter fix is released and the
// package is updated.
const fs = require('fs')
const path = require('path')

const wranglerPath = path.join(__dirname, '..', 'dist', 'server', 'wrangler.json')

if (!fs.existsSync(wranglerPath)) {
  console.log('patch-wrangler: dist/server/wrangler.json not found, skipping')
  process.exit(0)
}

const config = JSON.parse(fs.readFileSync(wranglerPath, 'utf8'))
if (config.assets?.binding) {
  delete config.assets.binding
  console.log('patch-wrangler: removed reserved ASSETS binding name (kept directory for Pages)')
} else {
  console.log('patch-wrangler: no ASSETS binding found, nothing to patch')
}
fs.writeFileSync(wranglerPath, JSON.stringify(config, null, 2))
