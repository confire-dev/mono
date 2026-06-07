import { useState, useEffect, useMemo } from 'react'
import type { User, Session } from '@supabase/supabase-js'
import { createBrowserClient } from '@/lib/supabase'

export interface AuthState {
  user:    User | null
  session: Session | null
  loading: boolean
}

export function useAuth(): AuthState & { signOut: () => Promise<void> } {
  const supabase = useMemo(() => createBrowserClient(), [])
  const [user,    setUser]    = useState<User | null>(null)
  const [session, setSession] = useState<Session | null>(null)
  // Start in loading=false when Supabase is unconfigured — no async work to await.
  const [loading, setLoading] = useState(supabase !== null)

  useEffect(() => {
    if (!supabase) return

    // 1. Get the initial session synchronously from the local cookie/storage.
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(session)
      setUser(session?.user ?? null)
      setLoading(false)
    })

    // 2. Subscribe to future changes (token refresh, sign-out, etc.)
    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      setSession(session)
      setUser(session?.user ?? null)
      setLoading(false)
    })

    return () => subscription.unsubscribe()
  }, [supabase])

  const signOut = async () => {
    await supabase?.auth.signOut()
    window.location.href = '/'
  }

  return { user, session, loading, signOut }
}
