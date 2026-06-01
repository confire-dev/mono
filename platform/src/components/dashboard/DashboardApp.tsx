"use client"

import { Button, LayerCard, Text } from '@cloudflare/kumo'
import { useAuth } from '@/hooks/use-auth'

export function DashboardApp() {
  const { user, loading, signOut } = useAuth()

  if (loading) {
    return (
      <div className="flex min-h-svh items-center justify-center">
        <Text variant="secondary" size="sm">Loading…</Text>
      </div>
    )
  }

  if (!user) {
    window.location.href = '/login'
    return null
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6">
      <LayerCard className="flex max-w-md flex-col items-center gap-6 rounded-lg p-8 text-center">
        <Text variant="heading3" as="span">Confire</Text>
        <Text variant="body" as="p">Logged in as {user.email}</Text>
        <Button variant="primary" onClick={signOut}>Sign out</Button>
      </LayerCard>
    </div>
  )
}
