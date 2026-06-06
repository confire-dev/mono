import type { APIRoute } from 'astro'
import { createSupabaseServer } from '@/lib/supabase'
import { createClient } from '@supabase/supabase-js'

function sbAdmin() {
  return createClient(
    import.meta.env.PUBLIC_SUPABASE_URL,
    import.meta.env.SUPABASE_SERVICE_KEY,
    { auth: { autoRefreshToken: false, persistSession: false } }
  )
}

async function sha256(input: string): Promise<string> {
  const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(input))
  return Array.from(new Uint8Array(buf)).map(b => b.toString(16).padStart(2, '0')).join('')
}

function secureRandom(bytes: number): string {
  const arr = new Uint8Array(bytes)
  crypto.getRandomValues(arr)
  return Array.from(arr).map(b => b.toString(16).padStart(2, '0')).join('')
}

export const POST: APIRoute = async ({ request, cookies }) => {
  const supabase = createSupabaseServer(request, cookies)

  const { data: { user } } = await supabase.auth.getUser()
  if (!user) return Response.json({ error: 'unauthorized' }, { status: 401 })

  const { deviceId, deviceName, callbackURL } = await request.json() as {
    deviceId?: string; deviceName?: string; callbackURL?: string
  }

  if (!callbackURL?.match(/^http:\/\/(127\.0\.0\.1|localhost):\d+/)) {
    return Response.json({ error: 'invalid callback URL' }, { status: 400 })
  }

  const admin = sbAdmin()

  // Upsert profile
  const { error: profileError } = await admin
    .from('profiles')
    .upsert({ id: user.id, email: user.email }, { onConflict: 'id' })
  if (profileError) {
    return Response.json({ error: `profile error: ${profileError.message}` }, { status: 500 })
  }

  // Generate API key
  const rawKey = `cf_live_${secureRandom(32)}`
  const hash   = await sha256(rawKey)

  const { error: keyError } = await admin.from('api_keys').insert({
    user_id:    user.id,
    key_hash:   hash,
    key_prefix: rawKey.slice(0, 12),
    key_suffix: rawKey.slice(-4),
    device_id:  deviceName ?? deviceId ?? null,
  })
  if (keyError) {
    return Response.json({ error: `key error: ${keyError.message}` }, { status: 500 })
  }

  return Response.json({
    apiKey:  rawKey,
    email:   user.email,
    message: `✓ Logged in as ${user.email}`,
  })
}
