import { useMemo, useState } from 'react'
import { Button, Field, Input, Text } from '@cloudflare/kumo'
import { createBrowserClient } from '@/lib/supabase'
import { cn } from '@/lib/utils'
import { AppShell } from '@/components/app/AppShell'

type Step = 'idle' | 'email-sent' | 'submitting'

interface Props {
  className?: string
  planIntent?: string
  next?: string
}

export function AuthForm({ className, planIntent = 'free', next = '' }: Props) {
  const supabase = useMemo(() => createBrowserClient(), [])
  const [step, setStep] = useState<Step>('idle')
  const [email, setEmail] = useState('')
  const [otp, setOtp] = useState('')
  const [error, setError] = useState('')

  const isFree = planIntent === 'free'

  // Detect CLI authorization flow from the `next` param
  const cliParams = useMemo(() => {
    if (!next) return null
    try {
      const nextUrl = new URL(next, 'http://x')
      if (!nextUrl.pathname.startsWith('/cli/login')) return null
      return {
        deviceName: nextUrl.searchParams.get('device_name') ?? '',
        cliVersion: nextUrl.searchParams.get('cli_version') ?? '',
      }
    } catch { return null }
  }, [next])

  const isCli = !!cliParams

  const callbackUrl = typeof window !== 'undefined'
    ? `${window.location.origin}/auth/callback?plan=${encodeURIComponent(planIntent)}${next ? `&next=${encodeURIComponent(next)}` : ''}`
    : `/auth/callback?plan=${encodeURIComponent(planIntent)}${next ? `&next=${encodeURIComponent(next)}` : ''}`

  const isLoading = step === 'submitting'

  async function signInWithProvider(provider: 'github' | 'google') {
    setError('')
    const { error: authError } = await supabase.auth.signInWithOAuth({
      provider,
      options: { redirectTo: callbackUrl },
    })
    if (authError) setError(authError.message)
  }

  async function sendEmail(e: React.FormEvent) {
    e.preventDefault()
    if (!email) return
    setStep('submitting')
    setError('')
    const { error: authError } = await supabase.auth.signInWithOtp({
      email,
      options: { shouldCreateUser: true, emailRedirectTo: callbackUrl },
    })
    if (authError) { setError(authError.message); setStep('idle') }
    else setStep('email-sent')
  }

  async function verifyOtp(e: React.FormEvent) {
    e.preventDefault()
    if (!otp || otp.length < 6) return
    setStep('submitting')
    setError('')
    const { error: authError } = await supabase.auth.verifyOtp({ email, token: otp, type: 'email' })
    if (authError) { setError(authError.message); setStep('email-sent') }
    else window.location.href = callbackUrl
  }

  const title = step === 'email-sent'
    ? 'Check your email'
    : isCli
      ? 'Sign in to authorize CLI'
      : isFree
        ? 'Start free with Confire'
        : 'Continue to checkout'

  const subtitle = step === 'email-sent'
    ? `We sent a code to ${email}`
    : isCli
      ? 'Sign in to connect your terminal to your Confire account.'
      : isFree
        ? 'Install Confire for Claude Code and optimize your first tool output.'
        : "Log in first, then we'll send you to secure checkout."

  return (
    <AppShell>
      <div className={cn('flex flex-col gap-6', className)}>
        <div className="flex flex-col items-center gap-2 text-center">
          <a href="/">
            <Text variant="heading3" as="span">Confire</Text>
          </a>
          <Text variant="heading2" as="h1">{title}</Text>
          <Text variant="secondary" size="sm">{subtitle}</Text>
        </div>

        {isCli && step !== 'email-sent' && (
          <div className="flex items-center gap-3 rounded-lg border border-kumo-hairline bg-kumo-tint px-4 py-3">
            <div className="flex size-8 shrink-0 items-center justify-center rounded-md bg-kumo-base">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M2 3.5A1.5 1.5 0 0 1 3.5 2h9A1.5 1.5 0 0 1 14 3.5v9a1.5 1.5 0 0 1-1.5 1.5h-9A1.5 1.5 0 0 1 2 12.5v-9Z" stroke="currentColor" strokeWidth="1.2"/>
                <path d="M5 6l2.5 2L5 10M8.5 10h2.5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <div className="min-w-0">
              <p className="text-sm font-medium text-kumo-default">Confire CLI wants access</p>
              <p className="truncate text-xs text-kumo-subtle">
                {cliParams?.deviceName
                  ? `From ${cliParams.deviceName}`
                  : 'From your terminal'}
                {cliParams?.cliVersion ? ` · ${cliParams.cliVersion}` : ''}
              </p>
            </div>
          </div>
        )}

        {step === 'email-sent' && (
          <form onSubmit={verifyOtp} className="flex flex-col gap-4">
            <Field label="One-time code">
              <Input
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
            </Field>
            <Button type="submit" variant="primary" className="w-full" loading={isLoading}>
              Verify code
            </Button>
            <button
              type="button"
              onClick={() => { setStep('idle'); setOtp('') }}
              className="text-center text-sm text-kumo-subtle underline-offset-4 hover:underline"
            >
              Use a different email
            </button>
          </form>
        )}

        {step !== 'email-sent' && (
          <>
            <form onSubmit={sendEmail} className="flex flex-col gap-4">
              <Field label="Email">
                <Input
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  required
                  autoComplete="email"
                />
              </Field>
              <Button type="submit" variant="primary" className="w-full" loading={isLoading}>
                Continue with email
              </Button>
            </form>

            <div className="relative text-center text-sm">
              <div className="absolute inset-0 top-1/2 border-t border-kumo-hairline" />
              <Text variant="secondary" size="sm" as="span" DANGEROUS_className="relative bg-kumo-base px-2">
                Or
              </Text>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <Button
                variant="outline"
                className="w-full"
                disabled={isLoading}
                onClick={() => signInWithProvider('github')}
                type="button"
              >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" className="size-4">
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" fill="currentColor" />
                </svg>
                GitHub
              </Button>
              <Button
                variant="outline"
                className="w-full"
                disabled={isLoading}
                onClick={() => signInWithProvider('google')}
                type="button"
              >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" className="size-4">
                  <path d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z" fill="currentColor" />
                </svg>
                Google
              </Button>
            </div>
          </>
        )}

        {error && (
          <Text variant="error" size="sm" as="p" DANGEROUS_className="text-center">
            {error}
          </Text>
        )}

        <Text variant="secondary" size="xs" as="p" DANGEROUS_className="text-balance text-center [&_a]:underline [&_a]:underline-offset-4 hover:[&_a]:text-kumo-link">
          By clicking continue, you agree to our{' '}
          <a href="/terms">Terms of Service</a> and{' '}
          <a href="/privacy">Privacy Policy</a>.
        </Text>
      </div>
    </AppShell>
  )
}
