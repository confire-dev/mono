import type { APIRoute } from 'astro'
import { createSupabaseServer } from '@/lib/supabase'

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const VALID_CLIENTS = ['Claude Code', 'Cursor', 'VS Code', 'Other']

export const POST: APIRoute = async ({ request, cookies }) => {
  const body = await request.json().catch(() => null) as {
    email?: string
    client?: string
    useCase?: string
    name?: string
    company?: string
    planInterest?: string
    websiteUrl?: string
  } | null

  if (!body) return Response.json({ error: 'invalid request body' }, { status: 400 })

  // Honeypot — bots fill hidden fields, humans don't. Pretend success either way.
  if (body.websiteUrl) return Response.json({ ok: true })

  const email = (body.email ?? '').trim().toLowerCase()
  if (!EMAIL_RE.test(email)) {
    return Response.json({ error: 'enter a valid email address' }, { status: 400 })
  }

  const client = VALID_CLIENTS.includes(body.client ?? '') ? body.client : null
  const planInterest = body.planInterest === 'team' ? 'team' : 'dev'

  const supabase = createSupabaseServer(request, cookies)
  const { error } = await supabase.from('early_access_requests').insert({
    email,
    client,
    use_case: (body.useCase ?? '').trim() || null,
    name: (body.name ?? '').trim() || null,
    company: (body.company ?? '').trim() || null,
    plan_interest: planInterest,
    source: 'pricing_page',
  })

  // Unique violation on (email, plan_interest) — already on the list, treat as success.
  if (error && error.code !== '23505') {
    return Response.json({ error: 'something went wrong — try again' }, { status: 500 })
  }

  return Response.json({ ok: true })
}
