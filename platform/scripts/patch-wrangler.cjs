// Removes the ASSETS binding from the Astro-generated dist/server/wrangler.json.
// Cloudflare Pages reserves the name "ASSETS" and provides it automatically at
// runtime — declaring it explicitly causes a deploy error.
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
