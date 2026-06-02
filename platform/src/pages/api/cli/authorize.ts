import type { APIRoute } from 'astro'
import { createSupabaseServer } from '@/lib/supabase'

export const POST: APIRoute = async ({ request, cookies }) => {
  const supabase = createSupabaseServer(request, cookies)

  // Verify the user is authenticated
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return Response.json({ error: 'unauthorized' }, { status: 401 })
  }

  // Get the session JWT to pass to the worker
  const { data: { session } } = await supabase.auth.getSession()
  if (!session?.access_token) {
    return Response.json({ error: 'no active session' }, { status: 401 })
  }

  const { deviceId, deviceName, callbackURL } = await request.json() as { deviceId?: string; deviceName?: string; callbackURL?: string }

  // Only allow callbacks to localhost (CLI callback server)
  if (!callbackURL?.match(/^http:\/\/(127\.0\.0\.1|localhost):\d+/)) {
    return Response.json({ error: 'invalid callback URL' }, { status: 400 })
  }

  // Call the worker to generate the API key
  const workerURL = import.meta.env.PUBLIC_WORKER_URL ?? 'http://localhost:8787'
  const res = await fetch(`${workerURL}/api/keys/generate`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${session.access_token}`,
      ...(deviceId   ? { 'X-Confire-Device':      deviceId   } : {}),
      ...(deviceName ? { 'X-Confire-Device-Name':  deviceName } : {}),
    },
  })

  if (!res.ok) {
    const text = await res.text()
    return Response.json({ error: `worker error: ${text}` }, { status: 502 })
  }

  const data = await res.json() as {
    apiKey: string
    message: string
    user: { email: string }
  }

  return Response.json({
    apiKey:   data.apiKey,
    email:    data.user.email,
    message:  data.message,
  })
}
