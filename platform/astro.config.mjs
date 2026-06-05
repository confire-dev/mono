import { defineConfig, sessionDrivers } from 'astro/config'
import react from '@astrojs/react'
import tailwindcss from '@tailwindcss/vite'
import cloudflare from '@astrojs/cloudflare'

export default defineConfig({
  integrations: [react()],
  vite: {
    plugins: [tailwindcss()],
    resolve: {
      dedupe: ['react', 'react-dom', 'react/jsx-runtime', 'react/jsx-dev-runtime'],
    },
  },
  output: 'server',
  // Prevent @astrojs/cloudflare v13+ from auto-enabling Cloudflare KV sessions.
  // Without an explicit driver, the adapter injects a SESSION KV binding into the
  // generated wrangler.json. That binding has no ID (auto-provisioning intent) and
  // fails when the Cloudflare Pages project has no SESSION namespace configured.
  // Using the memory driver satisfies the adapter check without requiring any binding.
  session: {
    driver: sessionDrivers.memory(),
  },
  // Disable Cloudflare Images runtime binding — no <Image> components are used in this project,
  // so the IMAGES binding injected by the adapter serves no purpose. Without this option the
  // adapter sets imageService to "cloudflare-binding" by default, which injects
  // { images: { binding: "IMAGES" } } into the generated wrangler.json. If the Cloudflare Pages
  // project doesn't have a Cloudflare Images subscription that binding is unresolvable and can
  // trigger a Worker exception (Error 1101). Using compile-time optimization avoids any runtime
  // binding requirement.
  adapter: cloudflare({ imageService: 'compile' }),
})
