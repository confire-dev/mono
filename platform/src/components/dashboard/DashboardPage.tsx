// DashboardPage wraps DashboardShell + page content as a single client:only component.
// This sidesteps Astro's SSR pass for dashboard routes — all Supabase calls are client-only.
import { useEffect, useState } from 'react'
import { createBrowserClient } from '@/lib/supabase'
import { DashboardShell } from './DashboardShell'

interface User { id: string; email: string; name?: string }

interface Props {
  title: string
  currentPath: string
  children: React.ReactNode
}

export function DashboardPage({ title, currentPath, children }: Props) {
  const [user, setUser] = useState<User | null>(null)

  useEffect(() => {
    const supabase = createBrowserClient()
    supabase.auth.getUser().then(async ({ data: { user: u } }) => {
      if (!u) { window.location.href = `/login?next=${encodeURIComponent(currentPath)}`; return }
      const { data: profile } = await supabase
        .from('profiles').select('name').eq('id', u.id).maybeSingle()
      setUser({ id: u.id, email: u.email ?? '', name: profile?.name ?? undefined })
    })
  }, [])

  if (!user) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-sm text-muted-foreground">Loading…</p>
      </div>
    )
  }

  return (
    <DashboardShell title={title} currentPath={currentPath} user={user}>
      {children}
    </DashboardShell>
  )
}
