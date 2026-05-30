// Self-contained: auth + shell + content in one client:only component.
// No Astro children needed — avoids the Astro SSR boundary issue.
import { useAuth } from '@/hooks/use-auth'
import { DashboardShell } from '../DashboardShell'
import { DashboardOverview } from '../DashboardOverview'

export function OverviewPage() {
  const { user, loading } = useAuth()

  if (loading) return <Loading />
  if (!user)   return <Redirect to="/dashboard" />

  return (
    <DashboardShell title="Dashboard" currentPath="/dashboard"
      user={{ email: user.email ?? '', name: user.user_metadata?.['name'] }}>
      <DashboardOverview userId={user.id} />
    </DashboardShell>
  )
}

// ── Stubs for remaining pages ─────────────────────────────────────────────
// Each page follows the same pattern: DashboardShell + page-specific content.

export function SessionsPage() {
  const { user, loading } = useAuth()
  if (loading) return <Loading />
  if (!user)   return <Redirect to="/dashboard/sessions" />
  return (
    <DashboardShell title="CLI Sessions" currentPath="/dashboard/sessions"
      user={{ email: user.email ?? '' }}>
      {/* TODO: <SessionList userId={user.id} /> */}
      {/* TODO: <ApiKeyList userId={user.id} /> */}
      <p className="text-sm text-muted-foreground">Coming soon.</p>
    </DashboardShell>
  )
}

export function BillingPage() {
  const { user, loading } = useAuth()
  if (loading) return <Loading />
  if (!user)   return <Redirect to="/dashboard/billing" />
  return (
    <DashboardShell title="Billing" currentPath="/dashboard/billing"
      user={{ email: user.email ?? '' }}>
      {/* TODO: <CurrentPlanCard userId={user.id} /> */}
      {/* TODO: <UsageHistory userId={user.id} /> */}
      {/* TODO: <UpgradeCard /> */}
      <p className="text-sm text-muted-foreground">Coming soon.</p>
    </DashboardShell>
  )
}

export function SettingsPage() {
  const { user, loading } = useAuth()
  if (loading) return <Loading />
  if (!user)   return <Redirect to="/dashboard/settings" />
  return (
    <DashboardShell title="Settings" currentPath="/dashboard/settings"
      user={{ email: user.email ?? '' }}>
      {/* TODO: <AccountSection userId={user.id} /> */}
      {/* TODO: <TelemetryToggle userId={user.id} /> */}
      {/* TODO: <PrivacySettings /> */}
      {/* TODO: <DangerZone /> */}
      <p className="text-sm text-muted-foreground">Coming soon.</p>
    </DashboardShell>
  )
}

// ── Shared micro-components ───────────────────────────────────────────────

function Loading() {
  return (
    <div className="flex h-screen items-center justify-center">
      <p className="text-sm text-muted-foreground">Loading…</p>
    </div>
  )
}

function Redirect({ to }: { to: string }) {
  if (typeof window !== 'undefined') {
    window.location.href = `/login?next=${encodeURIComponent(to)}`
  }
  return null
}
