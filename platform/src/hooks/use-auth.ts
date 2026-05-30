// useAuth — the correct pattern for Supabase Auth in client:only React islands.
//
// Creates the client via useMemo (synchronous, stable, no first-render null).
// Subscribes to onAuthStateChange so the session stays fresh on token refresh.
// Cleans up the subscription on unmount.

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
  const [loading, setLoading] = useState(true)

  useEffect(() => {
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
    await supabase.auth.signOut()
    window.location.href = '/'
  }

  return { user, session, loading, signOut }
}
