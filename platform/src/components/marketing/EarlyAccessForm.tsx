import { useEffect, useState } from 'react'
import {
  Button, Dialog, DialogClose, DialogDescription, DialogTitle,
  Field, Input, Select, Text, Textarea,
} from '@cloudflare/kumo'

type Step = 'idle' | 'submitting' | 'done'

const CLIENTS = ['Claude Code', 'Cursor', 'VS Code', 'Other']

const COPY = {
  dev: {
    title: 'Get Dev early access',
    description: 'Dev is opening soon for power users who want custom guardrails, higher remote limits, and full history.',
  },
  team: {
    title: 'Join the Team waitlist',
    description: 'Team is planned for engineering teams that need shared policies, audit logs, and centralized control.',
  },
}

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  planInterest?: 'dev' | 'team'
}

export function EarlyAccessForm({ open, onOpenChange, planInterest = 'dev' }: Props) {
  const [step, setStep] = useState<Step>('idle')
  const [email, setEmail] = useState('')
  const [client, setClient] = useState('')
  const [useCase, setUseCase] = useState('')
  const [websiteUrl, setWebsiteUrl] = useState('')
  const [error, setError] = useState('')

  const isLoading = step === 'submitting'
  const copy = COPY[planInterest]

  // Reset form state whenever the dialog closes.
  useEffect(() => {
    if (!open) {
      setStep('idle')
      setEmail('')
      setClient('')
      setUseCase('')
      setError('')
    }
  }, [open])

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    if (!email) return
    setStep('submitting')
    setError('')

    const res = await fetch('/api/early-access', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, client, useCase, planInterest, websiteUrl }),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => null) as { error?: string } | null
      setError(data?.error ?? 'something went wrong — try again')
      setStep('idle')
      return
    }

    setStep('done')
  }

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog className="p-8 w-full max-w-md">
        {step === 'done' ? (
          <div className="flex flex-col gap-4">
            <DialogTitle>You're on the list</DialogTitle>
            <DialogDescription>
              We'll email you when Dev opens.
            </DialogDescription>
            <DialogClose render={<Button variant="primary" className="w-full">Close</Button>} />
          </div>
        ) : (
          <form onSubmit={submit} className="flex flex-col gap-4">
            <div>
              <DialogTitle>{copy.title}</DialogTitle>
              <DialogDescription>{copy.description}</DialogDescription>
            </div>

            <Field label="Email">
              <Input
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
                autoComplete="email"
                autoFocus
              />
            </Field>

            <Select
              label="Which client do you use?"
              placeholder="Select a client"
              value={client || undefined}
              onValueChange={v => setClient(String(v))}
            >
              {CLIENTS.map(c => (
                <Select.Option key={c} value={c}>{c}</Select.Option>
              ))}
            </Select>

            <Field label="What are you building?">
              <Textarea
                placeholder="Optional — tell us what you're working on"
                value={useCase}
                onChange={e => setUseCase(e.target.value)}
                rows={3}
              />
            </Field>

            {/* Honeypot — hidden from real users, bots tend to fill every field. */}
            <input
              type="text"
              name="website_url"
              value={websiteUrl}
              onChange={e => setWebsiteUrl(e.target.value)}
              tabIndex={-1}
              autoComplete="off"
              style={{ position: 'absolute', left: '-9999px', width: 1, height: 1, opacity: 0 }}
              aria-hidden="true"
            />

            {error && (
              <Text variant="error" size="sm" as="p">{error}</Text>
            )}

            <Button type="submit" variant="primary" className="w-full" loading={isLoading}>
              Request access
            </Button>
          </form>
        )}
      </Dialog>
    </Dialog.Root>
  )
}
