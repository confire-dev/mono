'use client'

import { useState } from 'react'
import { createBrowserClient } from '@/lib/supabase'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { IconBrandGithub, IconBrandGoogle, IconBolt } from '@tabler/icons-react'

// States for the unified login/register flow
type Step =
  | 'idle'          // show social + email button
  | 'email-sent'    // email was sent — show OTP input + "or click link" message
  | 'submitting'    // awaiting Supabase response
  | 'error'         // show error, allow retry

interface Props {
  className?: string
  nextUrl?: string
}

export function AuthForm({ className, nextUrl = '/dashboard' }: Props) {
  const [step,  setStep]  = useState<Step>('idle')
  const [email, setEmail] = useState('')
  const [otp,   setOtp]   = useState('')
  const [error, setError] = useState('')
  const supabase = createBrowserClient()

  const redirectTo = typeof window !== 'undefined'
    ? `${window.location.origin}/auth/callback?next=${encodeURIComponent(nextUrl)}`
    : `/auth/callback?next=${encodeURIComponent(nextUrl)}`

  // ── OAuth (GitHub / Google) ────────────────────────────────────────────
  async function signInWithProvider(provider: 'github' | 'google') {
    setError('')
    const { error } = await supabase.auth.signInWithOAuth({
      provider,
      options: { redirectTo },
    })
    if (error) setError(error.message)
  }

  // ── Email: send magic link + OTP in one call ───────────────────────────
  // Supabase's signInWithOtp sends a single email containing both
  // a magic link and a 6-digit OTP. No pre-choice needed.
  async function sendEmail(e: React.FormEvent) {
    e.preventDefault()
    if (!email) return
    setStep('submitting')
    setError('')

    const { error } = await supabase.auth.signInWithOtp({
      email,
      options: {
        shouldCreateUser: true,   // unified: creates account if new
        emailRedirectTo: redirectTo,
      },
    })

    if (error) {
      setError(error.message)
      setStep('idle')
    } else {
      setStep('email-sent')
    }
  }

  // ── OTP verification ───────────────────────────────────────────────────
  async function verifyOtp(e: React.FormEvent) {
    e.preventDefault()
    if (!otp || otp.length < 6) return
    setStep('submitting')
    setError('')

    const { error } = await supabase.auth.verifyOtp({
      email,
      token: otp,
      type: 'email',
    })

    if (error) {
      setError(error.message)
      setStep('email-sent')
    }
    // On success, Supabase fires onAuthStateChange → middleware redirects
  }

  return (
    <div className={cn('flex flex-col gap-6', className)}>
      {/* Brand */}
      <div className="flex flex-col items-center gap-2 text-center">
        <a href="/" className="flex items-center gap-2 font-semibold">
          <IconBolt className="size-6" />
          <span className="text-lg">Confire</span>
        </a>
        <h1 className="text-xl font-bold">
          {step === 'email-sent' ? 'Check your email' : 'Continue to Confire'}
        </h1>
        <p className="text-sm text-muted-foreground">
          {step === 'email-sent'
            ? `We sent a sign-in link to ${email}`
            : 'Sign in or create a free account'}
        </p>
      </div>

      {/* ── Email-sent state ───────────────────────────────────────────── */}
      {step === 'email-sent' && (
        <div className="flex flex-col gap-4">
          <p className="text-center text-sm text-muted-foreground">
            Click the sign-in link in your email
          </p>

          <div className="relative flex items-center gap-3 text-xs text-muted-foreground">
            <div className="flex-1 border-t" />
            <span>or enter one-time code</span>
            <div className="flex-1 border-t" />
          </div>

          <form onSubmit={verifyOtp} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="otp">One-time code</Label>
              <Input
                id="otp"
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                placeholder="123456"
                value={otp}
                onChange={e => setOtp(e.target.value.replace(/\D/g, ''))}
                autoFocus
                required
              />
            </div>
            <Button type="submit" disabled={step === 'submitting'}>
              {step === 'submitting' ? 'Verifying…' : 'Verify code'}
            </Button>
          </form>

          <button
            type="button"
            onClick={() => { setStep('idle'); setOtp('') }}
            className="text-center text-xs text-muted-foreground underline-offset-4 hover:underline"
          >
            Use a different email
          </button>
        </div>
      )}

      {/* ── Idle state: social + email ─────────────────────────────────── */}
      {(step === 'idle' || step === 'submitting') && (
        <div className="flex flex-col gap-4">
          {/* Social */}
          <div className="grid gap-2">
            <Button
              variant="outline"
              type="button"
              onClick={() => signInWithProvider('github')}
              disabled={step === 'submitting'}
            >
              <IconBrandGithub className="size-4" />
              Continue with GitHub
            </Button>
            <Button
              variant="outline"
              type="button"
              onClick={() => signInWithProvider('google')}
              disabled={step === 'submitting'}
            >
              <IconBrandGoogle className="size-4" />
              Continue with Google
            </Button>
          </div>

          {/* Divider */}
          <div className="relative flex items-center gap-3 text-xs text-muted-foreground">
            <div className="flex-1 border-t" />
            <span>or continue with email</span>
            <div className="flex-1 border-t" />
          </div>

          {/* Email */}
          <form onSubmit={sendEmail} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
                autoComplete="email"
              />
            </div>
            <Button type="submit" disabled={step === 'submitting'}>
              {step === 'submitting' ? 'Sending…' : 'Continue with email'}
            </Button>
          </form>
        </div>
      )}

      {/* Error */}
      {error && (
        <p className="text-center text-sm text-destructive">{error}</p>
      )}

      {/* Footer */}
      <p className="text-center text-xs text-muted-foreground">
        By continuing you agree to our{' '}
        <a href="/terms" className="underline underline-offset-4 hover:text-foreground">Terms</a>{' '}
        and{' '}
        <a href="/privacy" className="underline underline-offset-4 hover:text-foreground">Privacy Policy</a>.
      </p>
    </div>
  )
}
