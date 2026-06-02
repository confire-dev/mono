import { useState } from 'react'
import { Button, Field, Input, Text } from '@cloudflare/kumo'
import { createBrowserClient } from '@/lib/supabase'
import { AppShell } from '@/components/app/AppShell'

interface Props {
  email: string
  userId: string
  nextUrl?: string
}

export function ProfileCompletion({ email, userId, nextUrl = '/dashboard' }: Props) {
  const [name, setName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const supabase = createBrowserClient()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setLoading(true)
    setError('')
    const { error: updateError } = await supabase
      .from('profiles')
      .update({ name: name.trim() })
      .eq('id', userId)
    if (updateError) { setError(updateError.message); setLoading(false) }
    else window.location.href = nextUrl
  }

  return (
    <AppShell>
      <div className="flex flex-col gap-6">
        <div className="flex flex-col items-center gap-2 text-center">
          <Text variant="heading3" as="span">Confire</Text>
          <Text variant="heading2" as="h1">Welcome to Confire</Text>
          <Text variant="secondary" size="sm">One quick step before we start.</Text>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <Field label="Your name">
            <Input
              type="text"
              placeholder="Your name"
              value={name}
              onChange={e => setName(e.target.value)}
              autoFocus
              required
            />
          </Field>
          <Text variant="secondary" size="xs" as="p" DANGEROUS_className="text-center">
            Signed in as <span className="font-medium text-kumo-default">{email}</span>
          </Text>
          {error && (
            <Text variant="error" size="sm" as="p" DANGEROUS_className="text-center">
              {error}
            </Text>
          )}
          <Button type="submit" variant="primary" className="w-full" loading={loading} disabled={!name.trim()}>
            Continue
          </Button>
        </form>
      </div>
    </AppShell>
  )
}
