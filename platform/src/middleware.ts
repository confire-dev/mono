import { defineMiddleware } from 'astro:middleware'

// Auth middleware — active only when Supabase env vars are present.
// During the scaffold phase (output: 'static', no adapter) this is a no-op.
// Enable fully when deploying with the Cloudflare adapter and real env vars.

const PROTECTED = ['/dashboard']

export const onRequest = defineMiddleware(async ({ url, cookies, redirect, locals }, next) => {
  const supabaseURL = import.meta.env.PUBLIC_SUPABASE_URL

  // No-op during static build / dev without Supabase configured
  if (!supabaseURL) return next()

  const isProtected = PROTECTED.some(p => url.pathname.startsWith(p))
  if (!isProtected) return next()

  // Lazy import so the Supabase client is only created when env vars exist
  const { createSupabaseServer } = await import('@/lib/supabase')
  const supabase = createSupabaseServer(cookies)
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    const next = encodeURIComponent(url.pathname + url.search)
    return redirect(`/login?next=${next}`)
  }

  locals.user    = user
  locals.supabase = supabase
  return next()
})
