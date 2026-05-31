import { useState } from "react"
import { Zap } from "lucide-react"
import { createBrowserClient } from "@/lib/supabase"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface Props {
  email: string
  userId: string
  nextUrl?: string
}

export function ProfileCompletion({ email, userId, nextUrl = "/dashboard" }: Props) {
  const [name,    setName]    = useState("")
  const [loading, setLoading] = useState(false)
  const [error,   setError]   = useState("")
  const supabase = createBrowserClient()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setLoading(true)
    setError("")
    const { error } = await supabase
      .from("profiles")
      .update({ name: name.trim() })
      .eq("id", userId)
      .is("name", null)
    if (error) { setError(error.message); setLoading(false) }
    else window.location.href = nextUrl
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col items-center gap-2 text-center">
        <div className="flex size-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <Zap className="size-5" />
        </div>
        <h1 className="text-xl font-bold">Welcome to Confire</h1>
        <p className="text-sm text-muted-foreground">One quick step before we start.</p>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="grid gap-2">
          <Label htmlFor="name">Your name</Label>
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
        <p className="text-center text-xs text-muted-foreground">
          Signed in as <span className="font-medium">{email}</span>
        </p>
        {error && <p className="text-center text-sm text-destructive">{error}</p>}
        <Button type="submit" className="w-full" disabled={loading || !name.trim()}>
          {loading ? "Saving…" : "Continue"}
        </Button>
      </form>
    </div>
  )
}
