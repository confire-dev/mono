import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { IconCheck, IconBolt, IconTerminal2 } from '@tabler/icons-react'

interface Props {
  user: { email: string; id: string }
  device: { id: string; cliVersion?: string }
  callbackURL?: string
  codeChallenge?: string
}

const REQUESTED_ACCESS = [
  'Connect this device to your account',
  'Send required usage events for billing and dashboard',
  'Use remote optimizers if your plan allows',
]

export function CliAuthorizeCard({ user, device, callbackURL, codeChallenge }: Props) {
  const [state, setState] = useState<'idle' | 'loading' | 'done' | 'error'>('idle')
  const [error, setError] = useState('')

  async function authorize() {
    setState('loading')
    setError('')
    try {
      const res = await fetch('/api/cli/authorize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ deviceId: device.id, codeChallenge }),
      })
      if (!res.ok) throw new Error(await res.text())
      setState('done')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Authorization failed')
      setState('error')
    }
  }

  if (state === 'done') {
    return (
      <div className="flex flex-col items-center gap-4 text-center">
        <div className="flex size-12 items-center justify-center rounded-full bg-green-500/10">
          <IconCheck className="size-6 text-green-500" />
        </div>
        <h1 className="text-xl font-bold">You're connected</h1>
        <p className="text-sm text-muted-foreground">You can return to your terminal.</p>
        <Badge variant="outline" className="text-xs">{user.email}</Badge>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex flex-col items-center gap-2 text-center">
        <div className="flex items-center gap-2">
          <IconBolt className="size-6" />
          <IconTerminal2 className="size-5 text-muted-foreground" />
        </div>
        <h1 className="text-xl font-bold">Authorize Confire CLI</h1>
      </div>

      {/* Device info */}
      <div className="rounded-lg border p-4 space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Signed in as</span>
          <span className="font-medium">{user.email}</span>
        </div>
        {device.cliVersion && (
          <div className="flex justify-between">
            <span className="text-muted-foreground">CLI version</span>
            <span className="font-medium">{device.cliVersion}</span>
          </div>
        )}
        {device.id && (
          <div className="flex justify-between">
            <span className="text-muted-foreground">Device ID</span>
            <span className="font-mono text-xs text-muted-foreground">{device.id.slice(0, 8)}…</span>
          </div>
        )}
      </div>

      {/* Requested access */}
      <div className="space-y-2">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          Requested access
        </p>
        <ul className="space-y-1.5">
          {REQUESTED_ACCESS.map(item => (
            <li key={item} className="flex items-start gap-2 text-sm">
              <IconCheck className="size-3.5 mt-0.5 shrink-0 text-green-500" />
              {item}
            </li>
          ))}
        </ul>
      </div>

      <Separator />

      {error && <p className="text-sm text-destructive text-center">{error}</p>}

      <Button onClick={authorize} disabled={state === 'loading'}>
        {state === 'loading' ? 'Authorizing…' : 'Authorize CLI'}
      </Button>

      <p className="text-center text-xs text-muted-foreground">
        This connects <strong>{user.email}</strong> to the CLI on this device.
      </p>
    </div>
  )
}
