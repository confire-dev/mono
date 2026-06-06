import { createBrowserClient as ssrBrowserClient, createServerClient, parseCookieHeader } from '@supabase/ssr'
import type { AstroCookies } from 'astro'

// Lazy getters so the module can be imported at build time without env vars.
function url()  { return import.meta.env.PUBLIC_SUPABASE_URL  ?? '' }
function anon() { return import.meta.env.PUBLIC_SUPABASE_ANON_KEY ?? '' }

// Browser client — uses @supabase/ssr so the session is synced to cookies,
// making it visible to the server-side middleware after OAuth / OTP callback.
// Returns null when Supabase env vars are not configured (safe for SSR).
export function createBrowserClient() {
  const u = url()
  const a = anon()
  if (!u || !a) return null
  return ssrBrowserClient(u, a)
}

export type SupabaseBrowserClient = NonNullable<ReturnType<typeof createBrowserClient>>

// Server client — for Astro pages and middleware.
export function createSupabaseServer(request: Request, cookies: AstroCookies) {
  return createServerClient(url(), anon(), {
    cookies: {
      getAll() {
        return parseCookieHeader(request.headers.get('cookie') ?? '')
          .filter((c): c is { name: string; value: string } => c.value !== undefined)
      },
      setAll(cookiesToSet) {
        cookiesToSet.forEach(({ name, value, options }) => {
          cookies.set(name, value, options)
        })
      },
    },
  })
}
