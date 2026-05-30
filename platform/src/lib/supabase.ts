import { createClient } from '@supabase/supabase-js'
import { createServerClient, parseCookieHeader } from '@supabase/ssr'
import type { AstroCookies } from 'astro'

// Lazy getters so the module can be imported at build time without env vars.
function url()  { return import.meta.env.PUBLIC_SUPABASE_URL  ?? '' }
function anon() { return import.meta.env.PUBLIC_SUPABASE_ANON_KEY ?? '' }

// Browser client — for React islands running in the browser.
// Call once per island, not at module level.
export function createBrowserClient() {
  return createClient(url(), anon())
}

// Server client — for Astro pages and middleware.
export function createSupabaseServer(cookies: AstroCookies) {
  return createServerClient(url(), anon(), {
    cookies: {
      getAll() {
        return parseCookieHeader(cookies.toString())
      },
      setAll(cookiesToSet) {
        cookiesToSet.forEach(({ name, value, options }) => {
          cookies.set(name, value, options)
        })
      },
    },
  })
}
