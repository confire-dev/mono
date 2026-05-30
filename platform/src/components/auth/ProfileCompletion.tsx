import { useState } from 'react'
import { createBrowserClient } from '@/lib/supabase'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { IconBolt } from '@tabler/icons-react'

interface Props {
  email: string
  userId: string
  nextUrl?: string
}

export function ProfileCompletion({ email, userId, nextUrl = '/dashboard' }: Props) {
  const [name,    setName]    = useState('')
  const [loading, setLoading] = useState(false)
  const [error,   setError]   = useState('')
  const supabase = createBrowserClient()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setLoading(true)
    setError('')

    const { error } = await supabase
      .from('profiles')
      .upsert({ id: userId, email, name: name.trim() })

    if (error) {
      setError(error.message)
      setLoading(false)
    } else {
      window.location.href = nextUrl
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col items-center gap-2 text-center">
        <IconBolt className="size-8" />
        <h1 className="text-xl font-bold">Welcome to Confire</h1>
        <p className="text-sm text-muted-foreground">One quick step before we start.</p>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            type="text"
            placeholder="Your name"
            value={name}
            onChange={e => setName(e.target.value)}
            autoFocus
            required
          />
        </div>

        {/* Read-only email */}
        <p className="text-center text-xs text-muted-foreground">
          Signed in as <span className="font-medium">{email}</span>
        </p>

        {error && <p className="text-sm text-destructive text-center">{error}</p>}

        <Button type="submit" disabled={loading || !name.trim()}>
          {loading ? 'Saving…' : 'Continue'}
        </Button>
      </form>
    </div>
  )
}
