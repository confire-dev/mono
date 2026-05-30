// DashboardPage: auth-aware wrapper around DashboardShell.
// Uses useAuth() — the correct client:only pattern (useMemo, not useEffect).
import { useAuth } from '@/hooks/use-auth'
import { DashboardShell } from './DashboardShell'

interface Props {
  title: string
  currentPath: string
  children: React.ReactNode
}

export function DashboardPage({ title, currentPath, children }: Props) {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-sm text-muted-foreground">Loading…</p>
      </div>
    )
  }

  if (!user) {
    // Redirect client-side — middleware handles server-side redirect in SSR mode.
    if (typeof window !== 'undefined') {
      window.location.href = `/login?next=${encodeURIComponent(currentPath)}`
    }
    return null
  }

  return (
    <DashboardShell
      title={title}
      currentPath={currentPath}
      user={{ email: user.email ?? '', name: user.user_metadata?.['name'] }}
    >
      {children}
    </DashboardShell>
  )
}
