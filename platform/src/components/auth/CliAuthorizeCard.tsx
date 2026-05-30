import { useState } from 'react'
import { Check, Zap, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

interface Props {
  user: { email: string; id: string }
  device: { id: string; cliVersion?: string }
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
        body: JSON.stringify({ deviceId: device.id, callbackURL }),
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
      // Redirect browser to the CLI's local callback server with the key
      const params = new URLSearchParams({ api_key: apiKey, email, message })
      setTimeout(() => { window.location.href = `${callbackURL}?${params}` }, 800)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Authorization failed')
      setState('error')
    }
  }

  if (state === 'done') {
    return (
      <div className="flex flex-col items-center gap-4 text-center">
        <div className="flex size-12 items-center justify-center rounded-full bg-green-500/10">
          <Check className="size-6 text-green-500" />
        </div>
        <h1 className="text-xl font-bold">You're connected</h1>
        <p className="text-sm text-muted-foreground">Returning to your terminal…</p>
        <Badge variant="outline" className="text-xs">{user.email}</Badge>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col items-center gap-2 text-center">
        <div className="flex items-center gap-3">
          <div className="flex size-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Zap className="size-5" />
          </div>
          <Terminal className="size-5 text-muted-foreground" />
        </div>
        <h1 className="text-xl font-bold">Authorize Confire CLI</h1>
        <p className="text-sm text-muted-foreground">Grant CLI access to your account</p>
      </div>

      <div className="rounded-lg border p-4 text-sm flex flex-col gap-2">
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

      <div className="flex flex-col gap-2">
        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Requested access</p>
        <ul className="flex flex-col gap-1.5">
          {REQUESTED_ACCESS.map(item => (
            <li key={item} className="flex items-start gap-2 text-sm">
              <Check className="mt-0.5 size-3.5 shrink-0 text-green-500" />
              {item}
            </li>
          ))}
        </ul>
      </div>

      <Separator />

      {state === 'error' && <p className="text-center text-sm text-destructive">{error}</p>}

      <Button onClick={authorize} disabled={state === 'loading'}>
        {state === 'loading' ? 'Authorizing…' : 'Authorize CLI'}
      </Button>

      <p className="text-center text-xs text-muted-foreground">
        This connects <strong>{user.email}</strong> to the CLI on this device.
        Revoke access anytime from your dashboard.
      </p>
    </div>
  )
}
