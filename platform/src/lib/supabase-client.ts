// Module-level singleton for the Supabase browser client.
//
// WHY not useEffect:
//   useEffect creates the client asynchronously after first render. Components
//   that need Supabase would be `null` on the first render, causing a flicker.
//
// WHY useMemo instead:
//   useMemo creates the client synchronously in render, guaranteed to be stable
//   across re-renders, and only runs once per component instance. No flicker.
//   Safe because all components using this are `client:only="react"` — they
//   never run on the server, so `import.meta.env` is always populated.
//
// USAGE:
//   const supabase = useMemo(() => getSupabaseClient(), [])
//
// Or import the singleton directly in utility hooks:
//   import { supabaseClient } from '@/lib/supabase-client'

import { createBrowserClient } from '@/lib/supabase'
import type { SupabaseClient } from '@supabase/supabase-js'

// Module singleton — created once when this module is first imported in the browser.
// Since this module is only imported by `client:only` components, it is always
// safe to call createBrowserClient() at module level here.
let _singleton: SupabaseClient | null = null

export function getSupabaseClient(): SupabaseClient {
  if (!_singleton) {
    _singleton = createBrowserClient()
  }
  return _singleton
}

// Re-export for convenience
export { getSupabaseClient as supabaseClient }
