import type { APIRoute } from 'astro'
import { createSupabaseServer } from '@/lib/supabase'

// cloudflare:workers is external — available in CF runtime, undefined elsewhere.
// @ts-ignore
import { env as cfEnv } from 'cloudflare:workers'

export const POST: APIRoute = async ({ request, cookies }) => {
  const supabase = createSupabaseServer(request, cookies)

  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return Response.json({ error: 'unauthorized' }, { status: 401 })
  }

  const { data: { session } } = await supabase.auth.getSession()
  if (!session?.access_token) {
    return Response.json({ error: 'no active session' }, { status: 401 })
  }

  const { deviceId, deviceName, callbackURL } = await request.json() as { deviceId?: string; deviceName?: string; callbackURL?: string }

  if (!callbackURL?.match(/^http:\/\/(127\.0\.0\.1|localhost):\d+/)) {
    return Response.json({ error: 'invalid callback URL' }, { status: 400 })
  }

  const headers: Record<string, string> = {
    Authorization: `Bearer ${session.access_token}`,
    ...(deviceId   ? { 'X-Confire-Device':      deviceId   } : {}),
    ...(deviceName ? { 'X-Confire-Device-Name':  deviceName } : {}),
  }

  // Use service binding when available (avoids Cloudflare same-zone HTTP
  // restrictions that return error 1003). Falls back to HTTP for local dev.
  const workerAPI = (cfEnv as any)?.WORKER_API as { fetch(req: Request): Promise<Response> } | undefined
  const res = workerAPI
    ? await workerAPI.fetch(new Request('http://worker/api/keys/generate', { method: 'POST', headers }))
    : await fetch(`${import.meta.env.WORKER_URL ?? 'http://localhost:8787'}/api/keys/generate`, { method: 'POST', headers })

  if (!res.ok) {
    const text = await res.text()
    return Response.json({ error: `worker error (${res.status}): ${text}` }, { status: 502 })
  }

  const data = await res.json() as { apiKey: string; message: string; user: { email: string } }

  return Response.json({
    apiKey:   data.apiKey,
    email:    data.user.email,
    message:  data.message,
  })
}
