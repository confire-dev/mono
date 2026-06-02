import { useState } from 'react'
import { Badge, Button, LayerCard, Text } from '@cloudflare/kumo'
import { CheckIcon } from '@phosphor-icons/react'
import { AppShell } from '@/components/app/AppShell'

interface Props {
  user: { email: string; id: string }
  device: { id: string; name?: string; cliVersion?: string }
  callbackURL: string
}

const REQUESTED_ACCESS = [
  'Connect this device to your Confire account',
  'Send usage events for billing and dashboard',
  'Use remote optimizers if your plan allows',
]

export function CliAuthorizeCard({ user, device, callbackURL }: Props) {
  const [state, setState] = useState<'idle' | 'loading' | 'done' | 'error'>('idle')
  const [error, setError] = useState('')

  async function authorize() {
    setState('loading')
    setError('')
    try {
      const res = await fetch('/api/cli/authorize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ deviceId: device.id, deviceName: device.name, callbackURL }),
      })
      if (!res.ok) {
        const { error: e } = await res.json() as { error: string }
        throw new Error(e ?? 'Authorization failed')
      }
      const { apiKey, email, message } = await res.json() as {
        apiKey: string
        email: string
        message: string
      }
      setState('done')
      const params = new URLSearchParams({ api_key: apiKey, email, message })
      setTimeout(() => { window.location.href = `${callbackURL}?${params}` }, 800)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Authorization failed')
      setState('error')
    }
  }

  if (state === 'done') {
    return (
      <AppShell>
        <div className="flex flex-col items-center gap-4 text-center">
          <div className="flex size-12 items-center justify-center rounded-full bg-kumo-success-tint/70">
            <CheckIcon className="size-6 text-kumo-success" weight="bold" />
          </div>
          <Text variant="heading2" as="h1">You're connected</Text>
          <Text variant="secondary" size="sm">Returning to your terminal…</Text>
          <Badge variant="outline">{user.email}</Badge>
        </div>
      </AppShell>
    )
  }

  return (
    <AppShell>
      <div className="flex flex-col gap-6">
        <div className="flex flex-col items-center gap-2 text-center">
          <Text variant="heading3" as="span">Confire</Text>
          <Text variant="heading2" as="h1">Authorize Confire CLI</Text>
          <Text variant="secondary" size="sm">Grant CLI access to your account</Text>
        </div>

        <LayerCard className="flex flex-col gap-2 rounded-lg p-4">
          <div className="flex justify-between text-sm">
            <Text variant="secondary" size="sm" as="span">Signed in as</Text>
            <Text size="sm" as="span" bold>{user.email}</Text>
          </div>
          {device.cliVersion && (
            <div className="flex justify-between text-sm">
              <Text variant="secondary" size="sm" as="span">CLI version</Text>
              <Text size="sm" as="span" bold>{device.cliVersion}</Text>
            </div>
          )}
          {(device.name || device.id) && (
            <div className="flex justify-between text-sm">
              <Text variant="secondary" size="sm" as="span">Device</Text>
              <Text size="sm" as="span" bold>{device.name || device.id.slice(0, 8) + '…'}</Text>
            </div>
          )}
        </LayerCard>

        <div className="flex flex-col gap-2">
          <Text variant="secondary" size="xs" as="p" DANGEROUS_className="font-medium uppercase tracking-wide">
            Requested access
          </Text>
          <ul className="flex flex-col gap-1.5">
            {REQUESTED_ACCESS.map(item => (
              <li key={item} className="flex items-start gap-2">
                <CheckIcon className="mt-0.5 size-3.5 shrink-0 text-kumo-success" weight="bold" />
                <Text size="sm" as="span">{item}</Text>
              </li>
            ))}
          </ul>
        </div>

        <hr className="border-kumo-hairline" />

        {state === 'error' && (
          <Text variant="error" size="sm" as="p" DANGEROUS_className="text-center">
            {error}
          </Text>
        )}

        <Button onClick={authorize} variant="primary" loading={state === 'loading'}>
          Authorize CLI
        </Button>

        <Text variant="secondary" size="xs" as="p" DANGEROUS_className="text-center">
          This connects <strong className="text-kumo-default">{user.email}</strong> to the CLI on this device.
          Revoke access anytime from your dashboard.
        </Text>
      </div>
    </AppShell>
  )
}
