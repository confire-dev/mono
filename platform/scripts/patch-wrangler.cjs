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
delete config.assets
fs.writeFileSync(wranglerPath, JSON.stringify(config, null, 2))
console.log('patch-wrangler: removed reserved ASSETS binding for Pages compatibility')
